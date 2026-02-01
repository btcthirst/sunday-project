package services

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// BackupService handles automatic backups of the database and salt
type BackupService struct {
	dataDir    string
	backupDir  string
	maxBackups int
}

// NewBackupService creates a new backup service
func NewBackupService(dataDir string) *BackupService {
	return &BackupService{
		dataDir:    dataDir,
		backupDir:  filepath.Join(dataDir, "backups"),
		maxBackups: 7,
	}
}

// RunBackup performs a backup of the database and salt files
func (s *BackupService) RunBackup() error {
	// 1. Ensure backup directory exists
	if err := os.MkdirAll(s.backupDir, 0700); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupSubDir := filepath.Join(s.backupDir, timestamp)

	if err := os.MkdirAll(backupSubDir, 0700); err != nil {
		return fmt.Errorf("failed to create timestamped backup directory: %w", err)
	}

	// 2. Backup database
	dbPath := filepath.Join(s.dataDir, "passport_desk.db")
	if _, err := os.Stat(dbPath); err == nil {
		if err := s.copyFile(dbPath, filepath.Join(backupSubDir, "passport_desk.db")); err != nil {
			return fmt.Errorf("failed to backup database: %w", err)
		}
	}

	// 3. Backup salt
	saltPath := filepath.Join(s.dataDir, "db.salt")
	if _, err := os.Stat(saltPath); err == nil {
		if err := s.copyFile(saltPath, filepath.Join(backupSubDir, "db.salt")); err != nil {
			return fmt.Errorf("failed to backup salt: %w", err)
		}
	}

	// 4. Cleanup old backups
	return s.CleanupOldBackups()
}

// CleanupOldBackups removes backups older than the max limit
func (s *BackupService) CleanupOldBackups() error {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return err
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() {
			backups = append(backups, entry.Name())
		}
	}

	if len(backups) <= s.maxBackups {
		return nil
	}

	sort.Strings(backups)
	toDelete := backups[:len(backups)-s.maxBackups]

	for _, name := range toDelete {
		if err := os.RemoveAll(filepath.Join(s.backupDir, name)); err != nil {
			fmt.Printf("Warning: failed to delete old backup %s: %v\n", name, err)
		}
	}

	return nil
}

func (s *BackupService) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
