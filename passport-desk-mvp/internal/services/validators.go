package services

import (
	"errors"
	"passport-desk-mvp/internal/models"
	"regexp"
	"time"
)

type CitizenValidator struct{}

func NewCitizenValidator() *CitizenValidator {
	return &CitizenValidator{}
}

var (
	emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	taxRegex   = regexp.MustCompile(`^\d{10}$`)
)

func (s *CitizenValidator) Validate(input *models.CitizenInput) error {
	if input.LastName == "" {
		return errors.New("прізвище є обов'язковим")
	}
	if input.FirstName == "" {
		return errors.New("ім'я є обов'язковим")
	}

	// Birth date validation
	if input.BirthDate == "" {
		return errors.New("дата народження є обов'язковою")
	}
	birthDate, err := time.Parse("2006-01-02", input.BirthDate)
	if err != nil {
		return errors.New("невірний формат дати народження (очікується РРРР-ММ-ДД)")
	}
	if birthDate.After(time.Now()) {
		return errors.New("дата народження не може бути в майбутньому")
	}

	// Gender validation
	if input.Gender != "M" && input.Gender != "F" {
		return errors.New("невірно вказана стать")
	}

	// Passport validation
	if input.PassportType == "old" {
		if len(input.PassportSeries) != 2 {
			return errors.New("серія паспорта старого зразка має містити 2 літери")
		}
		if len(input.PassportNumber) != 6 {
			return errors.New("номер паспорта старого зразка має містити 6 цифр")
		}
	} else if input.PassportType == "new" {
		if len(input.PassportNumber) != 9 {
			return errors.New("номер ID-картки має містити 9 цифр")
		}
	} else if input.PassportType == "" {
		return errors.New("тип паспорта є обов'язковим")
	}

	// Tax number validation (optional but must be valid if provided)
	if input.TaxNumber != "" && !taxRegex.MatchString(input.TaxNumber) {
		return errors.New("ІПН має містити 10 цифр")
	}

	// Email validation (optional)
	if input.Email != "" && !emailRegex.MatchString(input.Email) {
		return errors.New("невірний формат email")
	}

	return nil
}
