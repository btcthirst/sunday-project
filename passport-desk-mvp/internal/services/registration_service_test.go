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

func TestRegistrationService_Create(t *testing.T) {
	mockRepo := &MockRegistrationRepo{}
	crypto := security.NewCrypto([]byte("12345678901234567890123456789012"))
	mockAuditRepo := &MockAuditLogRepo{}
	sessionService := &SessionService{
		CurrentOperator: &models.Operator{ID: 1, Username: "test"},
		IsLocked:        false,
	}
	auditService := NewAuditLogService(mockAuditRepo, sessionService)

	service := NewRegistrationService(mockRepo, crypto, auditService)
	// Inject logger into context
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := internalLogger.NewContext(context.Background(), logger)

	mockRepo.CreateFunc = func(ctx context.Context, input *models.RegistrationInput) (*models.RegistrationOutput, error) {
		return &models.RegistrationOutput{ID: 10, CitizenID: input.CitizenID}, nil
	}

	auditCalled := false
	mockAuditRepo.LogAuditFunc = func(ctx context.Context, log *models.AuditLog) error {
		auditCalled = true
		if log.ActionType != "CREATE" {
			t.Errorf("Expected CREATE audit, got %s", log.ActionType)
		}
		return nil
	}

	input := &models.RegistrationInput{
		CitizenID:        5,
		RegistrationType: "permanent",
		Settlement:       "City",
	}

	reg, err := service.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if reg.ID != 10 {
		t.Errorf("Expected ID 10, got %d", reg.ID)
	}

	if !auditCalled {
		t.Error("Expected audit log call")
	}
}

func TestRegistrationService_Deregister(t *testing.T) {
	mockRepo := &MockRegistrationRepo{}
	mockAuditRepo := &MockAuditLogRepo{}
	sessionService := &SessionService{
		CurrentOperator: &models.Operator{ID: 1, Username: "test"},
		IsLocked:        false,
	}
	auditService := NewAuditLogService(mockAuditRepo, sessionService)
	// RegistrationService needs crypto? Yes, injected.
	crypto := security.NewCrypto([]byte("12345678901234567890123456789012"))

	service := NewRegistrationService(mockRepo, crypto, auditService)
	// Inject logger into context
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := internalLogger.NewContext(context.Background(), logger)

	mockRepo.DeregisterFunc = func(ctx context.Context, id int64, date string) error {
		if id != 10 {
			t.Errorf("Expected ID 10, got %d", id)
		}
		return nil
	}

	err := service.Deregister(ctx, 10, "2023-01-01")
	if err != nil {
		t.Fatalf("Deregister failed: %v", err)
	}
}

func TestRegistrationService_ListAll(t *testing.T) {
	mockRepo := &MockRegistrationRepo{}
	crypto := security.NewCrypto([]byte("12345678901234567890123456789012"))
	mockAuditRepo := &MockAuditLogRepo{}
	service := NewRegistrationService(mockRepo, crypto, NewAuditLogService(mockAuditRepo, nil))
	ctx := internalLogger.NewContext(context.Background(), slog.New(slog.NewTextHandler(os.Stdout, nil)))

	encPhone, _ := crypto.Encrypt("123-456")
	encTax, _ := crypto.Encrypt("999888")

	mockRepo.ListAllFunc = func(ctx context.Context, search string, isActive *bool, page, limit int) (*models.RegistrationListResult, error) {
		return &models.RegistrationListResult{
			Total: 1,
			Items: []models.RegistrationListItem{
				{
					RegistrationOutput: models.RegistrationOutput{
						ID:        1,
						CitizenID: 5,
					},
					CitizenName:  "John",
					CitizenPhone: encPhone,
					CitizenTax:   encTax,
				},
			},
		}, nil
	}

	result, err := service.ListAll(ctx, "", nil, 1, 10)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	if result.Items[0].CitizenPhone != "123-456" {
		t.Errorf("Expected decrypted phone 123-456, got %s", result.Items[0].CitizenPhone)
	}
}
