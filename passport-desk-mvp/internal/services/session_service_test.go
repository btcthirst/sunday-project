package services

import (
	"passport-desk-mvp/internal/models"
	"testing"
	"time"
)

func TestSessionService_Basic(t *testing.T) {
	operator := &models.Operator{
		ID:           1,
		Username:     "admin",
		PasswordHash: "hashed",
	}

	service := NewSessionService(operator)
	service.IsLocked = false
	service.LastActivity = time.Now()

	if !service.IsAuthenticated() {
		t.Error("Expected authenticated status")
	}

	op := service.GetCurrentOperator()
	if op == nil || op.Username != "admin" {
		t.Errorf("Unexpected current operator: %v", op)
	}

	// Lock
	service.Mu.Lock()
	service.IsLocked = true
	service.Mu.Unlock()

	if service.IsAuthenticated() {
		t.Error("Expected unauthenticated status after locking")
	}
}

func TestSessionService_Inactivity(t *testing.T) {
	operator := &models.Operator{ID: 1, Username: "admin"}
	service := NewSessionService(operator)
	service.IsLocked = false
	service.UpdateActivity()
	if !service.IsAuthenticated() {
		t.Error("Should be authenticated after update activity")
	}
}
