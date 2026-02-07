package services

import (
	"context"
	"passport-desk-mvp/internal/models"
	"passport-desk-mvp/internal/repository"
	"passport-desk-mvp/internal/security"
)

type RegistrationService struct {
	repo   *repository.RegistrationRepository
	crypto *security.Crypto
}

func NewRegistrationService(repo *repository.RegistrationRepository, crypto *security.Crypto) *RegistrationService {
	return &RegistrationService{repo: repo, crypto: crypto}
}

func (s *RegistrationService) Create(ctx context.Context, input *models.RegistrationInput) (*models.RegistrationOutput, error) {
	return s.repo.Create(input)
}

func (s *RegistrationService) GetByID(ctx context.Context, id int64) (*models.RegistrationOutput, error) {
	return s.repo.GetByID(id)
}

func (s *RegistrationService) GetByCitizenID(ctx context.Context, citizenID int64) ([]models.RegistrationOutput, error) {
	return s.repo.GetByCitizenID(citizenID)
}

func (s *RegistrationService) Deregister(ctx context.Context, id int64, deregistrationDate string) error {
	return s.repo.Deregister(id, deregistrationDate)
}

func (s *RegistrationService) ListAll(ctx context.Context, search string, isActive *bool, page, limit int) (*models.RegistrationListResult, error) {
	result, err := s.repo.ListAll(search, isActive, page, limit)
	if err != nil {
		return nil, err
	}

	for _, item := range result.Items {
		item.CitizenPhone, _ = s.crypto.Decrypt(item.CitizenPhone)
		item.CitizenTax, _ = s.crypto.Decrypt(item.CitizenTax)
	}

	return result, nil
}
