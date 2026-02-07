package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"passport-desk-mvp/internal/logger"
	"passport-desk-mvp/internal/models"
	"passport-desk-mvp/internal/repository"
	"passport-desk-mvp/internal/security"
)

// CitizenService handles citizen-related operations
type CitizenService struct {
	crypto    *security.Crypto
	repo      *repository.CitizenRepository
	validator *CitizenValidator
	auditLog  *AuditLogService
}

// NewCitizenService creates a new citizen service
func NewCitizenService(crypto *security.Crypto, repo *repository.CitizenRepository, auditLog *AuditLogService) *CitizenService {
	validator := NewCitizenValidator()
	return &CitizenService{crypto: crypto, repo: repo, validator: validator, auditLog: auditLog}
}

func (s *CitizenService) Create(ctx context.Context, input *models.CitizenInput) (*models.CitizenOutput, error) {
	log := logger.FromContext(ctx).With(
		slog.String("service", "citizen"),
		slog.String("method", "create"),
	)

	// SLOG: Tech log start operation
	log.Info("Creating citizen",
		slog.String("name", input.FirstName+" "+input.LastName),
	)

	// Validate input
	if err := s.validator.Validate(input); err != nil {
		log.Error("Validation failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	encryptedInput, err := s.encryptSensitiveFields(input)
	if err != nil {
		log.Error("Encryption failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	encryptedInput.BirthDate = normalizeDate(encryptedInput.BirthDate)

	// Create in DB
	id, err := s.repo.Create(ctx, encryptedInput)
	if err != nil {
		log.Error("Create failed", slog.String("error", err.Error()), slog.String("table", "citizens"))
		return nil, fmt.Errorf("create failed: %w", err)
	}

	log.Info("Citizen created successfully",
		slog.Int64("id", id),
		slog.Duration("duration", time.Since(time.Now())),
	)

	// AUDIT LOG: Business event
	if err := s.auditLog.LogAudit(ctx, &models.AuditLog{
		OperatorID:  getOperatorIDFromContext(ctx),
		ActionType:  "CREATE",
		TableName:   "citizens",
		RecordID:    id,
		Description: fmt.Sprintf("Citizen created: %s %s %s", input.LastName, input.FirstName, input.MiddleName),
	}); err != nil {
		log.Warn("Audit log failed", slog.String("error", err.Error()))
	}
	return s.repo.GetByID(ctx, id)
}

func (s *CitizenService) GetByID(ctx context.Context, id int64) (*models.CitizenOutput, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CitizenService) Update(ctx context.Context, id int64, input *models.CitizenInput) (*models.CitizenOutput, error) {
	log := logger.FromContext(ctx).With(
		slog.String("service", "citizen"),
		slog.String("method", "update"),
	)

	// Validate input
	if err := s.validator.Validate(input); err != nil {
		log.Error("Validation failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	encryptedInput, err := s.encryptSensitiveFields(input)
	if err != nil {
		log.Error("Encryption failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	encryptedInput.BirthDate = normalizeDate(encryptedInput.BirthDate)

	// Update in DB
	if err := s.repo.Update(ctx, id, encryptedInput); err != nil {
		log.Error("Update failed", slog.String("error", err.Error()), slog.String("table", "citizens"))
		return nil, fmt.Errorf("update failed: %w", err)
	}

	log.Info("Citizen updated successfully",
		slog.Int64("id", id),
		slog.Duration("duration", time.Since(time.Now())),
	)

	// AUDIT LOG: Business event
	if err := s.auditLog.LogAudit(ctx, &models.AuditLog{
		OperatorID:  getOperatorIDFromContext(ctx),
		ActionType:  "UPDATE",
		TableName:   "citizens",
		RecordID:    id,
		Description: fmt.Sprintf("Citizen updated: %s %s %s", input.LastName, input.FirstName, input.MiddleName),
	}); err != nil {
		log.Warn("Audit log failed", slog.String("error", err.Error()))
	}
	return s.repo.GetByID(ctx, id)
}

func (s *CitizenService) Delete(ctx context.Context, id int64) error {
	log := logger.FromContext(ctx).With(
		slog.String("service", "citizen"),
		slog.String("method", "Delete"),
		slog.Int64("id", id),
	)

	// Отримуємо громадянина для опису в аудиті
	citizen, err := s.GetByID(ctx, id)
	if err != nil {
		// SLOG: Технічна помилка
		log.Error("Failed to get citizen for deletion", slog.String("error", err.Error()))
		return err
	}

	// SLOG: Технічний лог операції
	log.Info("Deleting citizen", slog.String("name", citizen.FullName))

	// Soft delete
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		// SLOG: Технічна помилка
		log.Error("Database delete failed", slog.String("error", err.Error()))
		return err
	}

	// SLOG: Успішне видалення
	log.Info("Citizen deleted successfully")

	// AUDIT LOG: Критична бізнес-подія (видалення даних)
	if err := s.auditLog.LogAudit(ctx, &models.AuditLog{
		OperatorID:  getOperatorIDFromContext(ctx),
		ActionType:  "DELETE",
		TableName:   "citizens",
		RecordID:    id,
		Description: fmt.Sprintf("Soft deleted citizen: %s", citizen.FullName),
	}); err != nil {
		log.Error("Critical: Audit log failed for DELETE operation",
			slog.String("error", err.Error()))
		// У випадку видалення - це критично! Можливо повернути помилку
	}

	return nil
}

func (s *CitizenService) Restore(ctx context.Context, id int64) error {
	log := logger.FromContext(ctx).With(
		slog.String("service", "citizen"),
		slog.String("method", "Restore"),
		slog.Int64("id", id),
	)

	// Отримуємо громадянина для опису в аудиті
	citizen, err := s.GetByID(ctx, id)
	if err != nil {
		// SLOG: Технічна помилка
		log.Error("Failed to get citizen for deletion", slog.String("error", err.Error()))
		return err
	}

	// SLOG: Технічний лог операції
	log.Info("Restoring citizen", slog.String("name", citizen.FullName))

	// Soft delete
	if err := s.repo.Restore(ctx, id); err != nil {
		// SLOG: Технічна помилка
		log.Error("Database delete failed", slog.String("error", err.Error()))
		return err
	}

	// SLOG: Успішне видалення
	log.Info("Citizen restored successfully")

	// AUDIT LOG: Критична бізнес-подія (видалення даних)
	if err := s.auditLog.LogAudit(ctx, &models.AuditLog{
		OperatorID:  getOperatorIDFromContext(ctx),
		ActionType:  "RESTORE",
		TableName:   "citizens",
		RecordID:    id,
		Description: fmt.Sprintf("Soft restored citizen: %s", citizen.FullName),
	}); err != nil {
		log.Error("Critical: Audit log failed for RESTORE operation",
			slog.String("error", err.Error()))
		// У випадку видалення - це критично! Можливо повернути помилку
	}

	return nil
}

func (s *CitizenService) List(ctx context.Context, offset, limit int, includeDeleted bool) (*models.CitizenListResult, error) {
	citizenList, integer, err := s.repo.List(ctx, offset, limit, includeDeleted)
	if err != nil {
		return nil, err
	}
	citizerRezulr := models.CitizenListResult{
		Total: integer,
		Items: citizenList,
	}
	return &citizerRezulr, nil
}

func (s *CitizenService) Search(ctx context.Context, query string, field string) ([]*models.CitizenOutput, error) {
	log := logger.FromContext(ctx).With(
		slog.String("service", "citizen"),
		slog.String("method", "Search"),
	)
	switch field {
	case "name":
		log.Info("Searching by name", slog.String("query", query))
		return s.repo.SearchByName(ctx, query, 10)
	case "birth_date":
		log.Info("Searching by birth date", slog.String("query", query))
		return s.repo.SearchByBirthDate(ctx, query, 10)
	default:
		return nil, errors.New("invalid field")
	}
}

func (s *CitizenService) SearchByName(ctx context.Context, searchTerm string, limit int) ([]*models.CitizenOutput, error) {
	return s.repo.SearchByName(ctx, searchTerm, limit)
}

func (s *CitizenService) SearchByBirthDate(ctx context.Context, birthDate string, limit int) ([]*models.CitizenOutput, error) {
	return s.repo.SearchByBirthDate(ctx, birthDate, limit)
}

func (s *CitizenService) Exists(ctx context.Context, id int64) (bool, error) {
	return s.repo.Exists(ctx, id)
}

func (s *CitizenService) GetByPassport(ctx context.Context, passportSeries, passportNumber string) (*models.CitizenOutput, error) {
	return s.repo.GetByPassport(ctx, passportSeries, passportNumber)
}

func (s *CitizenService) GetAll(ctx context.Context) ([]*models.CitizenOutput, error) {
	return s.repo.GetAll(ctx)
}

func (s *CitizenService) GetFamilyMembers(ctx context.Context, citizenID int64) ([]*models.FamilyMemberOutput, error) {
	familyMembers, err := s.repo.GetFamilyMembers(ctx, citizenID)
	if err != nil {
		return nil, err
	}
	decryptedFamilyMembers := make([]*models.FamilyMemberOutput, len(familyMembers))
	for i, familyMember := range familyMembers {
		familyMember.BirthDate = normalizeDate(familyMember.BirthDate)
		familyMember.CitizenOutput, err = s.decryptSensitiveFields(familyMember.CitizenOutput)
		if err != nil {
			return nil, err
		}
		familyMember.CitizenOutput.BirthDate = normalizeDate(familyMember.CitizenOutput.BirthDate)
		decryptedFamilyMembers[i] = familyMember
	}
	return decryptedFamilyMembers, nil
}

func (s *CitizenService) RemoveFamilyMember(ctx context.Context, citizenID, memberID int64) error {
	return s.repo.RemoveFamilyMember(ctx, citizenID, memberID)
}

func (s *CitizenService) AddFamilyMember(ctx context.Context, citizenID, memberID int64, relationType string) error {
	return s.repo.AddFamilyMember(ctx, citizenID, memberID, relationType)
}

// Scanner interface for both Row and Rows
type scanner interface {
	Scan(dest ...interface{}) error
}

func (s *CitizenService) scanCitizen(row *sql.Row) (*models.CitizenOutput, error) {
	return s.scanCitizenFromScanner(row)
}

func (s *CitizenService) scanCitizenFromRows(rows *sql.Rows) (*models.CitizenOutput, error) {
	return s.scanCitizenFromScanner(rows)
}

func (s *CitizenService) scanCitizenFromScanner(sc scanner) (*models.CitizenOutput, error) {
	var c models.CitizenOutput
	var activeAddress sql.NullString

	err := sc.Scan(
		&c.ID, &c.LastName, &c.FirstName, &c.MiddleName, &c.BirthDate,
		&c.PassportSeries, &c.PassportNumber, &c.PassportType, &c.TaxNumber,
		&c.Gender, &c.BirthPlace, &c.Phone, &c.Email, &c.Notes,
		&c.Deleted, &c.CreatedAt, &c.UpdatedAt, &activeAddress,
	)
	if err != nil {
		return nil, err
	}

	// Decrypt sensitive fields
	c, err = s.decryptSensitiveFields(c)
	if err != nil {
		return nil, err
	}
	c.BirthDate = normalizeDate(c.BirthDate)

	if activeAddress.Valid {
		c.ActiveAddress = activeAddress.String
	}

	// Computed fields
	c = *s.computeFields(&c)

	return &c, nil
}

// maskPassport masks passport number for display
func maskPassport(series, number string) string {
	if series == "" && number == "" {
		return ""
	}
	if len(number) <= 2 {
		return series + " " + number
	}
	return series + " " + number[:2] + "****"
}

// maskTaxNumber masks IPN for display
func maskTaxNumber(taxNumber string) string {
	if len(taxNumber) <= 4 {
		return taxNumber
	}
	return taxNumber[:4] + "******"
}

// genderToUkrainian converts gender code to Ukrainian
func genderToUkrainian(gender string) string {
	switch gender {
	case "M":
		return "Чоловік"
	case "F":
		return "Жінка"
	default:
		return ""
	}
}

func (s *CitizenService) encryptSensitiveFields(input *models.CitizenInput) (*models.CitizenInput, error) {
	// Encrypt sensitive fields
	encPassportSeries, err := s.crypto.Encrypt(input.PassportSeries)
	if err != nil {
		return nil, err
	}
	encPassportNumber, err := s.crypto.Encrypt(input.PassportNumber)
	if err != nil {
		return nil, err
	}
	encTaxNumber, err := s.crypto.Encrypt(input.TaxNumber)
	if err != nil {
		return nil, err
	}
	encPhone, err := s.crypto.Encrypt(input.Phone)
	if err != nil {
		return nil, err
	}

	input.PassportSeries = encPassportSeries
	input.PassportNumber = encPassportNumber
	input.TaxNumber = encTaxNumber
	input.Phone = encPhone

	return input, nil
}

func (s *CitizenService) decryptSensitiveFields(input models.CitizenOutput) (models.CitizenOutput, error) {
	// Decrypt sensitive fields
	input.PassportSeries, _ = s.crypto.Decrypt(input.PassportSeries)
	input.PassportNumber, _ = s.crypto.Decrypt(input.PassportNumber)
	input.TaxNumber, _ = s.crypto.Decrypt(input.TaxNumber)
	input.Phone, _ = s.crypto.Decrypt(input.Phone)

	return input, nil
}

func (s *CitizenService) computeFields(input *models.CitizenOutput) *models.CitizenOutput {
	// Compute fields
	var createdAt, updatedAt time.Time
	createdAt, _ = time.Parse("2006-01-02 15:04:05", input.CreatedAt)
	updatedAt, _ = time.Parse("2006-01-02 15:04:05", input.UpdatedAt)
	input.FullName = strings.TrimSpace(input.LastName + " " + input.FirstName + " " + input.MiddleName)
	input.PassportMasked = maskPassport(input.PassportSeries, input.PassportNumber)
	input.TaxNumberMasked = maskTaxNumber(input.TaxNumber)
	input.GenderDisplay = genderToUkrainian(input.Gender)
	input.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
	input.UpdatedAt = updatedAt.Format("2006-01-02 15:04:05")

	return input
}

// normalizeDate strips time from ISO date string
func normalizeDate(d string) string {
	if idx := strings.Index(d, "T"); idx != -1 {
		return d[:idx]
	}
	return d
}
