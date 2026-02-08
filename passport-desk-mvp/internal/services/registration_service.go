package services

import (
	"context"
	"fmt"
	"log/slog"
	"passport-desk-mvp/internal/logger"
	"passport-desk-mvp/internal/models"
	"passport-desk-mvp/internal/repository"
	"passport-desk-mvp/internal/security"
)

type RegistrationService struct {
	repo     repository.RegistrationRepositoryInterface
	crypto   *security.Crypto
	auditLog *AuditLogService
}

func NewRegistrationService(repo repository.RegistrationRepositoryInterface, crypto *security.Crypto, auditLog *AuditLogService) *RegistrationService {
	return &RegistrationService{repo: repo, crypto: crypto, auditLog: auditLog}
}

func (s *RegistrationService) Create(ctx context.Context, input *models.RegistrationInput) (*models.RegistrationOutput, error) {
	log := logger.FromContext(ctx).With(
		slog.String("service", "registration"),
		slog.String("method", "Create"),
	)

	// SLOG: Tech log start operation
	log.Info("Creating registration",
		slog.Int64("citizen_id", input.CitizenID),
		slog.String("type", input.RegistrationType),
	)

	reg, err := s.repo.Create(ctx, input)
	if err != nil {
		log.Error("Create failed", slog.String("error", err.Error()))
		return nil, err
	}

	log.Info("Registration created successfully",
		slog.Int64("id", reg.ID),
	)

	// AUDIT LOG: Business event
	if err := s.auditLog.LogAudit(ctx, &models.AuditLog{
		OperatorID:  getOperatorIDFromContext(ctx),
		ActionType:  "CREATE",
		TableName:   "registrations",
		RecordID:    reg.ID,
		Description: fmt.Sprintf("Created registration for citizen %d", reg.CitizenID),
	}); err != nil {
		log.Warn("Audit log failed", slog.String("error", err.Error()))
	}

	return reg, nil
}

func (s *RegistrationService) GetByID(ctx context.Context, id int64) (*models.RegistrationOutput, error) {
	log := logger.FromContext(ctx).With(
		slog.String("service", "registration"),
		slog.String("method", "GetByID"),
		slog.Int64("id", id),
	)

	reg, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("GetByID failed", slog.String("error", err.Error()))
		return nil, err
	}
	return reg, nil
}

func (s *RegistrationService) GetByCitizenID(ctx context.Context, citizenID int64) ([]models.RegistrationOutput, error) {
	log := logger.FromContext(ctx).With(
		slog.String("service", "registration"),
		slog.String("method", "GetByCitizenID"),
		slog.Int64("citizen_id", citizenID),
	)

	regs, err := s.repo.GetByCitizenID(ctx, citizenID)
	if err != nil {
		log.Error("GetByCitizenID failed", slog.String("error", err.Error()))
		return nil, err
	}
	return regs, nil
}

func (s *RegistrationService) Deregister(ctx context.Context, id int64, deregistrationDate string) error {
	log := logger.FromContext(ctx).With(
		slog.String("service", "registration"),
		slog.String("method", "Deregister"),
		slog.Int64("id", id),
	)

	log.Info("Deregistering citizen")

	err := s.repo.Deregister(ctx, id, deregistrationDate)
	if err != nil {
		log.Error("Deregister failed", slog.String("error", err.Error()))
		return err
	}

	log.Info("Deregistration successful")

	// AUDIT LOG: Business event
	if err := s.auditLog.LogAudit(ctx, &models.AuditLog{
		OperatorID:  getOperatorIDFromContext(ctx),
		ActionType:  "UPDATE",
		TableName:   "registrations",
		RecordID:    id,
		Description: "Deregistered citizen",
	}); err != nil {
		log.Warn("Audit log failed", slog.String("error", err.Error()))
	}

	return nil
}

func (s *RegistrationService) ListAll(ctx context.Context, search string, isActive *bool, page, limit int) (*models.RegistrationListResult, error) {
	log := logger.FromContext(ctx).With(
		slog.String("service", "registration"),
		slog.String("method", "ListAll"),
	)

	result, err := s.repo.ListAll(ctx, search, isActive, page, limit)
	if err != nil {
		log.Error("ListAll failed", slog.String("error", err.Error()))
		return nil, err
	}

	for i := range result.Items {
		if d, err := s.crypto.Decrypt(result.Items[i].CitizenPhone); err == nil {
			result.Items[i].CitizenPhone = d
		} else {
			log.Warn("Failed to decrypt citizen phone", slog.Int64("id", result.Items[i].ID), slog.String("error", err.Error()))
		}

		if d, err := s.crypto.Decrypt(result.Items[i].CitizenTax); err == nil {
			result.Items[i].CitizenTax = d
		} else {
			log.Warn("Failed to decrypt citizen tax", slog.Int64("id", result.Items[i].ID), slog.String("error", err.Error()))
		}
	}

	return result, nil
}
