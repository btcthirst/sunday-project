package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

// Database wraps the SQLCipher connection
type Database struct {
	db *sql.DB
}

// New creates a new encrypted database connection
func New(dbPath string, encryptionKey string) (*Database, error) {
	// SQLCipher connection with encryption key
	dsn := fmt.Sprintf("file:%s?_pragma_key=x'%s'&_pragma_cipher_page_size=4096", dbPath, encryptionKey)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection by running a simple query
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to verify database connection: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	database := &Database{db: db}

	// Run migrations
	if err := database.Migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// DB returns the underlying sql.DB for advanced queries
func (d *Database) DB() *sql.DB {
	return d.db
}

/*
// CreateOperator creates a new operator
func (d *Database) CreateOperator(op *Operator) error {
	result, err := d.db.Exec(
		`INSERT INTO operators (username, password_hash, full_name) VALUES (?, ?, ?)`,
		op.Username, op.PasswordHash, op.FullName,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	op.ID = id
	return nil
}

// GetOperatorByUsername retrieves an operator by username
func (d *Database) GetOperatorByUsername(username string) (*Operator, error) {
	op := &Operator{}
	err := d.db.QueryRow(
		`SELECT id, username, password_hash, full_name, created_at, updated_at FROM operators WHERE username = ?`,
		username,
	).Scan(&op.ID, &op.Username, &op.PasswordHash, &op.FullName, &op.CreatedAt, &op.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return op, nil
}

// OperatorExists checks if any operators exist in the database
func (d *Database) OperatorExists() (bool, error) {
	var count int
	err := d.db.QueryRow(`SELECT COUNT(*) FROM operators`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

/*

// LogAudit creates an audit log entry
func (d *Database) LogAudit(log *AuditLog) error {
	_, err := d.db.Exec(
		`INSERT INTO audit_log (operator_id, action_type, table_name, record_id, description) VALUES (?, ?, ?, ?, ?)`,
		log.OperatorID, log.ActionType, log.TableName, log.RecordID, log.Description,
	)
	return err
}

// GetAuditLogs retrieves recent audit logs
func (d *Database) GetAuditLogs(limit int) ([]AuditLogOutput, error) {
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

	var logs []AuditLogOutput
	for rows.Next() {
		var l AuditLogOutput
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
*/
