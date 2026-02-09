package repository

import (
	"context"
	"passport-desk-mvp/internal/models"
	"testing"
)

func TestReportRepository_ExportRegisteredCitizens(t *testing.T) {
	db := setupTestDB(t)
	citizenRepo := NewCitizenRepository(db)
	regRepo := NewRegistrationRepository(db)
	repo := NewReportRepository(db)
	ctx := context.Background()

	c1, _ := citizenRepo.Create(ctx, &models.CitizenInput{LastName: "C1", FirstName: "U", BirthDate: "2000-01-01", Gender: "M"})
	c2, _ := citizenRepo.Create(ctx, &models.CitizenInput{LastName: "C2", FirstName: "U", BirthDate: "2000-01-02", Gender: "F"})

	regRepo.Create(ctx, &models.RegistrationInput{CitizenID: c1, RegistrationType: "permanent", Settlement: "City1", Street: "Street1", HouseNumber: "1", RegistrationDate: "2023-01-10"})
	regRepo.Create(ctx, &models.RegistrationInput{CitizenID: c2, RegistrationType: "permanent", Settlement: "City2", Street: "Street2", HouseNumber: "2", RegistrationDate: "2023-02-10"})

	// Test all time
	results, err := repo.ExportRegisteredCitizens(ctx, "", "")
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Test range
	results, err = repo.ExportRegisteredCitizens(ctx, "2023-01-01", "2023-01-31")
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for Jan 2023, got %d", len(results))
	}
	if results[0].LastName != "C1" {
		t.Errorf("Expected C1, got %s", results[0].LastName)
	}

	// Test no results
	results, err = repo.ExportRegisteredCitizens(ctx, "2024-01-01", "2024-01-31")
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestReportRepository_GetStats(t *testing.T) {
	db := setupTestDB(t)
	citizenRepo := NewCitizenRepository(db)
	regRepo := NewRegistrationRepository(db)
	repo := NewReportRepository(db)
	ctx := context.Background()

	c1, _ := citizenRepo.Create(ctx, &models.CitizenInput{LastName: "C1", FirstName: "U", BirthDate: "2000-01-01", Gender: "M"})
	c2, _ := citizenRepo.Create(ctx, &models.CitizenInput{LastName: "C2", FirstName: "U", BirthDate: "2000-01-02", Gender: "F"})

	regRepo.Create(ctx, &models.RegistrationInput{CitizenID: c1, RegistrationType: "permanent", RegistrationDate: "2023-01-01"})
	regRepo.Create(ctx, &models.RegistrationInput{CitizenID: c2, RegistrationType: "permanent", RegistrationDate: "2023-01-02"})

	stats, err := repo.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.TotalCitizens != 2 {
		t.Errorf("Expected 2 citizens, got %d", stats.TotalCitizens)
	}
	if stats.TotalRegistrations != 2 {
		t.Errorf("Expected 2 registrations, got %d", stats.TotalRegistrations)
	}
	if stats.ActiveRegistrations != 2 {
		t.Errorf("Expected 2 active registrations, got %d", stats.ActiveRegistrations)
	}
}
