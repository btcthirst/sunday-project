package repository

import (
	"context"
	"passport-desk-mvp/internal/models"
)

type AuditLogRepositoryInterface interface {
	LogAudit(ctx context.Context, log *models.AuditLog) error
	GetAuditLogs(ctx context.Context, limit int) ([]models.AuditLogOutput, error)
}

type CitizenRepositoryInterface interface {
	Create(ctx context.Context, citizen *models.CitizenInput) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.CitizenOutput, error)
	Update(ctx context.Context, id int64, citizen *models.CitizenInput) error
	SoftDelete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int, includeDeleted bool) ([]*models.CitizenOutput, int, error)
	SearchByName(ctx context.Context, searchTerm string, limit int) ([]*models.CitizenOutput, error)
	SearchByBirthDate(ctx context.Context, birthDate string, limit int) ([]*models.CitizenOutput, error)
	Exists(ctx context.Context, id int64) (bool, error)
	GetByPassport(ctx context.Context, passportSeries, passportNumber string) (*models.CitizenOutput, error)
	GetAll(ctx context.Context) ([]*models.CitizenOutput, error)
	AddFamilyMember(ctx context.Context, citizenID, memberID int64, relationType string) error
	RemoveFamilyMember(ctx context.Context, citizenID, memberID int64) error
	GetFamilyMembers(ctx context.Context, citizenID int64) ([]*models.FamilyMemberOutput, error)
}

type OperatorRepositoryInterface interface {
	Create(ctx context.Context, op *models.Operator) error
	GetByUsername(ctx context.Context, username string) (*models.Operator, error)
	Exists(ctx context.Context) (bool, error)
}

type RegistrationRepositoryInterface interface {
	Create(ctx context.Context, input *models.RegistrationInput) (*models.RegistrationOutput, error)
	GetByID(ctx context.Context, id int64) (*models.RegistrationOutput, error)
	GetByCitizenID(ctx context.Context, citizenID int64) ([]models.RegistrationOutput, error)
	Deregister(ctx context.Context, id int64, deregistrationDate string) error
	ListAll(ctx context.Context, search string, isActive *bool, page, limit int) (*models.RegistrationListResult, error)
}

type ReportRepositoryInterface interface {
	ExportRegisteredCitizens(ctx context.Context, from, to string) ([]models.CitizenOutput, error)
	GetStats(ctx context.Context) (*models.StatsOutput, error)
}
