package services

import (
	"context"
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
	repo      repository.CitizenRepositoryInterface
	validator *CitizenValidator
	auditLog  *AuditLogService
}

// NewCitizenService creates a new citizen service
func NewCitizenService(crypto *security.Crypto, repo repository.CitizenRepositoryInterface, auditLog *AuditLogService) *CitizenService {
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
	return s.GetByID(ctx, id)
}

func (s *CitizenService) GetByID(ctx context.Context, id int64) (*models.CitizenOutput, error) {
	log := logger.FromContext(ctx).With(
		slog.String("service", "citizen"),
		slog.String("method", "GetByID"),
		slog.Int64("id", id),
	)

	citizen, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get citizen", slog.String("error", err.Error()))
		return nil, err
	}

	if d, err := s.crypto.Decrypt(citizen.Phone); err == nil {
		citizen.Phone = d
	} else {
		log.Warn("Failed to decrypt phone", slog.String("error", err.Error()))
	}

	if d, err := s.crypto.Decrypt(citizen.Email); err == nil {
		citizen.Email = d
	} else {
		log.Warn("Failed to decrypt email", slog.String("error", err.Error()))
	}

	if d, err := s.crypto.Decrypt(citizen.TaxNumber); err == nil {
		citizen.TaxNumber = d
	} else {
		log.Warn("Failed to decrypt tax number", slog.String("error", err.Error()))
	}

	if d, err := s.crypto.Decrypt(citizen.PassportNumber); err == nil {
		citizen.PassportNumber = d
	} else {
		log.Warn("Failed to decrypt passport number", slog.String("error", err.Error()))
	}

	// AUDIT LOG: Business event - View Citizen
	if err := s.auditLog.LogAudit(ctx, &models.AuditLog{
		OperatorID:  getOperatorIDFromContext(ctx),
		ActionType:  "READ",
		TableName:   "citizens",
		RecordID:    id,
		Description: fmt.Sprintf("Viewed citizen: %s", citizen.FullName),
	}); err != nil {
		log.Warn("Audit log failed", slog.String("error", err.Error()))
	}

	return citizen, nil
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
	log := logger.FromContext(ctx).With(
		slog.String("service", "citizen"),
		slog.String("method", "List"),
	)

	citizenList, total, err := s.repo.List(ctx, offset, limit, includeDeleted)
	if err != nil {
		log.Error("List failed", slog.String("error", err.Error()))
		return nil, err
	}

	for i := range citizenList {
		decrypted, _ := s.decryptSensitiveFieldsWithLog(ctx, *citizenList[i])
		*citizenList[i] = decrypted
		s.computeFields(ctx, citizenList[i])
	}

	return &models.CitizenListResult{
		Total: total,
		Items: citizenList,
	}, nil
}

func (s *CitizenService) Search(ctx context.Context, query string, field string) ([]*models.CitizenOutput, error) {
	log := logger.FromContext(ctx).With(
		slog.String("service", "citizen"),
		slog.String("method", "Search"),
		slog.String("field", field),
		slog.String("query", query),
	)
	log.Info("Searching citizens")
	var results []*models.CitizenOutput
	var err error
	switch field {
	case "name":
		results, err = s.repo.SearchByName(ctx, query, 10)
	case "birth_date":
		results, err = s.repo.SearchByBirthDate(ctx, query, 10)
	default:
		log.Error("Invalid search field", slog.String("field", field))
		return nil, errors.New("invalid field")
	}

	if err != nil {
		log.Error("Search failed", slog.String("error", err.Error()))
		return nil, err
	}

	for i := range results {
		*results[i], _ = s.decryptSensitiveFieldsWithLog(ctx, *results[i])
		s.computeFields(ctx, results[i])
	}
	return results, nil
}

func (s *CitizenService) SearchByName(ctx context.Context, searchTerm string, limit int) ([]*models.CitizenOutput, error) {
	results, err := s.repo.SearchByName(ctx, searchTerm, limit)
	if err != nil {
		return nil, err
	}
	for i := range results {
		*results[i], _ = s.decryptSensitiveFieldsWithLog(ctx, *results[i])
		s.computeFields(ctx, results[i])
	}
	return results, nil
}

func (s *CitizenService) SearchByBirthDate(ctx context.Context, birthDate string, limit int) ([]*models.CitizenOutput, error) {
	results, err := s.repo.SearchByBirthDate(ctx, birthDate, limit)
	if err != nil {
		return nil, err
	}
	for i := range results {
		*results[i], _ = s.decryptSensitiveFieldsWithLog(ctx, *results[i])
		s.computeFields(ctx, results[i])
	}
	return results, nil
}

func (s *CitizenService) Exists(ctx context.Context, id int64) (bool, error) {
	return s.repo.Exists(ctx, id)
}

func (s *CitizenService) GetByPassport(ctx context.Context, passportSeries, passportNumber string) (*models.CitizenOutput, error) {
	citizen, err := s.repo.GetByPassport(ctx, passportSeries, passportNumber)
	if err != nil {
		return nil, err
	}
	if citizen != nil {
		*citizen, _ = s.decryptSensitiveFieldsWithLog(ctx, *citizen)
		s.computeFields(ctx, citizen)
	}
	return citizen, nil
}

func (s *CitizenService) GetAll(ctx context.Context) ([]*models.CitizenOutput, error) {
	results, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range results {
		*results[i], _ = s.decryptSensitiveFieldsWithLog(ctx, *results[i])
		s.computeFields(ctx, results[i])
	}
	return results, nil
}

func (s *CitizenService) GetFamilyMembers(ctx context.Context, citizenID int64) ([]*models.FamilyMemberOutput, error) {
	log := logger.FromContext(ctx).With(
		slog.String("service", "citizen"),
		slog.String("method", "GetFamilyMembers"),
		slog.Int64("citizen_id", citizenID),
	)

	familyMembers, err := s.repo.GetFamilyMembers(ctx, citizenID)
	if err != nil {
		log.Error("Failed to get family members", slog.String("error", err.Error()))
		return nil, err
	}
	decryptedFamilyMembers := make([]*models.FamilyMemberOutput, len(familyMembers))
	for i, familyMember := range familyMembers {
		familyMember.BirthDate = normalizeDate(familyMember.BirthDate)
		familyMember.CitizenOutput, err = s.decryptSensitiveFieldsWithLog(ctx, familyMember.CitizenOutput)
		if err != nil {
			// Technical error already logged in decryptSensitiveFieldsWithLog?
			// No, decryptSensitiveFieldsWithLog might not log error if we want it to be silent but it returns err.
			// Let's make it log.
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
	// Fallback to global logger if context is missing, but preferably we use the version with context
	return s.decryptSensitiveFieldsWithLog(context.Background(), input)
}

func (s *CitizenService) decryptSensitiveFieldsWithLog(ctx context.Context, input models.CitizenOutput) (models.CitizenOutput, error) {
	log := logger.FromContext(ctx)

	if d, err := s.crypto.Decrypt(input.PassportSeries); err == nil {
		input.PassportSeries = d
	} else {
		log.Warn("Failed to decrypt passport series", slog.String("error", err.Error()))
	}

	if d, err := s.crypto.Decrypt(input.PassportNumber); err == nil {
		input.PassportNumber = d
	} else {
		log.Warn("Failed to decrypt passport number", slog.String("error", err.Error()))
	}

	if d, err := s.crypto.Decrypt(input.TaxNumber); err == nil {
		input.TaxNumber = d
	} else {
		log.Warn("Failed to decrypt tax number", slog.String("error", err.Error()))
	}

	if d, err := s.crypto.Decrypt(input.Phone); err == nil {
		input.Phone = d
	} else {
		log.Warn("Failed to decrypt phone", slog.String("error", err.Error()))
	}

	return input, nil
}

func (s *CitizenService) computeFields(ctx context.Context, input *models.CitizenOutput) *models.CitizenOutput {
	log := logger.FromContext(ctx)

	// Compute fields
	var createdAt, updatedAt time.Time
	var err error
	createdAt, err = time.Parse("2006-01-02 15:04:05", input.CreatedAt)
	if err != nil {
		log.Warn("Failed to parse CreatedAt", slog.String("value", input.CreatedAt), slog.String("error", err.Error()))
	}

	updatedAt, err = time.Parse("2006-01-02 15:04:05", input.UpdatedAt)
	if err != nil {
		log.Warn("Failed to parse UpdatedAt", slog.String("value", input.UpdatedAt), slog.String("error", err.Error()))
	}
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
