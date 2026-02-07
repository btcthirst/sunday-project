package services

import "passport-desk-mvp/internal/models"

type CitizenValidator struct{}

func NewCitizenValidator() *CitizenValidator {
	return &CitizenValidator{}
}

func (s *CitizenValidator) Validate(input *models.CitizenInput) error {
	return nil
}
