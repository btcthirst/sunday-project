package services

import (
	"context"
	"log/slog"
	"os"
	"passport-desk-mvp/internal/logger"
	"passport-desk-mvp/internal/models"
	"testing"
)

func TestAuditLogService_LogAudit(t *testing.T) {
	mockRepo := &MockAuditLogRepo{}
	sessionService := NewSessionService(&models.Operator{ID: 1, Username: "admin"})
	sessionService.IsLocked = false
	service := NewAuditLogService(mockRepo, sessionService)

	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := logger.NewContext(context.Background(), l)

	called := false
	mockRepo.LogAuditFunc = func(ctx context.Context, log *models.AuditLog) error {
		called = true
		if log.ActionType != "TEST_ACTION" {
			t.Errorf("Expected action TEST_ACTION, got %s", log.ActionType)
		}
		return nil
	}

	err := service.LogAudit(ctx, &models.AuditLog{
		ActionType:  "TEST_ACTION",
		TableName:   "test_table",
		RecordID:    1,
		Description: "Testing log",
	})
	if err != nil {
		t.Fatalf("LogAudit failed: %v", err)
	}

	if !called {
		t.Error("Mock repo's LogAudit was not called")
	}
}

func TestAuditLogService_GetAuditLogs(t *testing.T) {
	mockRepo := &MockAuditLogRepo{}
	sessionService := NewSessionService(&models.Operator{ID: 1, Username: "admin"})
	sessionService.IsLocked = false
	service := NewAuditLogService(mockRepo, sessionService)

	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := logger.NewContext(context.Background(), l)

	mockRepo.GetAuditLogsFunc = func(ctx context.Context, limit int) ([]models.AuditLogOutput, error) {
		return []models.AuditLogOutput{
			{AuditLog: models.AuditLog{ID: 1, ActionType: "LOGIN"}},
			{AuditLog: models.AuditLog{ID: 2, ActionType: "CREATE"}},
		}, nil
	}

	logs, err := service.GetAuditLogs(ctx, 10)
	if err != nil {
		t.Fatalf("GetAuditLogs failed: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logs))
	}
}
