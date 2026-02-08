package services

import (
	"context"
	"passport-desk-mvp/internal/models"
)

// MockCitizenRepo implements repository.CitizenRepositoryInterface
type MockCitizenRepo struct {
	CreateFunc             func(ctx context.Context, citizen *models.CitizenInput) (int64, error)
	GetByIDFunc            func(ctx context.Context, id int64) (*models.CitizenOutput, error)
	UpdateFunc             func(ctx context.Context, id int64, citizen *models.CitizenInput) error
	SoftDeleteFunc         func(ctx context.Context, id int64) error
	RestoreFunc            func(ctx context.Context, id int64) error
	ListFunc               func(ctx context.Context, offset, limit int, includeDeleted bool) ([]*models.CitizenOutput, int, error)
	SearchByNameFunc       func(ctx context.Context, searchTerm string, limit int) ([]*models.CitizenOutput, error)
	SearchByBirthDateFunc  func(ctx context.Context, birthDate string, limit int) ([]*models.CitizenOutput, error)
	ExistsFunc             func(ctx context.Context, id int64) (bool, error)
	GetByPassportFunc      func(ctx context.Context, passportSeries, passportNumber string) (*models.CitizenOutput, error)
	GetAllFunc             func(ctx context.Context) ([]*models.CitizenOutput, error)
	AddFamilyMemberFunc    func(ctx context.Context, citizenID, memberID int64, relationType string) error
	RemoveFamilyMemberFunc func(ctx context.Context, citizenID, memberID int64) error
	GetFamilyMembersFunc   func(ctx context.Context, citizenID int64) ([]*models.FamilyMemberOutput, error)
}

func (m *MockCitizenRepo) Create(ctx context.Context, citizen *models.CitizenInput) (int64, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, citizen)
	}
	return 1, nil
}
func (m *MockCitizenRepo) GetByID(ctx context.Context, id int64) (*models.CitizenOutput, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return &models.CitizenOutput{ID: id}, nil
}
func (m *MockCitizenRepo) Update(ctx context.Context, id int64, citizen *models.CitizenInput) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, citizen)
	}
	return nil
}
func (m *MockCitizenRepo) SoftDelete(ctx context.Context, id int64) error {
	if m.SoftDeleteFunc != nil {
		return m.SoftDeleteFunc(ctx, id)
	}
	return nil
}
func (m *MockCitizenRepo) Restore(ctx context.Context, id int64) error {
	if m.RestoreFunc != nil {
		return m.RestoreFunc(ctx, id)
	}
	return nil
}
func (m *MockCitizenRepo) List(ctx context.Context, offset, limit int, includeDeleted bool) ([]*models.CitizenOutput, int, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, offset, limit, includeDeleted)
	}
	return []*models.CitizenOutput{}, 0, nil
}
func (m *MockCitizenRepo) SearchByName(ctx context.Context, searchTerm string, limit int) ([]*models.CitizenOutput, error) {
	if m.SearchByNameFunc != nil {
		return m.SearchByNameFunc(ctx, searchTerm, limit)
	}
	return []*models.CitizenOutput{}, nil
}
func (m *MockCitizenRepo) SearchByBirthDate(ctx context.Context, birthDate string, limit int) ([]*models.CitizenOutput, error) {
	if m.SearchByBirthDateFunc != nil {
		return m.SearchByBirthDateFunc(ctx, birthDate, limit)
	}
	return []*models.CitizenOutput{}, nil
}
func (m *MockCitizenRepo) Exists(ctx context.Context, id int64) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, id)
	}
	return true, nil
}
func (m *MockCitizenRepo) GetByPassport(ctx context.Context, passportSeries, passportNumber string) (*models.CitizenOutput, error) {
	if m.GetByPassportFunc != nil {
		return m.GetByPassportFunc(ctx, passportSeries, passportNumber)
	}
	return nil, nil // Not found by default
}
func (m *MockCitizenRepo) GetAll(ctx context.Context) ([]*models.CitizenOutput, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return []*models.CitizenOutput{}, nil
}
func (m *MockCitizenRepo) AddFamilyMember(ctx context.Context, citizenID, memberID int64, relationType string) error {
	if m.AddFamilyMemberFunc != nil {
		return m.AddFamilyMemberFunc(ctx, citizenID, memberID, relationType)
	}
	return nil
}
func (m *MockCitizenRepo) RemoveFamilyMember(ctx context.Context, citizenID, memberID int64) error {
	if m.RemoveFamilyMemberFunc != nil {
		return m.RemoveFamilyMemberFunc(ctx, citizenID, memberID)
	}
	return nil
}
func (m *MockCitizenRepo) GetFamilyMembers(ctx context.Context, citizenID int64) ([]*models.FamilyMemberOutput, error) {
	if m.GetFamilyMembersFunc != nil {
		return m.GetFamilyMembersFunc(ctx, citizenID)
	}
	return []*models.FamilyMemberOutput{}, nil
}

// MockAuditLogRepo implements repository.AuditLogRepositoryInterface
type MockAuditLogRepo struct {
	LogAuditFunc     func(ctx context.Context, log *models.AuditLog) error
	GetAuditLogsFunc func(ctx context.Context, limit int) ([]models.AuditLogOutput, error)
}

func (m *MockAuditLogRepo) LogAudit(ctx context.Context, log *models.AuditLog) error {
	if m.LogAuditFunc != nil {
		return m.LogAuditFunc(ctx, log)
	}
	return nil
}
func (m *MockAuditLogRepo) GetAuditLogs(ctx context.Context, limit int) ([]models.AuditLogOutput, error) {
	if m.GetAuditLogsFunc != nil {
		return m.GetAuditLogsFunc(ctx, limit)
	}
	return []models.AuditLogOutput{}, nil
}

// MockRegistrationRepo implements repository.RegistrationRepositoryInterface
type MockRegistrationRepo struct {
	CreateFunc         func(ctx context.Context, input *models.RegistrationInput) (*models.RegistrationOutput, error)
	GetByIDFunc        func(ctx context.Context, id int64) (*models.RegistrationOutput, error)
	GetByCitizenIDFunc func(ctx context.Context, citizenID int64) ([]models.RegistrationOutput, error)
	DeregisterFunc     func(ctx context.Context, id int64, deregistrationDate string) error
	ListAllFunc        func(ctx context.Context, search string, isActive *bool, page, limit int) (*models.RegistrationListResult, error)
}

func (m *MockRegistrationRepo) Create(ctx context.Context, input *models.RegistrationInput) (*models.RegistrationOutput, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}
	return &models.RegistrationOutput{ID: 1}, nil
}
func (m *MockRegistrationRepo) GetByID(ctx context.Context, id int64) (*models.RegistrationOutput, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return &models.RegistrationOutput{ID: id}, nil
}
func (m *MockRegistrationRepo) GetByCitizenID(ctx context.Context, citizenID int64) ([]models.RegistrationOutput, error) {
	if m.GetByCitizenIDFunc != nil {
		return m.GetByCitizenIDFunc(ctx, citizenID)
	}
	return []models.RegistrationOutput{}, nil
}
func (m *MockRegistrationRepo) Deregister(ctx context.Context, id int64, deregistrationDate string) error {
	if m.DeregisterFunc != nil {
		return m.DeregisterFunc(ctx, id, deregistrationDate)
	}
	return nil
}
func (m *MockRegistrationRepo) ListAll(ctx context.Context, search string, isActive *bool, page, limit int) (*models.RegistrationListResult, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx, search, isActive, page, limit)
	}
	return &models.RegistrationListResult{}, nil
}
