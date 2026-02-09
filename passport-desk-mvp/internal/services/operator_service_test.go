package services

import (
	"context"
	"log/slog"
	"os"
	"passport-desk-mvp/internal/logger"
	"passport-desk-mvp/internal/models"
	"testing"
)

func init() {
	logger.Log = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func TestOperatorService_GetCurrentOperator(t *testing.T) {
	mockRepo := &MockOperatorRepo{}

	// Case 1: Not logged in
	sessionService := NewSessionService(nil)
	service := NewOperatorService(sessionService, mockRepo)

	opInfo := service.GetCurrentOperator(context.Background())
	if opInfo != nil {
		t.Error("Expected nil operator info when not logged in")
	}

	// Case 2: Logged in
	operator := &models.Operator{
		ID:       1,
		Username: "admin",
		FullName: "Administrator",
	}
	sessionService = NewSessionService(operator)
	sessionService.IsLocked = false
	service = NewOperatorService(sessionService, mockRepo)

	opInfo = service.GetCurrentOperator(context.Background())
	if opInfo == nil || opInfo.Username != "admin" {
		t.Errorf("Unexpected operator info: %v", opInfo)
	}
}

func TestOperatorService_CreateOperator(t *testing.T) {
	mockRepo := &MockOperatorRepo{}
	sessionService := NewSessionService(nil)
	service := NewOperatorService(sessionService, mockRepo)
	ctx := context.Background()

	op := &models.Operator{Username: "newuser"}
	called := false
	mockRepo.CreateFunc = func(ctx context.Context, o *models.Operator) error {
		called = true
		if o.Username != "newuser" {
			t.Errorf("Expected newuser, got %s", o.Username)
		}
		return nil
	}

	err := service.CreateOperator(ctx, op)
	if err != nil {
		t.Fatalf("CreateOperator failed: %v", err)
	}
	if !called {
		t.Error("Mock repository Create was not called")
	}
}

func TestOperatorService_GetOperatorByUsername(t *testing.T) {
	mockRepo := &MockOperatorRepo{}
	sessionService := NewSessionService(nil)
	service := NewOperatorService(sessionService, mockRepo)
	ctx := context.Background()

	mockRepo.GetByUsernameFunc = func(ctx context.Context, username string) (*models.Operator, error) {
		if username == "admin" {
			return &models.Operator{ID: 1, Username: "admin"}, nil
		}
		return nil, nil
	}

	op, err := service.GetOperatorByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("GetOperatorByUsername failed: %v", err)
	}
	if op == nil || op.Username != "admin" {
		t.Errorf("Unexpected operator: %v", op)
	}
}

func TestOperatorService_OperatorExists(t *testing.T) {
	mockRepo := &MockOperatorRepo{}
	sessionService := NewSessionService(nil)
	service := NewOperatorService(sessionService, mockRepo)
	ctx := context.Background()

	mockRepo.ExistsFunc = func(ctx context.Context) (bool, error) {
		return true, nil
	}

	exists, err := service.OperatorExists(ctx)
	if err != nil {
		t.Fatalf("OperatorExists failed: %v", err)
	}
	if !exists {
		t.Error("Expected OperatorExists to return true")
	}
}
