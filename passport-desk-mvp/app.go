package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"passport-desk-mvp/internal/database"
	"passport-desk-mvp/internal/logger"
	"passport-desk-mvp/internal/security"
	"passport-desk-mvp/internal/services"
)

const (
	// Auto-lock after 5 minutes of inactivity as per TZ
	inactivityTimeout = 5 * time.Minute
)

// App struct holds the application state
type App struct {
	ctx context.Context

	// Database
	db       *database.Database
	keystore *security.Keystore
	crypto   *security.Crypto

	// Services
	citizenService      *services.CitizenService
	registrationService *services.RegistrationService
	reportService       *services.ReportService
	backupService       *services.BackupService
	importExportService *services.ImportExportService

	// Session state
	mu              sync.RWMutex
	currentOperator *database.Operator
	isLocked        bool
	lastActivity    time.Time
	encryptionKey   []byte

	// Data directory
	dataDir string
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	dataDir := filepath.Join(configDir, "passport-desk-mvp")

	return &App{
		dataDir:       dataDir,
		keystore:      security.NewKeystore(dataDir),
		backupService: services.NewBackupService(dataDir),
		isLocked:      true,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Отримати директорію для даних
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	dataDir := filepath.Join(configDir, "passport-desk-mvp")

	// Ініціалізувати logger
	logConfig := logger.Config{
		Level:      getLogLevel(),
		LogDir:     filepath.Join(dataDir, "logs"),
		LogFile:    "passport-desk.log",
		MaxSize:    10, // 10 MB
		MaxAge:     30, // 30 днів
		MaxBackups: 10, // 10 backup файлів
		Compress:   true,
		Console:    isDevelopment(),
		JSON:       false, // Text формат для читабельності
	}

	if err := logger.Init(logConfig); err != nil {
		// Fallback на стандартний log якщо не вдалось ініціалізувати
		panic("Failed to initialize logger: " + err.Error())
	}

	logger.Info("Application started",
		slog.String("version", "1.0.0"),
		slog.String("data_dir", dataDir),
	)

	// Ensure data directory exists
	if err := a.keystore.EnsureDataDir(); err != nil {
		logger.Error("Failed to create data directory", slog.String("error", err.Error()))
	}

	// Run backup
	if err := a.backupService.RunBackup(); err != nil {
		logger.Error("Failed to run backup", slog.String("error", err.Error()))
	}

	// Start inactivity checker
	go a.inactivityChecker()
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	logger.Info("Application shutting down")
	if a.db != nil {
		a.db.Close()
	}
	logger.Close()
}

// getLogLevel визначає рівень логування з environment або config
func getLogLevel() string {
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		return level
	}
	if isDevelopment() {
		return "debug"
	}
	return "info"
}

// isDevelopment перевіряє чи це dev режим
func isDevelopment() bool {
	env := os.Getenv("ENV")
	return env == "development" || env == "dev" || env == ""
}

// inactivityChecker monitors for inactivity and locks the app
func (a *App) inactivityChecker() {
	logger.Info("Inactivity checker started")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.mu.Lock()
			if !a.isLocked && a.currentOperator != nil {
				if time.Since(a.lastActivity) > inactivityTimeout {
					a.isLocked = true
					runtime.EventsEmit(a.ctx, "session-locked")
				}
			}
			a.mu.Unlock()
		case <-a.ctx.Done():
			return
		}
	}
}

// UpdateActivity updates the last activity timestamp
func (a *App) UpdateActivity() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastActivity = time.Now()
}

// --- Authentication Methods ---

// IsFirstRun checks if this is the first run (no database exists)
func (a *App) IsFirstRun() bool {
	logger.Info("Checking if this is the first run")
	return !a.keystore.SaltExists()
}

// SetupInitialOperator creates the first operator and initializes the database
func (a *App) SetupInitialOperator(username, password, fullName string) error {
	logger.Info("Setting up initial operator")
	if !a.IsFirstRun() {
		logger.Error("Initial setup already completed")
		return errors.New("initial setup already completed")
	}

	// Generate salt
	salt, err := security.GenerateSalt()
	if err != nil {
		logger.Error("Failed to generate salt", slog.String("error", err.Error()))
		return err
	}

	// Save salt
	if err := a.keystore.SaveSalt(salt); err != nil {
		logger.Error("Failed to save salt", slog.String("error", err.Error()))
		return err
	}

	// Derive encryption key
	key := security.DeriveKey(password, salt)
	a.encryptionKey = key

	// Initialize database
	db, err := database.New(a.keystore.GetDBPath(), security.KeyToHex(key))
	if err != nil {
		logger.Error("Failed to initialize database", slog.String("error", err.Error()))
		return err
	}
	a.db = db
	a.crypto = security.NewCrypto(key)
	a.citizenService = services.NewCitizenService(db, a.crypto)
	a.registrationService = services.NewRegistrationService(db, a.crypto)
	a.reportService = services.NewReportService(db, a.citizenService, a.registrationService)

	// Hash password
	passwordHash, err := security.HashPassword(password)
	if err != nil {
		logger.Error("Failed to hash password", slog.String("error", err.Error()))
		return err
	}

	// Create operator
	operator := &database.Operator{
		Username:     username,
		PasswordHash: passwordHash,
		FullName:     fullName,
	}
	if err := a.db.CreateOperator(operator); err != nil {
		logger.Error("Failed to create operator", slog.String("error", err.Error()))
		return err
	}

	// Log to audit
	a.db.LogAudit(&database.AuditLog{
		OperatorID:  operator.ID,
		ActionType:  "CREATE",
		TableName:   "operators",
		RecordID:    operator.ID,
		Description: "Initial operator created",
	})

	// Set session
	a.mu.Lock()
	a.currentOperator = operator
	a.isLocked = false
	a.lastActivity = time.Now()
	a.mu.Unlock()

	return nil
}

// Login authenticates an operator
func (a *App) Login(username, password string) error {
	// Load salt
	salt, err := a.keystore.LoadSalt()
	if err != nil {
		logger.Error("Failed to load salt", slog.String("error", err.Error()))
		return errors.New("database not initialized")
	}

	// Derive key from password
	key := security.DeriveKey(password, salt)

	// Try to open database with this key
	db, err := database.New(a.keystore.GetDBPath(), security.KeyToHex(key))
	if err != nil {
		logger.Error("Failed to open database", slog.String("error", err.Error()))
		return errors.New("неправильний пароль")
	}

	// Get operator
	operator, err := db.GetOperatorByUsername(username)
	if err != nil {
		db.Close()
		logger.Error("Failed to get operator", slog.String("error", err.Error()))
		return errors.New("користувача не знайдено")
	}

	// Verify password
	if !security.VerifyPassword(operator.PasswordHash, password) {
		db.Close()
		logger.Error("Failed to verify password", slog.String("error", err.Error()))
		return errors.New("неправильний пароль")
	}

	// Success - update state
	a.mu.Lock()
	if a.db != nil {
		a.db.Close()
	}
	a.db = db
	a.encryptionKey = key
	a.crypto = security.NewCrypto(key)
	a.citizenService = services.NewCitizenService(db, a.crypto)
	a.registrationService = services.NewRegistrationService(db, a.crypto)
	a.reportService = services.NewReportService(a.db, a.citizenService, a.registrationService)
	a.backupService = services.NewBackupService(a.dataDir)
	a.importExportService = services.NewImportExportService(a.db, a.citizenService)
	a.currentOperator = operator
	a.isLocked = false
	a.lastActivity = time.Now()
	a.mu.Unlock()

	// Log to audit
	a.db.LogAudit(&database.AuditLog{
		OperatorID:  operator.ID,
		ActionType:  "READ",
		TableName:   "operators",
		RecordID:    operator.ID,
		Description: "Operator logged in",
	})

	return nil
}

// Logout logs out the current operator
func (a *App) Logout() {
	logger.Info("Logging out")
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentOperator != nil && a.db != nil {
		a.db.LogAudit(&database.AuditLog{
			OperatorID:  a.currentOperator.ID,
			ActionType:  "READ",
			TableName:   "operators",
			RecordID:    a.currentOperator.ID,
			Description: "Operator logged out",
		})
	}

	a.currentOperator = nil
	a.isLocked = true
	if a.db != nil {
		a.db.Close()
		a.db = nil
	}
	a.crypto = nil
	a.citizenService = nil
	a.registrationService = nil
	a.reportService = nil
	a.encryptionKey = nil
}

// Unlock unlocks the app with password (after auto-lock)
func (a *App) Unlock(password string) error {
	logger.Info("Unlocking")
	a.mu.RLock()
	operator := a.currentOperator
	a.mu.RUnlock()

	if operator == nil {
		logger.Error("Not logged in")
		return errors.New("not logged in")
	}

	if !security.VerifyPassword(operator.PasswordHash, password) {
		err := errors.New("неправильний пароль")
		logger.Error("Failed to verify password", slog.String("error", err.Error()))
		return err
	}

	a.mu.Lock()
	a.isLocked = false
	a.lastActivity = time.Now()
	a.mu.Unlock()

	return nil
}

// IsLocked returns whether the app is locked
func (a *App) IsLocked() bool {
	logger.Info("Checking if locked")
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isLocked
}

// IsAuthenticated returns whether a user is authenticated
func (a *App) IsAuthenticated() bool {
	logger.Info("Checking if authenticated")
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentOperator != nil && !a.isLocked
}

// GetCurrentOperator returns the current operator info
func (a *App) GetCurrentOperator() *OperatorInfo {
	logger.Info("Getting current operator")
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.currentOperator == nil {
		logger.Error("Not logged in")
		return nil
	}

	return &OperatorInfo{
		ID:       a.currentOperator.ID,
		Username: a.currentOperator.Username,
		FullName: a.currentOperator.FullName,
	}
}

// OperatorInfo is a safe struct for frontend (no password hash)
type OperatorInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

// --- Citizen Methods ---

// CreateCitizen creates a new citizen
func (a *App) CreateCitizen(input services.CitizenInput) (*services.CitizenOutput, error) {
	if !a.IsAuthenticated() {
		logger.Error("CreateCitizen: Not authenticated")
		return nil, errors.New("unauthorized")
	}
	a.UpdateActivity()

	citizen, err := a.citizenService.Create(&input)
	if err == nil {
		a.db.LogAudit(&database.AuditLog{
			OperatorID:  a.currentOperator.ID,
			ActionType:  "CREATE",
			TableName:   "citizens",
			RecordID:    citizen.ID,
			Description: fmt.Sprintf("Created citizen: %s", citizen.FullName),
		})
	}
	return citizen, err
}

// GetCitizen retrieves a citizen by ID
func (a *App) GetCitizen(id int64) (*services.CitizenOutput, error) {
	if !a.IsAuthenticated() {
		logger.Error("GetCitizen: Not authenticated")
		return nil, errors.New("unauthorized")
	}
	a.UpdateActivity()

	citizen, err := a.citizenService.GetByID(id)
	if err == nil {
		a.db.LogAudit(&database.AuditLog{
			OperatorID:  a.currentOperator.ID,
			ActionType:  "READ",
			TableName:   "citizens",
			RecordID:    id,
			Description: fmt.Sprintf("Viewed citizen: %s", citizen.FullName),
		})
	}
	return citizen, err
}

// UpdateCitizen updates a citizen
func (a *App) UpdateCitizen(id int64, input services.CitizenInput) error {
	if !a.IsAuthenticated() {
		logger.Error("UpdateCitizen: Not authenticated")
		return errors.New("unauthorized")
	}
	a.UpdateActivity()

	err := a.citizenService.Update(id, &input)
	if err == nil {
		a.db.LogAudit(&database.AuditLog{
			OperatorID:  a.currentOperator.ID,
			ActionType:  "UPDATE",
			TableName:   "citizens",
			RecordID:    id,
			Description: fmt.Sprintf("Updated citizen: %s %s %s", input.LastName, input.FirstName, input.MiddleName),
		})
	}
	return err
}

// DeleteCitizen soft deletes a citizen
func (a *App) DeleteCitizen(id int64) error {
	if !a.IsAuthenticated() {
		logger.Error("DeleteCitizen: Not authenticated")
		return errors.New("unauthorized")
	}
	a.UpdateActivity()

	err := a.citizenService.Delete(id)
	if err == nil {
		a.db.LogAudit(&database.AuditLog{
			OperatorID:  a.currentOperator.ID,
			ActionType:  "DELETE",
			TableName:   "citizens",
			RecordID:    id,
			Description: "Soft deleted citizen",
		})
	}
	return err
}

// RestoreCitizen restores a soft-deleted citizen
func (a *App) RestoreCitizen(id int64) error {
	if !a.IsAuthenticated() {
		logger.Error("RestoreCitizen: Not authenticated")
		return errors.New("unauthorized")
	}
	a.UpdateActivity()

	err := a.citizenService.Restore(id)
	if err == nil {
		a.db.LogAudit(&database.AuditLog{
			OperatorID:  a.currentOperator.ID,
			ActionType:  "UPDATE",
			TableName:   "citizens",
			RecordID:    id,
			Description: "Restored soft-deleted citizen",
		})
	}
	return err
}

// AddFamilyMember adds a connection between citizens
func (a *App) AddFamilyMember(citizenID, memberID int64, relationType string) error {
	if !a.IsAuthenticated() {
		logger.Error("AddFamilyMember: Not authenticated")
		return errors.New("unauthorized")
	}
	a.UpdateActivity()
	return a.citizenService.AddFamilyMember(citizenID, memberID, relationType)
}

// RemoveFamilyMember removes a connection
func (a *App) RemoveFamilyMember(citizenID, memberID int64) error {
	if !a.IsAuthenticated() {
		logger.Error("RemoveFamilyMember: Not authenticated")
		return errors.New("unauthorized")
	}
	a.UpdateActivity()
	return a.citizenService.RemoveFamilyMember(citizenID, memberID)
}

// GetFamilyMembers returns family members
func (a *App) GetFamilyMembers(citizenID int64) ([]services.FamilyMemberOutput, error) {
	if !a.IsAuthenticated() {
		logger.Error("GetFamilyMembers: Not authenticated")
		return nil, errors.New("unauthorized")
	}
	a.UpdateActivity()
	return a.citizenService.GetFamilyMembers(citizenID)
}

// SearchCitizens searches citizens by field
func (a *App) SearchCitizens(query string, field string) ([]services.CitizenOutput, error) {
	if !a.IsAuthenticated() {
		logger.Error("SearchCitizens: Not authenticated")
		return nil, errors.New("unauthorized")
	}
	a.UpdateActivity()
	return a.citizenService.Search(query, field)
}

// ListCitizens returns paginated citizens
func (a *App) ListCitizens(page, limit int, includeDeleted bool) (*services.CitizenListResult, error) {
	if !a.IsAuthenticated() {
		logger.Error("ListCitizens: Not authenticated")
		return nil, errors.New("unauthorized")
	}
	a.UpdateActivity()

	return a.citizenService.List(page, limit, includeDeleted)
}

// --- Registration Methods ---

// CreateRegistration creates a new registration
func (a *App) CreateRegistration(input database.RegistrationInput) (*database.RegistrationOutput, error) {
	if !a.IsAuthenticated() {
		logger.Error("CreateRegistration: Not authenticated")
		return nil, errors.New("unauthorized")
	}
	a.UpdateActivity()

	reg, err := a.registrationService.Create(&input)
	if err == nil {
		a.db.LogAudit(&database.AuditLog{
			OperatorID:  a.currentOperator.ID,
			ActionType:  "CREATE",
			TableName:   "registrations",
			RecordID:    reg.ID,
			Description: fmt.Sprintf("Created registration for citizen %d", reg.CitizenID),
		})
	}
	return reg, err
}

// GetRegistrationHistory returns registration history for a citizen
func (a *App) GetRegistrationHistory(citizenID int64) ([]database.RegistrationOutput, error) {
	if !a.IsAuthenticated() {
		logger.Error("GetRegistrationHistory: Not authenticated")
		return nil, errors.New("unauthorized")
	}
	a.UpdateActivity()
	return a.registrationService.GetByCitizenID(citizenID)
}

// DeregisterCitizen deactivates a registration
func (a *App) DeregisterCitizen(id int64, date string) error {
	if !a.IsAuthenticated() {
		logger.Error("DeregisterCitizen: Not authenticated")
		return errors.New("unauthorized")
	}
	a.UpdateActivity()

	err := a.registrationService.Deregister(id, date)
	if err == nil {
		a.db.LogAudit(&database.AuditLog{
			OperatorID:  a.currentOperator.ID,
			ActionType:  "UPDATE", // Logically an update
			TableName:   "registrations",
			RecordID:    id,
			Description: "Deregistered citizen",
		})
	}
	return err
}

// --- Report Methods ---

// GenerateCitizenCertificate generates a registration certificate for a citizen and saves it
func (a *App) GenerateCitizenCertificate(citizenID int64, opts services.FamilyCertificateOptions) (string, error) {
	if !a.IsAuthenticated() {
		logger.Error("GenerateCitizenCertificate: Not authenticated")
		return "", errors.New("unauthorized")
	}
	a.UpdateActivity()

	// Ask user where to save
	filename := fmt.Sprintf("certificate_%d_%s.pdf", citizenID, time.Now().Format("20060102"))
	filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Зберегти довідку",
		DefaultFilename: filename,
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF Files (*.pdf)", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return "", err
	}
	if filepath == "" {
		return "cancelled", nil
	}

	// Create audit log
	a.db.LogAudit(&database.AuditLog{
		OperatorID:  a.currentOperator.ID,
		ActionType:  "READ",
		TableName:   "registrations",
		RecordID:    citizenID,
		Description: fmt.Sprintf("Generated certificate for citizen %d", citizenID),
	})

	base64Data, err := a.reportService.GenerateRegistrationCertificate(citizenID, opts)
	if err != nil {
		logger.Error("Failed to generate certificate", slog.String("error", err.Error()))
		return "", err
	}

	// Decode PDF
	pdfBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		logger.Error("Failed to decode PDF", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to decode PDF: %w", err)
	}

	// Save to selected file
	if err := os.WriteFile(filepath, pdfBytes, 0644); err != nil {
		logger.Error("Failed to save file", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filepath, nil
}

// ExportCitizens exports registered citizens to Excel and saves it
func (a *App) ExportCitizens(from, to string) (string, error) {
	if !a.IsAuthenticated() {
		logger.Error("ExportCitizens: Not authenticated")
		return "", errors.New("unauthorized")
	}
	a.UpdateActivity()

	// Ask user where to save
	filename := fmt.Sprintf("registrations_%s.xlsx", time.Now().Format("20060102"))
	filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Зберегти звіт",
		DefaultFilename: filename,
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil {
		logger.Error("Failed to open save dialog", slog.String("error", err.Error()))
		return "", err
	}
	if filepath == "" {
		logger.Info("User cancelled export")
		return "cancelled", nil // User cancelled
	}

	a.db.LogAudit(&database.AuditLog{
		OperatorID:  a.currentOperator.ID,
		ActionType:  "READ",
		TableName:   "registrations",
		Description: fmt.Sprintf("Exported citizens list from %s to %s", from, to),
	})

	base64Data, err := a.reportService.ExportRegisteredCitizens(from, to)
	if err != nil {
		logger.Error("Failed to export citizens", slog.String("error", err.Error()))
		return "", err
	}

	// Decode Excel
	xlsxBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		logger.Error("Failed to decode Excel", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to decode Excel: %w", err)
	}

	// Save file
	if err := os.WriteFile(filepath, xlsxBytes, 0644); err != nil {
		logger.Error("Failed to save file", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filepath, nil
}

// ExportCitizensToExcel exports all citizens to Excel file
func (a *App) ExportCitizensToExcel() (string, error) {
	if !a.IsAuthenticated() {
		logger.Error("ExportCitizensToExcel: Not authenticated")
		return "", errors.New("unauthorized")
	}
	a.UpdateActivity()

	filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Експорт громадян",
		DefaultFilename: fmt.Sprintf("citizens_export_%s.xlsx", time.Now().Format("20060102")),
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil || filepath == "" {
		logger.Error("Failed to open save dialog", slog.String("error", err.Error()))
		return "cancelled", err
	}

	data, err := a.importExportService.ExportToExcel()
	if err != nil {
		logger.Error("Failed to export citizens", slog.String("error", err.Error()))
		return "", err
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		logger.Error("Failed to save file", slog.String("error", err.Error()))
		return "", err
	}

	a.db.LogAudit(&database.AuditLog{
		OperatorID:  a.currentOperator.ID,
		ActionType:  "READ",
		TableName:   "citizens",
		Description: "Exported all citizens to Excel",
	})

	return filepath, nil
}

// ImportCitizens imports citizens from selected Excel file
func (a *App) ImportCitizens() (int, error) {
	if !a.IsAuthenticated() {
		return 0, errors.New("unauthorized")
	}
	a.UpdateActivity()

	filepath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Виберіть файл для імпорту",
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil || filepath == "" {
		logger.Error("Failed to open save dialog", slog.String("error", err.Error()))
		return 0, err
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		logger.Error("Failed to read file", slog.String("error", err.Error()))
		return 0, err
	}

	count, err := a.importExportService.ImportFromExcel(data)
	if err == nil {
		a.db.LogAudit(&database.AuditLog{
			OperatorID:  a.currentOperator.ID,
			ActionType:  "CREATE",
			TableName:   "citizens",
			Description: fmt.Sprintf("Imported %d citizens from Excel", count),
		})
	}

	return count, err
}

// ExportCustomCitizensToExcel exports selected citizens and columns to Excel
func (a *App) ExportCustomCitizensToExcel(ids []int64, columns []string) (string, error) {
	if !a.IsAuthenticated() {
		return "", errors.New("unauthorized")
	}
	a.UpdateActivity()

	if len(ids) == 0 {
		return "", errors.New("no citizens selected")
	}

	filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Експорт обраних громадян",
		DefaultFilename: fmt.Sprintf("citizens_custom_export_%s.xlsx", time.Now().Format("20060102")),
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel Files (*.xlsx)", Pattern: "*.xlsx"},
		},
	})
	if err != nil || filepath == "" {
		logger.Error("Failed to open save dialog", slog.String("error", err.Error()))
		return "cancelled", err
	}

	data, err := a.importExportService.ExportCustomToExcel(ids, columns)
	if err != nil {
		logger.Error("Failed to export citizens", slog.String("error", err.Error()))
		return "", err
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		logger.Error("Failed to save file", slog.String("error", err.Error()))
		return "", err
	}

	a.db.LogAudit(&database.AuditLog{
		OperatorID:  a.currentOperator.ID,
		ActionType:  "READ",
		TableName:   "citizens",
		Description: fmt.Sprintf("Exported %d selected citizens to Excel with %d columns", len(ids), len(columns)),
	})

	return filepath, nil
}

// GetDashboardStats returns dashboard statistics
func (a *App) GetDashboardStats() (*services.StatsOutput, error) {
	if !a.IsAuthenticated() {
		return nil, errors.New("unauthorized")
	}
	a.UpdateActivity()
	return a.reportService.GetStats()
}

// GetAuditLogs returns recent audit logs
func (a *App) GetAuditLogs(limit int) ([]database.AuditLogOutput, error) {
	if !a.IsAuthenticated() {
		logger.Error("GetAuditLogs: Not authenticated")
		return nil, errors.New("unauthorized")
	}
	a.UpdateActivity()
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	return a.db.GetAuditLogs(limit)
}

// GenerateFamilyStatusCertificate generates a custom family certificate and saves it
func (a *App) GenerateFamilyStatusCertificate(opts services.FamilyCertificateOptions) (string, error) {
	if !a.IsAuthenticated() {
		logger.Error("GenerateFamilyStatusCertificate: Not authenticated")
		return "", errors.New("unauthorized")
	}
	a.UpdateActivity()

	if len(opts.CitizenIDs) == 0 {
		return "", errors.New("no citizens selected")
	}

	// Ask user where to save
	filename := fmt.Sprintf("family_certificate_%s.pdf", time.Now().Format("20060102_150405"))
	filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Зберегти довідку про склад сім'ї",
		DefaultFilename: filename,
		Filters: []runtime.FileFilter{
			{DisplayName: "PDF Files (*.pdf)", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		logger.Error("Failed to open save dialog", slog.String("error", err.Error()))
		return "", err
	}
	if filepath == "" {
		logger.Info("User cancelled export")
		return "cancelled", nil
	}

	// Create audit log
	a.db.LogAudit(&database.AuditLog{
		OperatorID:  a.currentOperator.ID,
		ActionType:  "READ",
		TableName:   "citizens",
		Description: fmt.Sprintf("Generated family certificate for %d citizens", len(opts.CitizenIDs)),
	})

	base64Data, err := a.reportService.GenerateFamilyStatusCertificate(opts)
	if err != nil {
		logger.Error("Failed to generate certificate", slog.String("error", err.Error()))
		return "", err
	}

	// Decode PDF
	pdfBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		logger.Error("Failed to decode PDF", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to decode PDF: %w", err)
	}

	// Save to selected file
	if err := os.WriteFile(filepath, pdfBytes, 0644); err != nil {
		logger.Error("Failed to save file", slog.String("error", err.Error()))
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	return filepath, nil
}
func (a *App) ListRegistrations(search string, isActive *bool, page, limit int) (*services.RegistrationListResult, error) {
	if !a.IsAuthenticated() {
		logger.Error("ListRegistrations: Not authenticated")
		return nil, errors.New("unauthorized")
	}
	a.UpdateActivity()
	return a.registrationService.ListAll(search, isActive, page, limit)
}
