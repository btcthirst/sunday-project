package security

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

func TestDatabaseEncryption(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "security_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	key := "test-encryption-key-123"

	// 1. Create an encrypted database and a table
	dsnEnc := fmt.Sprintf("file:%s?_pragma_key=x'%s'&_pragma_cipher_page_size=4096", dbPath, hexKey(key))
	db, err := sql.Open("sqlite3", dsnEnc)
	if err != nil {
		t.Fatalf("failed to open encrypted db: %v", err)
	}

	_, err = db.Exec("CREATE TABLE secret (id INTEGER PRIMARY KEY, data TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	_, err = db.Exec("INSERT INTO secret (data) VALUES ('SUPER SECRET DATA')")
	if err != nil {
		t.Fatalf("failed to insert data: %v", err)
	}
	db.Close()

	// 2. Attempt to open without SQLCipher params (standard SQLite)
	dsnNoKey := dbPath
	dbNoKey, err := sql.Open("sqlite3", dsnNoKey)
	if err != nil {
		t.Fatalf("failed to open db without key: %v", err)
	}
	defer dbNoKey.Close()

	// This should fail because the file is encrypted
	var data string
	err = dbNoKey.QueryRow("SELECT data FROM secret").Scan(&data)
	if err == nil {
		t.Error("should NOT be able to read data from encrypted database without a key")
	} else {
		t.Logf("Correctly failed to read without key: %v", err)
	}

	// 3. Attempt to open with WRONG key
	dsnWrongKey := fmt.Sprintf("file:%s?_pragma_key=x'%s'", dbPath, hexKey("wrong-key"))
	dbWrong, _ := sql.Open("sqlite3", dsnWrongKey)
	defer dbWrong.Close()

	err = dbWrong.QueryRow("SELECT data FROM secret").Scan(&data)
	if err == nil {
		t.Error("should NOT be able to read data from encrypted database with WRONG key")
	} else {
		t.Logf("Correctly failed to read with wrong key: %v", err)
	}
}

// Helper to simulate key derivation hex
func hexKey(p string) string {
	hash := sha256.Sum256([]byte(p))
	return hex.EncodeToString(hash[:])
}
