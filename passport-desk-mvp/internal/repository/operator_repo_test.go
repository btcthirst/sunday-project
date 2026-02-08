package repository

import (
	"context"
	"passport-desk-mvp/internal/models"
	"testing"
)

func TestOperatorRepository_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOperatorRepository(db)
	ctx := context.Background()

	op := &models.Operator{
		Username:     "admin",
		PasswordHash: "hashed_secret",
		FullName:     "Administrator",
	}

	err := repo.Create(ctx, op)
	if err != nil {
		t.Fatalf("Failed to create operator: %v", err)
	}

	fetchedOp, err := repo.GetByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("Failed to get operator: %v", err)
	}

	if fetchedOp.FullName != "Administrator" {
		t.Errorf("Expected Administrator, got %s", fetchedOp.FullName)
	}
}

func TestOperatorRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOperatorRepository(db)
	ctx := context.Background()

	exists, err := repo.Exists(ctx)
	if err != nil {
		t.Fatalf("Failed to check exists: %v", err)
	}
	if exists {
		t.Errorf("Expected false on empty DB")
	}

	repo.Create(ctx, &models.Operator{Username: "u", PasswordHash: "p", FullName: "f"})

	exists, err = repo.Exists(ctx)
	if err != nil {
		t.Fatalf("Failed to check exists: %v", err)
	}
	if !exists {
		t.Errorf("Expected true after creation")
	}
}
