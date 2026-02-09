package services

import (
	"fmt"
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

	// Test CleanupOldBackups
	// Create more than 7 backups (maxBackups is 7)
	for i := 0; i < 10; i++ {
		timestamp := fmt.Sprintf("2023010%d_000000", i)
		os.MkdirAll(filepath.Join(tempDir, "backups", timestamp), 0700)
	}

	err = service.CleanupOldBackups()
	if err != nil {
		t.Fatalf("CleanupOldBackups failed: %v", err)
	}

	entries, _ := os.ReadDir(filepath.Join(tempDir, "backups"))
	// Should be 7 max
	if len(entries) != 7 {
		t.Errorf("Expected 7 backups after cleanup, got %d", len(entries))
	}
}
