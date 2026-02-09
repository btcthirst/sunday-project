package services

import (
	"passport-desk-mvp/internal/models"
	"strings"
	"testing"
)

func TestCitizenValidator_Validate(t *testing.T) {
	v := NewCitizenValidator()

	tests := []struct {
		name    string
		input   *models.CitizenInput
		wantErr bool
		msg     string
	}{
		// ... tests ...
		{
			name: "valid input old passport",
			input: &models.CitizenInput{
				LastName:       "LastName",
				FirstName:      "FirstName",
				BirthDate:      "1990-01-01",
				Gender:         "M",
				PassportType:   "old",
				PassportSeries: "AA",
				PassportNumber: "123456",
			},
			wantErr: false,
		},
		{
			name: "valid input new passport",
			input: &models.CitizenInput{
				LastName:       "LastName",
				FirstName:      "FirstName",
				BirthDate:      "1990-01-01",
				Gender:         "F",
				PassportType:   "new",
				PassportNumber: "123456789",
			},
			wantErr: false,
		},
		{
			name: "missing last name",
			input: &models.CitizenInput{
				FirstName: "FirstName",
			},
			wantErr: true,
			msg:     "прізвище є обов'язковим",
		},
		{
			name: "invalid birth date format",
			input: &models.CitizenInput{
				LastName:  "L",
				FirstName: "F",
				BirthDate: "01.01.1990",
			},
			wantErr: true,
			msg:     "невірний формат дати народження",
		},
		{
			name: "birth date in future",
			input: &models.CitizenInput{
				LastName:  "L",
				FirstName: "F",
				BirthDate: "2099-01-01",
			},
			wantErr: true,
			msg:     "в майбутньому",
		},
		{
			name: "invalid gender",
			input: &models.CitizenInput{
				LastName:  "L",
				FirstName: "F",
				BirthDate: "1990-01-01",
				Gender:    "X",
			},
			wantErr: true,
			msg:     "невірно вказана стать",
		},
		{
			name: "invalid old passport series",
			input: &models.CitizenInput{
				LastName:       "L",
				FirstName:      "F",
				BirthDate:      "1990-01-01",
				Gender:         "M",
				PassportType:   "old",
				PassportSeries: "A",
				PassportNumber: "123456",
			},
			wantErr: true,
			msg:     "серія паспорта старого зразка має містити 2 літери",
		},
		{
			name: "invalid old passport number",
			input: &models.CitizenInput{
				LastName:       "L",
				FirstName:      "F",
				BirthDate:      "1990-01-01",
				Gender:         "M",
				PassportType:   "old",
				PassportSeries: "AA",
				PassportNumber: "12345",
			},
			wantErr: true,
			msg:     "номер паспорта старого зразка має містити 6 цифр",
		},
		{
			name: "invalid new passport number",
			input: &models.CitizenInput{
				LastName:       "L",
				FirstName:      "F",
				BirthDate:      "1990-01-01",
				Gender:         "M",
				PassportType:   "new",
				PassportNumber: "12345678",
			},
			wantErr: true,
			msg:     "номер ID-картки має містити 9 цифр",
		},
		{
			name: "invalid tax number",
			input: &models.CitizenInput{
				LastName:       "L",
				FirstName:      "F",
				BirthDate:      "1990-01-01",
				Gender:         "M",
				PassportType:   "new",
				PassportNumber: "123456789",
				TaxNumber:      "123",
			},
			wantErr: true,
			msg:     "ІПН має містити 10 цифр",
		},
		{
			name: "invalid email",
			input: &models.CitizenInput{
				LastName:       "L",
				FirstName:      "F",
				BirthDate:      "1990-01-01",
				Gender:         "M",
				PassportType:   "new",
				PassportNumber: "123456789",
				Email:          "invalid",
			},
			wantErr: true,
			msg:     "невірний формат email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.msg != "" {
				if !strings.Contains(err.Error(), tt.msg) {
					t.Errorf("Validate() error message = %v, want to contain %v", err, tt.msg)
				}
			}
		})
	}
}
