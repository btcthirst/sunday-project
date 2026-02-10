package services

import (
	"context"
	"errors"
	"passport-desk-mvp/internal/logger"
	"passport-desk-mvp/internal/models"
	"passport-desk-mvp/internal/repository"
)

type OperatorService struct {
	sessionService *SessionService
	operatorRepo   repository.OperatorRepositoryInterface
}

// NewOperatorService creates a new OperatorService
func NewOperatorService(sessionService *SessionService, operatorRepo repository.OperatorRepositoryInterface) *OperatorService {
	return &OperatorService{
		sessionService: sessionService,
		operatorRepo:   operatorRepo,
	}
}

// GetCurrentOperator returns the current operator info
func (a *OperatorService) GetCurrentOperator(ctx context.Context) *models.OperatorInfo {
	logger.Info("Getting current operator")
	a.sessionService.Mu.RLock()
	defer a.sessionService.Mu.RUnlock()

	if a.sessionService.CurrentOperator == nil {
		logger.Error("Not logged in")
		return nil
	}

	return &models.OperatorInfo{
		ID:       a.sessionService.CurrentOperator.ID,
		Username: a.sessionService.CurrentOperator.Username,
		FullName: a.sessionService.CurrentOperator.FullName,
	}
}

func (a *OperatorService) CreateOperator(ctx context.Context, op *models.Operator) error {
	return a.operatorRepo.Create(ctx, op)
}

func (a *OperatorService) GetOperatorByUsername(ctx context.Context, username string) (*models.Operator, error) {
	return a.operatorRepo.GetByUsername(ctx, username)
}

func (a *OperatorService) OperatorExists(ctx context.Context) (bool, error) {
	return a.operatorRepo.Exists(ctx)
}

// UpdateProfile updates the current operator's full name
func (a *OperatorService) UpdateProfile(ctx context.Context, fullName string) error {
	a.sessionService.Mu.Lock()
	defer a.sessionService.Mu.Unlock()

	if a.sessionService.CurrentOperator == nil {
		return errors.New("not logged in")
	}

	operator := a.sessionService.CurrentOperator
	operator.FullName = fullName

	if err := a.operatorRepo.Update(ctx, operator); err != nil {
		return err
	}

	return nil
}

func (a *OperatorService) UpdatePassword(ctx context.Context, op *models.Operator) error {
	return a.operatorRepo.Update(ctx, op)
}
