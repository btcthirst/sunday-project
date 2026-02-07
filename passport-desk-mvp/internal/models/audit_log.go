package models

import "time"

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID          int64     `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	OperatorID  int64     `json:"operator_id"`
	ActionType  string    `json:"action_type"` // CREATE, READ, UPDATE, DELETE, EXPORT, LOGIN, LOGOUT
	TableName   string    `json:"table_name"`
	RecordID    int64     `json:"record_id,omitempty"`
	Description string    `json:"description,omitempty"`

	// Додаткові поля для безпеки
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Success   bool   `json:"success"`
	ErrorMsg  string `json:"error_msg,omitempty"`

	// Metadata
	Metadata  string    `json:"metadata,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuditLogOutput includes operator name for display
type AuditLogOutput struct {
	AuditLog
	OperatorName string `json:"operator_name"`
}

var RequireAudit = map[string]bool{
	"LOGIN":           true, // Вхід в систему
	"LOGOUT":          true, // Вихід
	"FAILED_LOGIN":    true, // Невдалий вхід
	"CREATE":          true, // Створення будь-якого запису
	"UPDATE":          true, // Оновлення даних
	"DELETE":          true, // Видалення
	"EXPORT":          true, // Експорт даних
	"IMPORT":          true, // Імпорт даних
	"GENERATE_CERT":   true, // Генерація документів
	"VIEW_SENSITIVE":  true, // Перегляд чутливих даних (паспорт, ІПН)
	"UNLOCK":          true, // Розблокування після неактивності
	"PASSWORD_CHANGE": true, // Зміна пароля
}
