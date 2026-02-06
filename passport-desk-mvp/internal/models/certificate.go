package models

import "time"

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
