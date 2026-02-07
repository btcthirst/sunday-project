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

// FamilyRelationInput represents input for a family relationship
type FamilyRelationInput struct {
	MemberID     int64  `json:"member_id"`
	RelationType string `json:"relation_type"`
}

// CitizenInput is the input structure for creating/updating citizens
type CitizenInput struct {
	LastName        string                `json:"last_name"`
	FirstName       string                `json:"first_name"`
	MiddleName      string                `json:"middle_name"`
	BirthDate       string                `json:"birth_date"` // YYYY-MM-DD
	PassportSeries  string                `json:"passport_series"`
	PassportNumber  string                `json:"passport_number"`
	PassportType    string                `json:"passport_type"`
	TaxNumber       string                `json:"tax_number"`
	Gender          string                `json:"gender"` // M or F
	BirthPlace      string                `json:"birth_place"`
	Phone           string                `json:"phone"`
	Email           string                `json:"email"`
	Notes           string                `json:"notes"`
	FamilyRelations []FamilyRelationInput `json:"family_relations"`
}

// FamilyMemberOutput represents a family member with their details
type FamilyMemberOutput struct {
	CitizenOutput
	RelationType string `json:"relation_type"`
}

// CitizenOutput is the output structure (with decrypted fields)
type CitizenOutput struct {
	ID              int64  `json:"id"`
	LastName        string `json:"last_name"`
	FirstName       string `json:"first_name"`
	MiddleName      string `json:"middle_name"`
	FullName        string `json:"full_name"` // Computed
	BirthDate       string `json:"birth_date"`
	PassportSeries  string `json:"passport_series"`
	PassportNumber  string `json:"passport_number"`
	PassportType    string `json:"passport_type"`
	PassportMasked  string `json:"passport_masked"` // For display
	TaxNumber       string `json:"tax_number"`
	TaxNumberMasked string `json:"tax_number_masked"` // For display
	Gender          string `json:"gender"`
	GenderDisplay   string `json:"gender_display"` // Чоловік/Жінка
	BirthPlace      string `json:"birth_place"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	Notes           string `json:"notes"`
	Deleted         bool   `json:"deleted"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	ActiveAddress   string `json:"active_address,omitempty"`
}

// CitizenListResult contains paginated list results
type CitizenListResult struct {
	Items      []*CitizenOutput `json:"items"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}
