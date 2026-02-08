package repository

import (
	"context"
	"passport-desk-mvp/internal/models"
	"testing"
)

func TestAuditLogRepository_LogAudit(t *testing.T) {
	db := setupTestDB(t)
	// Create operator first
	opRepo := NewOperatorRepository(db)
	auditRepo := NewAuditLogRepository(db)
	ctx := context.Background()

	op := &models.Operator{
		Username:     "testop",
		PasswordHash: "hash",
		FullName:     "Test Operator",
	}
	err := opRepo.Create(ctx, op)
	if err != nil {
		t.Fatalf("Failed to create op: %v", err)
	}

	// Fetch op to get ID
	createdOp, err := opRepo.GetByUsername(ctx, "testop")
	if err != nil {
		t.Fatalf("Failed to get op: %v", err)
	}

	log := &models.AuditLog{
		OperatorID:  createdOp.ID,
		ActionType:  "TEST_ACTION",
		TableName:   "test_table",
		RecordID:    1,
		Description: "Test Description",
	}

	err = auditRepo.LogAudit(ctx, log)
	if err != nil {
		t.Fatalf("Failed to log audit: %v", err)
	}

	// Verify retrieval
	logs, err := auditRepo.GetAuditLogs(ctx, 10)
	if err != nil {
		t.Fatalf("Failed to get logs: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	if logs[0].ActionType != "TEST_ACTION" {
		t.Errorf("Expected TEST_ACTION, got %s", logs[0].ActionType)
	}
}
