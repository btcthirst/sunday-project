package container

import (
	"passport-desk-mvp/internal/database"
	"passport-desk-mvp/internal/repository"
	"passport-desk-mvp/internal/security"
	"passport-desk-mvp/internal/services"
)

type Container struct {
	DB       *database.Database
	Keystore *security.Keystore
	Crypto   *security.Crypto

	// Repositories
	// Repositories
	AuditLogRepo     repository.AuditLogRepositoryInterface
	CitizenRepo      repository.CitizenRepositoryInterface
	RegistrationRepo repository.RegistrationRepositoryInterface
	ReportRepo       repository.ReportRepositoryInterface
	OperatorRepo     repository.OperatorRepositoryInterface

	// Services
	AuditLogService     *services.AuditLogService
	CitizenService      *services.CitizenService
	RegistrationService *services.RegistrationService
	ReportService       *services.ReportService
	BackupService       *services.BackupService
	ImportExportService *services.ImportExportService
	SessionService      *services.SessionService
	OperatorService     *services.OperatorService
}

func NewContainer(dataDir string) *Container {
	keystore := security.NewKeystore(dataDir)
	sessionService := services.NewSessionService(nil)
	return &Container{
		Keystore:       keystore,
		BackupService:  services.NewBackupService(dataDir),
		SessionService: sessionService,
	}
}

func (c *Container) InitializeWithKey(key []byte, dbPath string) error {
	db, err := database.New(dbPath, security.KeyToHex(key))
	if err != nil {
		return err
	}
	c.DB = db
	c.Crypto = security.NewCrypto(key)

	c.AuditLogRepo = repository.NewAuditLogRepository(db)
	c.CitizenRepo = repository.NewCitizenRepository(db)
	c.RegistrationRepo = repository.NewRegistrationRepository(db)
	c.ReportRepo = repository.NewReportRepository(db)
	c.OperatorRepo = repository.NewOperatorRepository(db)

	c.AuditLogService = services.NewAuditLogService(c.AuditLogRepo, c.SessionService)
	c.CitizenService = services.NewCitizenService(c.Crypto, c.CitizenRepo, c.AuditLogService)
	c.RegistrationService = services.NewRegistrationService(c.RegistrationRepo, c.Crypto, c.AuditLogService)
	c.ReportService = services.NewReportService(c.ReportRepo, c.CitizenService, c.RegistrationService, c.AuditLogService)
	c.ImportExportService = services.NewImportExportService(c.CitizenService, c.AuditLogService)
	c.OperatorService = services.NewOperatorService(c.SessionService, c.OperatorRepo)
	c.SessionService.EncryptionKey = key
	return nil
}
