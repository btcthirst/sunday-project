package repository

import (
	"context"
	"database/sql"
	"passport-desk-mvp/internal/database"
	"testing"
)

// setupTestDB creates an in-memory database for testing
func setupTestDB(t *testing.T) *database.Database {
	// Use in-memory database
	// db.go adds "file:" prefix and parameters. passing ":memory:" results in "file::memory:?..."
	db, err := database.New(":memory:", "test-key")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// cleanTable clears a table for fresh tests (if not re-creating DB every time)
// Since we use in-memory DB in New(), if we call setupTestDB for each test, we get a fresh DB?
// "file::memory:?cache=shared" might share data if connection is open.
// But database.New creates a new sql.Open.
// To be safe, we can truncate tables or just rely on t.Cleanup closing it.
// Actually, with "mode=memory", once the last connection closes, data is lost.
// So setupTestDB per test function is safe.
func cleanTable(t *testing.T, db *database.Database, tableName string) {
	_, err := db.DB().Exec("DELETE FROM " + tableName)
	if err != nil {
		t.Fatalf("Failed to clean table %s: %v", tableName, err)
	}
}

// Helper to get raw SQL DB for assertions
func getRawDB(db *database.Database) *sql.DB {
	return db.DB()
}

// Mock context with timeout
func testContext(t *testing.T) context.Context {
	return context.Background()
}
