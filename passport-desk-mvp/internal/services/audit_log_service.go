package services

import (
	"errors"
	"passport-desk-mvp/internal/logger"
	"passport-desk-mvp/internal/models"
	"passport-desk-mvp/internal/repository"
)

type AuditLogService struct {
	auditLogRepo   *repository.AuditLogRepository
	sessionService *SessionService
}

// GetAuditLogs returns recent audit logs
func (a *AuditLogService) GetAuditLogs(limit int) ([]models.AuditLogOutput, error) {
	if !a.sessionService.IsAuthenticated() {
		logger.Error("GetAuditLogs: Not authenticated")
		return nil, errors.New("unauthorized")
	}
	a.sessionService.UpdateActivity()
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	return a.auditLogRepo.GetAuditLogs(limit)
}
