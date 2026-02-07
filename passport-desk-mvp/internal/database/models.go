package database

import "time"

/*
// Operator represents a system operator
type Operator struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never sent to frontend
	FullName     string    `json:"full_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

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

/*
// AuditLog represents an audit trail entry
type AuditLog struct {
	ID          int64     `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	OperatorID  int64     `json:"operator_id"`
	ActionType  string    `json:"action_type"` // CREATE, READ, UPDATE, DELETE
	TableName   string    `json:"table_name"`
	RecordID    int64     `json:"record_id,omitempty"`
	Description string    `json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AuditLogOutput includes operator name for display
type AuditLogOutput struct {
	AuditLog
	OperatorName string `json:"operator_name"`
}*/

// Certificate represents an issued certificate
type Certificate struct {
	ID                int64     `json:"id"`
	CertificateNumber string    `json:"certificate_number"`
	CitizenID         int64     `json:"citizen_id"`
	IssueDate         string    `json:"issue_date"`
	Purpose           string    `json:"purpose,omitempty"`
	IssuedBy          int64     `json:"issued_by"` // operator_id
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
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
