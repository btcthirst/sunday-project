package database

// Migrate runs database migrations
func (d *Database) Migrate() error {
	migrations := []string{
		// Operators table
		`CREATE TABLE IF NOT EXISTS operators (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			full_name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
			FOREIGN KEY (operator_id) REFERENCES operators(id)
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
			FOREIGN KEY (citizen_id) REFERENCES citizens(id)
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
			FOREIGN KEY (citizen_id) REFERENCES citizens(id),
			FOREIGN KEY (issued_by) REFERENCES operators(id)
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

	return nil
}
