package services

import (
	"context"
	"passport-desk-mvp/internal/models"
	"testing"
)

func TestAuditLogService_LogAudit(t *testing.T) {
	mockRepo := &MockAuditLogRepo{}

	// Case 1: Unauthenticated
	sessionSvc := &SessionService{}
	service := NewAuditLogService(mockRepo, sessionSvc)

	log := &models.AuditLog{ActionType: "CREATE", TableName: "test"}
	err := service.LogAudit(context.Background(), log)
	if err == nil || err.Error() != "not authenticated" {
		t.Errorf("Expected not authenticated error, got %v", err)
	}

	// Case 2: Authenticated, operator ID should be filled
	sessionSvc = &SessionService{
		CurrentOperator: &models.Operator{ID: 100},
	}
	service = NewAuditLogService(mockRepo, sessionSvc)

	repoCalled := false
	mockRepo.LogAuditFunc = func(ctx context.Context, l *models.AuditLog) error {
		repoCalled = true
		if l.OperatorID != 100 {
			t.Errorf("Expected OperatorID 100, got %d", l.OperatorID)
		}
		return nil
	}

	err = service.LogAudit(context.Background(), &models.AuditLog{ActionType: "UPDATE"})
	if err != nil {
		t.Fatalf("LogAudit failed: %v", err)
	}
	if !repoCalled {
		t.Error("Repo LogAudit was not called")
	}
}

func TestAuditLogService_GetAuditLogs(t *testing.T) {
	mockRepo := &MockAuditLogRepo{}
	sessionSvc := &SessionService{
		CurrentOperator: &models.Operator{ID: 1},
	}
	service := NewAuditLogService(mockRepo, sessionSvc)

	mockRepo.GetAuditLogsFunc = func(ctx context.Context, limit int) ([]models.AuditLogOutput, error) {
		if limit != 100 { // Default limit if invalid passed
			t.Errorf("Expected limit 100, got %d", limit)
		}
		return []models.AuditLogOutput{{AuditLog: models.AuditLog{ID: 1}}}, nil
	}

	logs, err := service.GetAuditLogs(context.Background(), 0)
	if err != nil {
		t.Fatalf("GetAuditLogs failed: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}
}
