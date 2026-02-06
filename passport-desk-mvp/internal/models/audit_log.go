package models

import "time"

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
}
