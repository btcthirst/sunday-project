package repository

import (
	"database/sql"
	"sunday-project/passport-desk-mvp/internal/database"
	"sunday-project/passport-desk-mvp/internal/models"
)

// LogAudit creates an audit log entry
func (d *database.Database) LogAudit(log *models.AuditLog) error {
	_, err := d.db.Exec(
		`INSERT INTO audit_log (operator_id, action_type, table_name, record_id, description) VALUES (?, ?, ?, ?, ?)`,
		log.OperatorID, log.ActionType, log.TableName, log.RecordID, log.Description,
	)
	return err
}

// GetAuditLogs retrieves recent audit logs
func (d *database.Database) GetAuditLogs(limit int) ([]models.AuditLogOutput, error) {
	rows, err := d.db.Query(`
		SELECT a.id, a.timestamp, a.operator_id, a.action_type, a.table_name, a.record_id, a.description, a.updated_at, o.full_name
		FROM audit_log a
		LEFT JOIN operators o ON a.operator_id = o.id
		ORDER BY a.timestamp DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLogOutput
	for rows.Next() {
		var l models.AuditLogOutput
		var opName sql.NullString
		err := rows.Scan(&l.ID, &l.Timestamp, &l.OperatorID, &l.ActionType, &l.TableName, &l.RecordID, &l.Description, &l.UpdatedAt, &opName)
		if err != nil {
			return nil, err
		}
		if opName.Valid {
			l.OperatorName = opName.String
		} else {
			l.OperatorName = "Unknown"
		}
		logs = append(logs, l)
	}
	return logs, nil
}
