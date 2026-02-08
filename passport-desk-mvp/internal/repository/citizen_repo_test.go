package repository

import (
	"context"
	"passport-desk-mvp/internal/models"
	"testing"
)

func TestCitizenRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)

	ctx := context.Background()

	input := &models.CitizenInput{
		LastName:       "Test",
		FirstName:      "User",
		BirthDate:      "1990-01-01",
		PassportSeries: "AA",
		PassportNumber: "123456",
		TaxNumber:      "1234567890",
		Gender:         "M",
		Phone:          "123456789",
		Email:          "test@example.com",
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Failed to create citizen: %v", err)
	}

	if id == 0 {
		t.Fatalf("Expected valid ID, got 0")
	}

	// Verify retrieval
	citizen, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get citizen: %v", err)
	}

	if citizen.LastName != input.LastName {
		t.Errorf("Expected Last Name %s, got %s", input.LastName, citizen.LastName)
	}
}

func TestCitizenRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)
	ctx := context.Background()

	input := &models.CitizenInput{
		LastName:       "Original",
		FirstName:      "Name",
		BirthDate:      "1990-01-01",
		PassportSeries: "AA",
		PassportNumber: "123456",
		Gender:         "M",
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Failed to create: %v", err)
	}

	input.LastName = "Updated"
	err = repo.Update(ctx, id, input)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	updated, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get citizen: %v", err)
	}

	if updated.LastName != "Updated" {
		t.Errorf("Update failed, expected 'Updated', got %s", updated.LastName)
	}
}

func TestCitizenRepository_SoftDeleteAndRestore(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)
	ctx := context.Background()

	input := &models.CitizenInput{
		LastName:       "Delete",
		FirstName:      "Me",
		BirthDate:      "1990-01-01",
		PassportSeries: "AA",
		PassportNumber: "123456",
		Gender:         "M",
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Failed to create: %v", err)
	}

	// Soft Delete
	err = repo.SoftDelete(ctx, id)
	if err != nil {
		t.Fatalf("Failed to soft delete: %v", err)
	}

	// Verify deleted flag
	citizen, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get citizen: %v", err)
	}
	if !citizen.Deleted {
		t.Errorf("Expected citizen to be deleted")
	}

	// Restore
	err = repo.Restore(ctx, id)
	if err != nil {
		t.Fatalf("Failed to restore: %v", err)
	}

	citizen, err = repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get citizen: %v", err)
	}
	if citizen.Deleted {
		t.Errorf("Expected citizen to be restored")
	}
}

func TestCitizenRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)
	ctx := context.Background()

	// Create 2 citizens
	repo.Create(ctx, &models.CitizenInput{LastName: "A", FirstName: "User", BirthDate: "2000-01-01", PassportSeries: "AA", PassportNumber: "111", Gender: "M"})
	repo.Create(ctx, &models.CitizenInput{LastName: "B", FirstName: "User", BirthDate: "2000-01-02", PassportSeries: "AB", PassportNumber: "222", Gender: "F"})

	list, total, err := repo.List(ctx, 0, 10, false)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}
	if len(list) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list))
	}
}

func TestCitizenRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &models.CitizenInput{LastName: "Smith", FirstName: "John", BirthDate: "1980-05-15", PassportSeries: "S", PassportNumber: "1", Gender: "M"})
	repo.Create(ctx, &models.CitizenInput{LastName: "Doe", FirstName: "Jane", BirthDate: "1990-10-20", PassportSeries: "D", PassportNumber: "1", Gender: "F"})

	// Search by name
	results, err := repo.SearchByName(ctx, "Smi", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].LastName != "Smith" {
		t.Errorf("Search failed, expected Smith, got %v", results)
	}

	// Search by birth date
	dateResults, err := repo.SearchByBirthDate(ctx, "1990-10-20", 10)
	if err != nil {
		t.Fatalf("Search by date failed: %v", err)
	}
	if len(dateResults) != 1 || dateResults[0].FirstName != "Jane" {
		t.Errorf("Search failed, expected Jane, got %v", dateResults)
	}
}

func TestCitizenRepository_ContextCancellation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := repo.GetByID(ctx, 1)
	if err == nil {
		t.Error("Expected error due to cancelled context, got nil")
	}
}
