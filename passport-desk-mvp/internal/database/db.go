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

// Rekey changes the database encryption key
func (d *Database) Rekey(newKey string) error {
	_, err := d.db.Exec(fmt.Sprintf("PRAGMA rekey = x'%s'", newKey))
	return err
}
