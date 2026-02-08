package services

import (
	"context"
	"errors"
	"passport-desk-mvp/internal/models"
	"passport-desk-mvp/internal/repository"
)

type AuditLogService struct {
	auditLogRepo   repository.AuditLogRepositoryInterface
	sessionService *SessionService
}

// NewAuditLogService creates a new audit log service
func NewAuditLogService(auditLogRepo repository.AuditLogRepositoryInterface, sessionService *SessionService) *AuditLogService {
	return &AuditLogService{auditLogRepo: auditLogRepo, sessionService: sessionService}
}

func (s *AuditLogService) LogAudit(ctx context.Context, log *models.AuditLog) error {
	if !s.sessionService.IsAuthenticated() {
		return errors.New("not authenticated")
	}
	s.sessionService.UpdateActivity()

	if log.OperatorID == 0 {
		s.sessionService.Mu.RLock()
		if s.sessionService.CurrentOperator != nil {
			log.OperatorID = s.sessionService.CurrentOperator.ID
		}
		s.sessionService.Mu.RUnlock()
	}

	return s.auditLogRepo.LogAudit(ctx, log)
}

// GetAuditLogs returns recent audit logs
func (s *AuditLogService) GetAuditLogs(ctx context.Context, limit int) ([]models.AuditLogOutput, error) {
	if !s.sessionService.IsAuthenticated() {
		return nil, errors.New("unauthorized")
	}
	s.sessionService.UpdateActivity()
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	return s.auditLogRepo.GetAuditLogs(ctx, limit)
}
