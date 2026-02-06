package models

import "time"

// Citizen represents a citizen record
type Citizen struct {
	ID             int64     `json:"id"`
	LastName       string    `json:"last_name"`
	FirstName      string    `json:"first_name"`
	MiddleName     string    `json:"middle_name,omitempty"`
	BirthDate      string    `json:"birth_date"`           // YYYY-MM-DD
	PassportSeries string    `json:"passport_series"`      // Encrypted
	PassportNumber string    `json:"passport_number"`      // Encrypted
	PassportType   string    `json:"passport_type"`        // 'old' or 'new'
	TaxNumber      string    `json:"tax_number,omitempty"` // IPN, Encrypted
	Gender         string    `json:"gender"`               // M or F
	BirthPlace     string    `json:"birth_place,omitempty"`
	Phone          string    `json:"phone,omitempty"` // Encrypted
	Email          string    `json:"email,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	Deleted        bool      `json:"deleted"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ActiveAddress  string    `json:"active_address,omitempty"` // Computed field
}

// FamilyRelation represents a relationship between two citizens
type FamilyRelation struct {
	ID           int64     `json:"id"`
	CitizenID    int64     `json:"citizen_id"`
	MemberID     int64     `json:"member_id"`
	RelationType string    `json:"relation_type"` // spouse, child, parent, etc.
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
