package security

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "passport-desk-mvp"
	saltKeyName = "db-salt"
)

var (
	ErrSaltNotFound = errors.New("salt not found")
)

// Keystore handles secure storage of keys and salt
type Keystore struct {
	dataDir string
}

// NewKeystore creates a new keystore
func NewKeystore(dataDir string) *Keystore {
	return &Keystore{dataDir: dataDir}
}

// SaveSalt saves the database salt to a file
func (k *Keystore) SaveSalt(salt []byte) error {
	saltPath := filepath.Join(k.dataDir, "db.salt")
	return os.WriteFile(saltPath, salt, 0600)
}

// LoadSalt loads the database salt from file
func (k *Keystore) LoadSalt() ([]byte, error) {
	saltPath := filepath.Join(k.dataDir, "db.salt")
	salt, err := os.ReadFile(saltPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSaltNotFound
		}
		return nil, err
	}
	return salt, nil
}

// SaltExists checks if salt file exists
func (k *Keystore) SaltExists() bool {
	saltPath := filepath.Join(k.dataDir, "db.salt")
	_, err := os.Stat(saltPath)
	return err == nil
}

// SaveToSystemKeyring attempts to save a secret to system keyring
func (k *Keystore) SaveToSystemKeyring(key, value string) error {
	return keyring.Set(serviceName, key, value)
}

// LoadFromSystemKeyring attempts to load a secret from system keyring
func (k *Keystore) LoadFromSystemKeyring(key string) (string, error) {
	value, err := keyring.Get(serviceName, key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", fmt.Errorf("key not found in keyring: %s", key)
		}
		return "", err
	}
	return value, nil
}

// DeleteFromSystemKeyring removes a secret from system keyring
func (k *Keystore) DeleteFromSystemKeyring(key string) error {
	return keyring.Delete(serviceName, key)
}

// EnsureDataDir ensures the data directory exists
func (k *Keystore) EnsureDataDir() error {
	return os.MkdirAll(k.dataDir, 0700)
}

// GetDBPath returns the path to the encrypted database
func (k *Keystore) GetDBPath() string {
	return filepath.Join(k.dataDir, "passport_desk.db")
}
