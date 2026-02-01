package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"passport-desk-mvp/internal/database"
	"passport-desk-mvp/internal/security"
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
		dataDir:  dataDir,
		keystore: security.NewKeystore(dataDir),
		isLocked: true,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Ensure data directory exists
	if err := a.keystore.EnsureDataDir(); err != nil {
		println("Warning: failed to create data directory:", err.Error())
	}

	// Start inactivity checker
	go a.inactivityChecker()
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		a.db.Close()
	}
}

// inactivityChecker monitors for inactivity and locks the app
func (a *App) inactivityChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.mu.Lock()
			if !a.isLocked && a.currentOperator != nil {
				if time.Since(a.lastActivity) > inactivityTimeout {
					a.isLocked = true
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
	return !a.keystore.SaltExists()
}

// SetupInitialOperator creates the first operator and initializes the database
func (a *App) SetupInitialOperator(username, password, fullName string) error {
	if !a.IsFirstRun() {
		return errors.New("initial setup already completed")
	}

	// Generate salt
	salt, err := security.GenerateSalt()
	if err != nil {
		return err
	}

	// Save salt
	if err := a.keystore.SaveSalt(salt); err != nil {
		return err
	}

	// Derive encryption key
	key := security.DeriveKey(password, salt)
	a.encryptionKey = key

	// Initialize database
	db, err := database.New(a.keystore.GetDBPath(), security.KeyToHex(key))
	if err != nil {
		return err
	}
	a.db = db
	a.crypto = security.NewCrypto(key)

	// Hash password
	passwordHash, err := security.HashPassword(password)
	if err != nil {
		return err
	}

	// Create operator
	operator := &database.Operator{
		Username:     username,
		PasswordHash: passwordHash,
		FullName:     fullName,
	}
	if err := a.db.CreateOperator(operator); err != nil {
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
		return errors.New("database not initialized")
	}

	// Derive key from password
	key := security.DeriveKey(password, salt)

	// Try to open database with this key
	db, err := database.New(a.keystore.GetDBPath(), security.KeyToHex(key))
	if err != nil {
		return errors.New("неправильний пароль")
	}

	// Get operator
	operator, err := db.GetOperatorByUsername(username)
	if err != nil {
		db.Close()
		return errors.New("користувача не знайдено")
	}

	// Verify password
	if !security.VerifyPassword(operator.PasswordHash, password) {
		db.Close()
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
	a.encryptionKey = nil
}

// Unlock unlocks the app with password (after auto-lock)
func (a *App) Unlock(password string) error {
	a.mu.RLock()
	operator := a.currentOperator
	a.mu.RUnlock()

	if operator == nil {
		return errors.New("not logged in")
	}

	if !security.VerifyPassword(operator.PasswordHash, password) {
		return errors.New("неправильний пароль")
	}

	a.mu.Lock()
	a.isLocked = false
	a.lastActivity = time.Now()
	a.mu.Unlock()

	return nil
}

// IsLocked returns whether the app is locked
func (a *App) IsLocked() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isLocked
}

// IsAuthenticated returns whether a user is authenticated
func (a *App) IsAuthenticated() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentOperator != nil && !a.isLocked
}

// GetCurrentOperator returns the current operator info
func (a *App) GetCurrentOperator() *OperatorInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.currentOperator == nil {
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
