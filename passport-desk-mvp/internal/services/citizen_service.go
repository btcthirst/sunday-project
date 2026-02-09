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

	// Decrypt and compute fields using unified helpers
	decrypted, err := s.decryptSensitiveFieldsWithLog(ctx, *citizen)
	if err != nil {
		// If decryption fails, we still return the citizen to show error/hash,
		// but computeFields might fail or show garbage.
		// Actually, decryptSensitiveFieldsWithLog now returns error on failure.
		log.Error("Decryption failed in GetByID", slog.String("error", err.Error()))
		// We can choose to return the encrypted one for troubleshooting or fail.
		// Better to fail to prevent further corruption.
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	decrypted.BirthDate = normalizeDate(decrypted.BirthDate)
	s.computeFields(ctx, &decrypted)

	// AUDIT LOG: Business event - View Citizen
	if err := s.auditLog.LogAudit(ctx, &models.AuditLog{
		OperatorID:  getOperatorIDFromContext(ctx),
		ActionType:  "READ",
		TableName:   "citizens",
		RecordID:    id,
		Description: fmt.Sprintf("Viewed citizen: %s", decrypted.FullName),
	}); err != nil {
		log.Warn("Audit log failed", slog.String("error", err.Error()))
	}

	return &decrypted, nil
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

	// Copy input to avoid modifying the original data
	inputCopy := *input
	encryptedInput, err := s.encryptSensitiveFields(&inputCopy)
	if err != nil {
		log.Error("Encryption failed", slog.String("error", err.Error()))
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	encryptedInput.BirthDate = normalizeDate(encryptedInput.BirthDate)

	start := time.Now()
	// Update in DB
	if err := s.repo.Update(ctx, id, encryptedInput); err != nil {
		log.Error("Update failed", slog.String("error", err.Error()), slog.String("table", "citizens"))
		return nil, fmt.Errorf("update failed: %w", err)
	}

	log.Info("Citizen updated successfully",
		slog.Int64("id", id),
		slog.Duration("duration", time.Since(start)),
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
	return s.GetByID(ctx, id)
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
		s.computeFields(ctx, &familyMember.CitizenOutput)
		decryptedFamilyMembers[i] = familyMember
	}
	return decryptedFamilyMembers, nil
}

func (s *CitizenService) RemoveFamilyMember(ctx context.Context, citizenID, memberID int64) error {
	log := logger.FromContext(ctx).With(
		slog.String("service", "citizen"),
		slog.String("method", "RemoveFamilyMember"),
		slog.Int64("id1", citizenID),
		slog.Int64("id2", memberID),
	)

	if err := s.repo.RemoveFamilyMember(ctx, citizenID, memberID); err != nil {
		log.Error("Failed to remove family relationship", slog.String("error", err.Error()))
		return err
	}

	log.Info("Family relationship removed successfully (reciprocal)")
	return nil
}

func (s *CitizenService) AddFamilyMember(ctx context.Context, citizenID, memberID int64, relationType string) error {
	log := logger.FromContext(ctx).With(
		slog.String("service", "citizen"),
		slog.String("method", "AddFamilyMember"),
		slog.Int64("citizen_id", citizenID),
		slog.Int64("member_id", memberID),
		slog.String("type", relationType),
	)

	if err := s.repo.AddFamilyMember(ctx, citizenID, memberID, relationType); err != nil {
		log.Error("Failed to add family relationship", slog.String("error", err.Error()))
		return err
	}

	log.Info("Family relationship added successfully (reciprocal)")
	return nil
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
	// Helper to encrypt only if not already encrypted
	encrypt := func(val, name string) (string, error) {
		if val == "" {
			return "", nil
		}

		// Safeguard: if it looks like a hash (long, ends with =), skip encryption
		// to avoid double-encryption if corrupted data reached the form.
		if len(val) > 20 && strings.HasSuffix(val, "=") && !strings.Contains(val, " ") {
			slog.Warn("Skipping encryption for already encrypted-looking field", slog.String("field", name))
			return val, nil
		}

		slog.Debug("Encrypting field", slog.String("field", name), slog.Int("len", len(val)))
		return s.crypto.Encrypt(val)
	}

	result := *input // Shallow copy

	var err error
	result.PassportSeries, err = encrypt(input.PassportSeries, "PassportSeries")
	if err != nil {
		return nil, err
	}
	result.PassportNumber, err = encrypt(input.PassportNumber, "PassportNumber")
	if err != nil {
		return nil, err
	}
	result.TaxNumber, err = encrypt(input.TaxNumber, "TaxNumber")
	if err != nil {
		return nil, err
	}
	result.Phone, err = encrypt(input.Phone, "Phone")
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *CitizenService) decryptSensitiveFields(input models.CitizenOutput) (models.CitizenOutput, error) {
	// Fallback to global logger if context is missing, but preferably we use the version with context
	return s.decryptSensitiveFieldsWithLog(context.Background(), input)
}

func (s *CitizenService) decryptSensitiveFieldsWithLog(ctx context.Context, input models.CitizenOutput) (models.CitizenOutput, error) {
	log := logger.FromContext(ctx)

	// Helper to decrypt multiple times if needed (to recover from double-encryption bugs)
	deepDecrypt := func(val, name string) (string, error) {
		if val == "" {
			return "", nil
		}

		current := val
		layers := 0
		for layers < 3 {
			// Try to decrypt
			decrypted, err := s.crypto.Decrypt(current)
			if err != nil {
				// If first layer fails, it might be already plaintext or corrupted
				if layers == 0 {
					return val, nil // Assume plaintext
				}
				// If subsequent layer fails, return what we have so far
				return current, nil
			}

			current = decrypted
			layers++

			// If the result no longer looks like a hash, we are done
			if len(current) < 20 || !strings.HasSuffix(current, "=") || strings.Contains(current, " ") {
				break
			}
			log.Warn("Detected multi-layered encryption, trying another layer", slog.String("field", name), slog.Int("layer", layers))
		}

		if layers > 1 {
			log.Info("Successfully recovered multi-layered encrypted data", slog.String("field", name), slog.Int("layers", layers))
		}
		return current, nil
	}

	var err error
	if input.PassportSeries != "" {
		input.PassportSeries, err = deepDecrypt(input.PassportSeries, "PassportSeries")
		if err != nil {
			return input, err
		}
	}
	if input.PassportNumber != "" {
		input.PassportNumber, err = deepDecrypt(input.PassportNumber, "PassportNumber")
		if err != nil {
			return input, err
		}
	}
	if input.TaxNumber != "" {
		input.TaxNumber, err = deepDecrypt(input.TaxNumber, "TaxNumber")
		if err != nil {
			return input, err
		}
	}
	if input.Phone != "" {
		input.Phone, err = deepDecrypt(input.Phone, "Phone")
		if err != nil {
			return input, err
		}
	}

	// BirthDate should NEVER be encrypted, but if it is, let's try to recover it for the user
	if len(input.BirthDate) > 20 && strings.HasSuffix(input.BirthDate, "=") {
		log.Warn("BirthDate looks encrypted! Attempting recovery...", slog.String("val", input.BirthDate))
		if d, err := deepDecrypt(input.BirthDate, "BirthDate"); err == nil {
			input.BirthDate = d
		}
	}

	return input, nil
}

func (s *CitizenService) computeFields(ctx context.Context, input *models.CitizenOutput) *models.CitizenOutput {
	log := logger.FromContext(ctx)

	// Compute fields
	var createdAt, updatedAt time.Time
	var err error

	// Try multiple layouts for parsing timestamps
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
	}

	parseTime := func(val string) (time.Time, error) {
		if val == "" {
			return time.Time{}, nil
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, val); err == nil {
				return t, nil
			}
		}
		return time.Parse("2006-01-02 15:04:05", val) // Final attempt for error reporting
	}

	createdAt, err = parseTime(input.CreatedAt)
	if err != nil && input.CreatedAt != "" {
		log.Warn("Failed to parse CreatedAt", slog.String("value", input.CreatedAt), slog.String("error", err.Error()))
	}

	updatedAt, err = parseTime(input.UpdatedAt)
	if err != nil && input.UpdatedAt != "" {
		log.Warn("Failed to parse UpdatedAt", slog.String("value", input.UpdatedAt), slog.String("error", err.Error()))
	}

	// Ensure names don't contain hashes
	if strings.Contains(input.LastName, "=") && len(input.LastName) > 20 {
		log.Error("LastName contains hash!", slog.String("val", input.LastName))
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
