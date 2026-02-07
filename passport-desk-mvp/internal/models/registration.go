package models

import "time"

// Registration represents a citizen's address registration
type Registration struct {
	ID                 int64     `json:"id"`
	CitizenID          int64     `json:"citizen_id"`
	RegistrationType   string    `json:"registration_type"`  // permanent or temporary
	Region             string    `json:"region"`             // Oblast
	District           string    `json:"district,omitempty"` // Raion
	Settlement         string    `json:"settlement"`         // City/Village
	Street             string    `json:"street"`
	HouseNumber        string    `json:"house_number"`
	ApartmentNumber    string    `json:"apartment_number,omitempty"`
	RegistrationDate   string    `json:"registration_date"` // YYYY-MM-DD
	DeregistrationDate string    `json:"deregistration_date,omitempty"`
	BasisDocument      string    `json:"basis_document,omitempty"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// RegistrationInput represents input for registration creation
type RegistrationInput struct {
	CitizenID        int64  `json:"citizen_id"`
	RegistrationType string `json:"registration_type"`
	Region           string `json:"region"`
	District         string `json:"district"`
	Settlement       string `json:"settlement"`
	Street           string `json:"street"`
	HouseNumber      string `json:"house_number"`
	ApartmentNumber  string `json:"apartment_number"`
	RegistrationDate string `json:"registration_date"`
	BasisDocument    string `json:"basis_document"`
}

// RegistrationOutput represents output for registration
type RegistrationOutput struct {
	ID                 int64     `json:"id"`
	CitizenID          int64     `json:"citizen_id"`
	RegistrationType   string    `json:"registration_type"`
	Region             string    `json:"region"`
	District           string    `json:"district,omitempty"`
	Settlement         string    `json:"settlement"`
	Street             string    `json:"street"`
	HouseNumber        string    `json:"house_number"`
	ApartmentNumber    string    `json:"apartment_number,omitempty"`
	RegistrationDate   string    `json:"registration_date"`
	DeregistrationDate string    `json:"deregistration_date,omitempty"`
	BasisDocument      string    `json:"basis_document,omitempty"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type RegistrationListItem struct {
	RegistrationOutput
	CitizenName  string `json:"citizen_name"`
	CitizenBirth string `json:"citizen_birth"`
	CitizenPhone string `json:"citizen_phone"`
	CitizenTax   string `json:"citizen_tax"`
}

type RegistrationListResult struct {
	Items      []RegistrationListItem `json:"items"`
	Total      int                    `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
}
