package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupService(t *testing.T) {
	// Setup temp dir for backups
	tempDir, err := os.MkdirTemp("", "backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy DB file to backup
	dbPath := filepath.Join(tempDir, "passport_desk.db")
	err = os.WriteFile(dbPath, []byte("dummy data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy DB: %v", err)
	}

	service := NewBackupService(tempDir)

	// Test RunBackup
	err = service.RunBackup()
	if err != nil {
		t.Fatalf("RunBackup failed: %v", err)
	}

	// Verify backup was created (it creates a subfolder with timestamp)
	entries, _ := os.ReadDir(filepath.Join(tempDir, "backups"))
	if len(entries) == 0 {
		t.Error("No backup directory created")
	}
}
