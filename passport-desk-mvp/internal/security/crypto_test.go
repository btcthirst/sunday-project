package security

import (
	"bytes"
	"testing"
)

func TestGenerateSalt(t *testing.T) {
	s1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt 1: %v", err)
	}
	s2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt 2: %v", err)
	}

	if len(s1) != saltLength {
		t.Errorf("expected salt length %d, got %d", saltLength, len(s1))
	}

	if bytes.Equal(s1, s2) {
		t.Error("salts should be unique")
	}
}

func TestDeriveKey(t *testing.T) {
	password := "verysecretpassword"
	salt, _ := GenerateSalt()

	k1 := DeriveKey(password, salt)
	k2 := DeriveKey(password, salt)

	if !bytes.Equal(k1, k2) {
		t.Error("derived keys should be consistent for same password/salt")
	}

	if len(k1) != keyLength {
		t.Errorf("expected key length %d, got %d", keyLength, len(k1))
	}

	k3 := DeriveKey("otherpassword", salt)
	if bytes.Equal(k1, k3) {
		t.Error("different passwords should produce different keys")
	}
}

func TestEncryptionDecryption(t *testing.T) {
	key := []byte("01234567890123456789012345678901") // 32 bytes
	crypto := NewCrypto(key)
	plaintext := "Sensitive data example 123!"

	ciphertext, err := crypto.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if ciphertext == plaintext {
		t.Error("ciphertext should not match plaintext")
	}

	decrypted, err := crypto.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDecryptionFailure(t *testing.T) {
	key1 := []byte("keynumberone12345678901234567890")
	key2 := []byte("keynumbertwo12345678901234567890")

	c1 := NewCrypto(key1)
	c2 := NewCrypto(key2)

	plaintext := "Secret info"
	ciphertext, _ := c1.Encrypt(plaintext)

	_, err := c2.Decrypt(ciphertext)
	if err == nil {
		t.Error("decryption should fail with incorrect key")
	}
}

func TestEmptyPlaintext(t *testing.T) {
	key := make([]byte, 32)
	crypto := NewCrypto(key)

	ciphertext, err := crypto.Encrypt("")
	if err != nil {
		t.Fatal(err)
	}
	if ciphertext != "" {
		t.Error("empty plaintext should result in empty ciphertext")
	}

	decrypted, err := crypto.Decrypt("")
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != "" {
		t.Error("empty ciphertext should result in empty plaintext")
	}
}
