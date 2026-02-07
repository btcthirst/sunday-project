package repository

import (
	"context"
	"database/sql"
	"passport-desk-mvp/internal/database"
	"passport-desk-mvp/internal/models"
)

type AuditLogRepository struct {
	db *database.Database
}

func NewAuditLogRepository(db *database.Database) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// LogAudit creates an audit log entry
func (a *AuditLogRepository) LogAudit(ctx context.Context, log *models.AuditLog) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := a.db.DB().ExecContext(
		ctx,
		`INSERT INTO audit_log (operator_id, action_type, table_name, record_id, description) VALUES (?, ?, ?, ?, ?)`,
		log.OperatorID, log.ActionType, log.TableName, log.RecordID, log.Description,
	)
	return err
}

// GetAuditLogs retrieves recent audit logs
func (d *AuditLogRepository) GetAuditLogs(ctx context.Context, limit int) ([]models.AuditLogOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := d.db.DB().QueryContext(
		ctx,
		`
		SELECT a.id, a.timestamp, a.operator_id, a.action_type, a.table_name, a.record_id, a.description, a.updated_at, o.full_name
		FROM audit_log a
		LEFT JOIN operators o ON a.operator_id = o.id
		ORDER BY a.timestamp DESC
		LIMIT ?`,
		limit,
	)
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
