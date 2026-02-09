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

func TestCitizenRepository_FamilyRelations(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)
	ctx := context.Background()

	// 1. Create two citizens
	id1, _ := repo.Create(ctx, &models.CitizenInput{LastName: "Father", FirstName: "John", Gender: "M", BirthDate: "1970-01-01", PassportSeries: "AA", PassportNumber: "1"})
	id2, _ := repo.Create(ctx, &models.CitizenInput{LastName: "Son", FirstName: "Bob", Gender: "M", BirthDate: "2000-01-01", PassportSeries: "AA", PassportNumber: "2"})

	// 2. Add relationship: John is Bob's Father
	err := repo.AddFamilyMember(ctx, id1, id2, "Батько")
	if err != nil {
		t.Fatalf("AddFamilyMember failed: %v", err)
	}

	// 3. Verify reciprocity: Bob should have John as Father
	members1, _ := repo.GetFamilyMembers(ctx, id1)
	if len(members1) != 1 || members1[0].ID != id2 || members1[0].RelationType != "Батько" {
		t.Errorf("Expected John to have Bob as son, got %v", members1)
	}

	members2, _ := repo.GetFamilyMembers(ctx, id2)
	if len(members2) != 1 || members2[0].ID != id1 || members2[0].RelationType != "Син" {
		t.Errorf("Expected Bob to have John as father, got %v", members2)
	}
}

func TestCitizenRepository_Reciprocity(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)
	ctx := context.Background()

	id1, _ := repo.Create(ctx, &models.CitizenInput{LastName: "A", FirstName: "A", Gender: "M", BirthDate: "1970-01-01", PassportSeries: "AA", PassportNumber: "1"})
	id2, _ := repo.Create(ctx, &models.CitizenInput{LastName: "B", FirstName: "B", Gender: "M", BirthDate: "2000-01-01", PassportSeries: "AA", PassportNumber: "2"})

	// A (Male) adds B as "Син"
	repo.AddFamilyMember(ctx, id1, id2, "Син")

	// Get A's family: should have B as "Син"
	f1, _ := repo.GetFamilyMembers(ctx, id1)
	if len(f1) != 1 || f1[0].RelationType != "Син" {
		t.Errorf("Expected A to have B as 'Син', got %s", f1[0].RelationType)
	}

	// Get B's family: should have A as "Батько" (inverse of A being male and B being his son)
	f2, _ := repo.GetFamilyMembers(ctx, id2)
	if len(f2) != 1 || f2[0].RelationType != "Батько" {
		t.Errorf("Expected B to have A as 'Батько', got %s", f2[0].RelationType)
	}
}

func TestCitizenRepository_UpdateWithRelations(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)
	ctx := context.Background()

	id1, _ := repo.Create(ctx, &models.CitizenInput{LastName: "A", FirstName: "A", Gender: "M", BirthDate: "1970-01-01", PassportSeries: "AA", PassportNumber: "1"})
	id2, _ := repo.Create(ctx, &models.CitizenInput{LastName: "B", FirstName: "B", Gender: "F", BirthDate: "1975-01-01", PassportSeries: "AA", PassportNumber: "2"})

	// Update A to add B as "Дружина"
	input := &models.CitizenInput{
		LastName: "A", FirstName: "A", Gender: "M", BirthDate: "1970-01-01",
		PassportSeries: "AA", PassportNumber: "1",
		FamilyRelations: []models.FamilyRelationInput{
			{MemberID: id2, RelationType: "Дружина"},
		},
	}
	err := repo.Update(ctx, id1, input)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify A has B as wife
	f1, _ := repo.GetFamilyMembers(ctx, id1)
	if len(f1) != 1 || f1[0].RelationType != "Дружина" {
		t.Fatalf("Expected A to have B as wife, got %v", f1)
	}

	// Verify B has A as husband
	f2, _ := repo.GetFamilyMembers(ctx, id2)
	if len(f2) != 1 || f2[0].RelationType != "Чоловік" {
		t.Fatalf("Expected B to have A as husband, got %v", f2)
	}
}

func TestCitizenRepository_RemoveFamilyMember(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)
	ctx := context.Background()

	id1, _ := repo.Create(ctx, &models.CitizenInput{LastName: "A", FirstName: "A", Gender: "M", BirthDate: "1970-01-01", PassportSeries: "AA", PassportNumber: "1"})
	id2, _ := repo.Create(ctx, &models.CitizenInput{LastName: "B", FirstName: "B", Gender: "F", BirthDate: "1975-01-01", PassportSeries: "AA", PassportNumber: "2"})

	_ = repo.AddFamilyMember(ctx, id1, id2, "Дружина")

	// Verify they are linked
	f1, _ := repo.GetFamilyMembers(ctx, id1)
	if len(f1) != 1 {
		t.Fatalf("Expected 1 family member, got %d", len(f1))
	}

	// Remove relationship
	err := repo.RemoveFamilyMember(ctx, id1, id2)
	if err != nil {
		t.Fatalf("RemoveFamilyMember failed: %v", err)
	}

	// Verify they are unlinked
	f1, _ = repo.GetFamilyMembers(ctx, id1)
	if len(f1) != 0 {
		t.Errorf("Expected 0 family members for A, got %d", len(f1))
	}

	f2, _ := repo.GetFamilyMembers(ctx, id2)
	if len(f2) != 0 {
		t.Errorf("Expected 0 family members for B, got %d", len(f2))
	}
}

func TestCitizenRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)
	ctx := context.Background()

	id, _ := repo.Create(ctx, &models.CitizenInput{LastName: "A", FirstName: "A", Gender: "M", BirthDate: "1970-01-01", PassportSeries: "AA", PassportNumber: "1"})

	exists, err := repo.Exists(ctx, id)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected citizen to exist")
	}

	exists, err = repo.Exists(ctx, 999)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Expected citizen to not exist")
	}
}

func TestCitizenRepository_GetByPassport(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)
	ctx := context.Background()

	id, _ := repo.Create(ctx, &models.CitizenInput{LastName: "A", FirstName: "A", Gender: "M", BirthDate: "1970-01-01", PassportSeries: "AA", PassportNumber: "123"})

	// GetByPassport currently returns the FIRST citizen in the results (as noted in implementation)
	// This will change once hash indexing is implemented.
	citizen, err := repo.GetByPassport(ctx, "AA", "123")
	if err != nil {
		t.Fatalf("GetByPassport failed: %v", err)
	}
	if citizen.ID != id {
		t.Errorf("Expected citizen ID %d, got %d", id, citizen.ID)
	}
}

func TestCitizenRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCitizenRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &models.CitizenInput{LastName: "B", FirstName: "B", Gender: "M", BirthDate: "1970-01-01", PassportSeries: "AA", PassportNumber: "1"})
	repo.Create(ctx, &models.CitizenInput{LastName: "A", FirstName: "A", Gender: "F", BirthDate: "1975-01-01", PassportSeries: "AA", PassportNumber: "2"})

	citizens, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(citizens) != 2 {
		t.Errorf("Expected 2 citizens, got %d", len(citizens))
	}

	// Should be ordered by last name
	if citizens[0].LastName != "A" {
		t.Errorf("Expected first citizen to be A, got %s", citizens[0].LastName)
	}
}
