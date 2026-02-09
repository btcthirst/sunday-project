package services

import (
	"context"
	"log/slog"
	"os"
	internalLogger "passport-desk-mvp/internal/logger"
	"passport-desk-mvp/internal/models"
	"passport-desk-mvp/internal/security"
	"testing"
)

func TestCitizenService_Create(t *testing.T) {
	// Setup dependencies
	mockRepo := &MockCitizenRepo{}

	// Real crypto with dummy key
	crypto := security.NewCrypto([]byte("12345678901234567890123456789012")) // 32 bytes

	mockAuditRepo := &MockAuditLogRepo{}
	sessionService := &SessionService{
		CurrentOperator: &models.Operator{ID: 1, Username: "test"},
		IsLocked:        false,
	}
	auditService := NewAuditLogService(mockAuditRepo, sessionService)

	service := NewCitizenService(crypto, mockRepo, auditService)
	// Inject logger into context
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := internalLogger.NewContext(context.Background(), logger)

	// Mock behavior
	mockRepo.CreateFunc = func(ctx context.Context, citizen *models.CitizenInput) (int64, error) {
		// Verify encryption happened
		if citizen.PassportNumber == "123456" {
			t.Error("Passport number should be encrypted")
		}
		return 1, nil
	}
	mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.CitizenOutput, error) {
		return &models.CitizenOutput{ID: id}, nil
	}

	input := &models.CitizenInput{
		LastName:       "Test",
		FirstName:      "User",
		BirthDate:      "1990-01-01",
		PassportSeries: "AA",
		PassportNumber: "123456",
		PassportType:   "old",
		TaxNumber:      "1234567890",
		Gender:         "M",
		Phone:          "123",
	}

	result, err := service.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
}

func TestCitizenService_GetByID(t *testing.T) {
	mockRepo := &MockCitizenRepo{}
	crypto := security.NewCrypto([]byte("12345678901234567890123456789012"))
	mockAuditRepo := &MockAuditLogRepo{}
	sessionService := &SessionService{
		CurrentOperator: &models.Operator{ID: 1, Username: "test"},
		IsLocked:        false,
	}
	auditService := NewAuditLogService(mockAuditRepo, sessionService)

	service := NewCitizenService(crypto, mockRepo, auditService)
	// Inject logger into context
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := internalLogger.NewContext(context.Background(), logger)

	// Prepare encrypted data
	encPhone, _ := crypto.Encrypt("555-0100")

	mockRepo.GetByIDFunc = func(ctx context.Context, id int64) (*models.CitizenOutput, error) {
		return &models.CitizenOutput{
			ID:             id,
			LastName:       "Doe",
			FirstName:      "John",
			BirthDate:      "1990-01-01T00:00:00Z",
			PassportSeries: "AA", // Mocking as if encrypted strings are just placeholder strings for now unless we really encrypt
			Phone:          encPhone,
		}, nil
	}

	// Verify audit log call
	auditCalled := false
	mockAuditRepo.LogAuditFunc = func(ctx context.Context, log *models.AuditLog) error {
		auditCalled = true
		if log.ActionType != "READ" {
			t.Errorf("Expected action READ, got %s", log.ActionType)
		}
		return nil
	}

	citizen, err := service.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if citizen.Phone != "555-0100" {
		t.Errorf("Expected decrypted phone 555-0100, got %s", citizen.Phone)
	}

	if !auditCalled {
		t.Error("Expected audit log to be called")
	}
}
