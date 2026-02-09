package repository

import (
	"context"
	"passport-desk-mvp/internal/models"
	"testing"
)

func TestRegistrationRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	// Need citizen first
	citizenRepo := NewCitizenRepository(db)
	regRepo := NewRegistrationRepository(db)
	ctx := context.Background()

	citizenInput := &models.CitizenInput{
		LastName:  "RegTest",
		FirstName: "User",
		BirthDate: "2000-01-01",
		Gender:    "M",
	}
	citizenID, err := citizenRepo.Create(ctx, citizenInput)
	if err != nil {
		t.Fatalf("Failed to create citizen: %v", err)
	}

	regInput := &models.RegistrationInput{
		CitizenID:        citizenID,
		RegistrationType: "permanent",
		Region:           "Test Region",
		Settlement:       "Test City",
		Street:           "Main St",
		HouseNumber:      "1",
		RegistrationDate: "2023-01-01",
	}

	reg, err := regRepo.Create(ctx, regInput)
	if err != nil {
		t.Fatalf("Failed to create registration: %v", err)
	}

	if reg.ID == 0 {
		t.Fatalf("Expected valid ID, got 0")
	}

	// Verify retrieval
	fetchedReg, err := regRepo.GetByID(ctx, reg.ID)
	if err != nil {
		t.Fatalf("Failed to get registration: %v", err)
	}
	if fetchedReg.Settlement != "Test City" {
		t.Errorf("Expected Test City, got %s", fetchedReg.Settlement)
	}
}

func TestRegistrationRepository_Deregister(t *testing.T) {
	db := setupTestDB(t)
	citizenRepo := NewCitizenRepository(db)
	regRepo := NewRegistrationRepository(db)
	ctx := context.Background()

	citizenID, _ := citizenRepo.Create(ctx, &models.CitizenInput{LastName: "Dereg", FirstName: "User", BirthDate: "2000-01-01", Gender: "M"})
	reg, _ := regRepo.Create(ctx, &models.RegistrationInput{CitizenID: citizenID, RegistrationType: "permanent", Region: "R", Settlement: "S", Street: "St", HouseNumber: "1", RegistrationDate: "2023-01-01"})

	deregDate := "2023-12-31"
	err := regRepo.Deregister(ctx, reg.ID, deregDate)
	if err != nil {
		t.Fatalf("Failed to deregister: %v", err)
	}

	fetchedReg, err := regRepo.GetByID(ctx, reg.ID)
	if err != nil {
		t.Fatalf("Failed to fetch: %v", err)
	}

	if fetchedReg.IsActive {
		t.Errorf("Expected registration to be inactive")
	}
	if fetchedReg.DeregistrationDate != deregDate {
		t.Errorf("Expected dereg date %s, got %s", deregDate, fetchedReg.DeregistrationDate)
	}
}

func TestRegistrationRepository_GetByCitizenID(t *testing.T) {
	db := setupTestDB(t)
	citizenRepo := NewCitizenRepository(db)
	regRepo := NewRegistrationRepository(db)
	ctx := context.Background()

	citizenID, _ := citizenRepo.Create(ctx, &models.CitizenInput{LastName: "Multi", FirstName: "Reg", BirthDate: "2000-01-01", Gender: "M"})

	// Create 2 registrations
	regRepo.Create(ctx, &models.RegistrationInput{CitizenID: citizenID, RegistrationType: "permanent", Region: "R1", Settlement: "S1", Street: "St1", HouseNumber: "1", RegistrationDate: "2020-01-01"})
	regRepo.Create(ctx, &models.RegistrationInput{CitizenID: citizenID, RegistrationType: "temporary", Region: "R2", Settlement: "S2", Street: "St2", HouseNumber: "2", RegistrationDate: "2021-01-01"})

	regs, err := regRepo.GetByCitizenID(ctx, citizenID)
	if err != nil {
		t.Fatalf("Failed to get regs: %v", err)
	}

	if len(regs) != 2 {
		t.Errorf("Expected 2 registrations, got %d", len(regs))
	}
}

func TestRegistrationRepository_ListAll(t *testing.T) {
	db := setupTestDB(t)
	citizenRepo := NewCitizenRepository(db)
	regRepo := NewRegistrationRepository(db)
	ctx := context.Background()

	c1, _ := citizenRepo.Create(ctx, &models.CitizenInput{LastName: "C1", FirstName: "U", BirthDate: "2000-01-01", Gender: "M"})
	c2, _ := citizenRepo.Create(ctx, &models.CitizenInput{LastName: "C2", FirstName: "U", BirthDate: "2000-01-02", Gender: "F"})

	regRepo.Create(ctx, &models.RegistrationInput{CitizenID: c1, RegistrationType: "permanent", Region: "R1", Settlement: "S1", Street: "St1", HouseNumber: "1", RegistrationDate: "2020-01-01"})
	regRepo.Create(ctx, &models.RegistrationInput{CitizenID: c2, RegistrationType: "temporary", Region: "R2", Settlement: "S2", Street: "St2", HouseNumber: "2", RegistrationDate: "2021-01-01"})

	// ListAll(ctx, search, isActive, page, limit)
	result, err := regRepo.ListAll(ctx, "", nil, 1, 10)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}

	if result.Total != 2 {
		t.Errorf("Expected total 2, got %d", result.Total)
	}
	if len(result.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result.Items))
	}
}
