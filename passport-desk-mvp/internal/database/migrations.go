package database

import (
	"fmt"
	"strings"
)

// Migrate runs database migrations
func (d *Database) Migrate() error {
	migrations := []string{
		// Operators table
		`CREATE TABLE IF NOT EXISTS operators (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			full_name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Citizens table
		`CREATE TABLE IF NOT EXISTS citizens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			last_name TEXT NOT NULL,
			first_name TEXT NOT NULL,
			middle_name TEXT,
			birth_date DATE NOT NULL,
			passport_series TEXT NOT NULL,
			passport_number TEXT NOT NULL,
			tax_number TEXT,
			gender TEXT CHECK(gender IN ('M', 'F')),
			birth_place TEXT,
			phone TEXT,
			email TEXT,
			notes TEXT,
			deleted BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Registrations table
		`CREATE TABLE IF NOT EXISTS registrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			citizen_id INTEGER NOT NULL,
			registration_type TEXT CHECK(registration_type IN ('permanent', 'temporary')),
			region TEXT NOT NULL,
			district TEXT,
			settlement TEXT NOT NULL,
			street TEXT NOT NULL,
			house_number TEXT NOT NULL,
			apartment_number TEXT,
			registration_date DATE NOT NULL,
			deregistration_date DATE,
			basis_document TEXT,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (citizen_id) REFERENCES citizens(id)
		)`,

		// Audit log table
		`CREATE TABLE IF NOT EXISTS audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			operator_id INTEGER NOT NULL,
			action_type TEXT NOT NULL,
			table_name TEXT NOT NULL,
			record_id INTEGER,
			description TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (operator_id) REFERENCES operators(id)
		)`,

		// Certificates table
		`CREATE TABLE IF NOT EXISTS certificates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			certificate_number TEXT UNIQUE NOT NULL,
			citizen_id INTEGER NOT NULL,
			issue_date DATE NOT NULL,
			purpose TEXT,
			issued_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (citizen_id) REFERENCES citizens(id),
			FOREIGN KEY (issued_by) REFERENCES operators(id)
		)`,

		// Family relations table
		`CREATE TABLE IF NOT EXISTS family_relations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			citizen_id INTEGER NOT NULL,
			member_id INTEGER NOT NULL,
			relation_type TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (citizen_id) REFERENCES citizens(id),
			FOREIGN KEY (member_id) REFERENCES citizens(id),
			UNIQUE(citizen_id, member_id)
		)`,

		// Indexes for better performance
		`CREATE INDEX IF NOT EXISTS idx_citizens_last_name ON citizens(last_name)`,
		`CREATE INDEX IF NOT EXISTS idx_citizens_passport ON citizens(passport_series, passport_number)`,
		`CREATE INDEX IF NOT EXISTS idx_registrations_citizen ON registrations(citizen_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_log(timestamp)`,
	}

	for _, m := range migrations {
		if _, err := d.db.Exec(m); err != nil {
			return err
		}
	}

	// Add missing columns for existing databases
	tablesToUpdate := []struct {
		table   string
		columns []string
	}{
		{"operators", []string{"updated_at DATETIME DEFAULT CURRENT_TIMESTAMP"}},
		{"registrations", []string{"updated_at DATETIME DEFAULT CURRENT_TIMESTAMP"}},
		{"certificates", []string{"updated_at DATETIME DEFAULT CURRENT_TIMESTAMP"}},
		{"audit_log", []string{"updated_at DATETIME DEFAULT CURRENT_TIMESTAMP"}},
		{"family_relations", []string{
			"created_at DATETIME DEFAULT CURRENT_TIMESTAMP",
			"updated_at DATETIME DEFAULT CURRENT_TIMESTAMP",
		}},
	}

	for _, t := range tablesToUpdate {
		for _, col := range t.columns {
			colName := strings.Split(col, " ")[0]
			var count int
			err := d.db.QueryRow(fmt.Sprintf("SELECT count(*) FROM pragma_table_info('%s') WHERE name='%s'", t.table, colName)).Scan(&count)
			if err == nil && count == 0 {
				d.db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", t.table, col))
			}
		}
	}

	// Add triggers for automatic updated_at
	triggerTables := []string{"operators", "citizens", "registrations", "audit_log", "certificates", "family_relations"}
	for _, table := range triggerTables {
		triggerQuery := fmt.Sprintf(`
			CREATE TRIGGER IF NOT EXISTS trg_%s_updated_at
			AFTER UPDATE ON %s
			BEGIN
				UPDATE %s SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
			END;`, table, table, table)
		d.db.Exec(triggerQuery)
	}

	// Add passport_type if not exists for existing databases
	var count int
	err := d.db.QueryRow("SELECT count(*) FROM pragma_table_info('citizens') WHERE name='passport_type'").Scan(&count)
	if err == nil && count == 0 {
		d.db.Exec("ALTER TABLE citizens ADD COLUMN passport_type TEXT DEFAULT 'old'")
	}

	return nil
}
