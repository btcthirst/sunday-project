package security

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestGenerateDBSecret(t *testing.T) {
	secret1, err := GenerateDBSecret()
	if err != nil {
		t.Fatalf("Failed to generate DB secret: %v", err)
	}

	if len(secret1) != 32 {
		t.Errorf("Expected secret length 32, got %d", len(secret1))
	}

	secret2, err := GenerateDBSecret()
	if err != nil {
		t.Fatalf("Failed to generate second DB secret: %v", err)
	}

	if bytes.Equal(secret1, secret2) {
		t.Error("GenerateDBSecret produced identical secrets (lack of randomness)")
	}
}

func TestGenerateRecoveryKey(t *testing.T) {
	key1, err := GenerateRecoveryKey()
	if err != nil {
		t.Fatalf("Failed to generate recovery key: %v", err)
	}

	// 16 bytes = 32 hex chars
	if len(key1) != 32 {
		t.Errorf("Expected recovery key length 32 (hex), got %d", len(key1))
	}

	// Verify it's valid hex
	_, err = hex.DecodeString(key1)
	if err != nil {
		t.Errorf("Recovery key is not valid hex: %v", err)
	}

	key2, err := GenerateRecoveryKey()
	if err != nil {
		t.Fatalf("Failed to generate second recovery key: %v", err)
	}

	if key1 == key2 {
		t.Error("GenerateRecoveryKey produced identical keys (lack of randomness)")
	}
}

func TestWrapUnwrapSecret(t *testing.T) {
	secret := []byte("this is a high entropy secret key")
	wrappingKey := make([]byte, 32)
	for i := range wrappingKey {
		wrappingKey[i] = byte(i)
	}

	// Test Round-trip
	wrapped, err := WrapSecret(secret, wrappingKey)
	if err != nil {
		t.Fatalf("WrapSecret failed: %v", err)
	}

	if bytes.Equal(secret, wrapped) {
		t.Error("Wrapped secret is identical to plaintext")
	}

	unwrapped, err := UnwrapSecret(wrapped, wrappingKey)
	if err != nil {
		t.Fatalf("UnwrapSecret failed: %v", err)
	}

	if !bytes.Equal(secret, unwrapped) {
		t.Errorf("Unwrapped secret does not match original. Expected %x, got %x", secret, unwrapped)
	}

	// Test with wrong key
	wrongKey := make([]byte, 32)
	wrongKey[0] = 0xFF
	_, err = UnwrapSecret(wrapped, wrongKey)
	if err == nil {
		t.Error("UnwrapSecret should have failed with wrong key")
	}

	// Test with invalid wrapped data length
	_, err = UnwrapSecret([]byte("too-short"), wrappingKey)
	if err == nil {
		t.Error("UnwrapSecret should have failed with too short data")
	}

	// Test tampering
	if len(wrapped) > 12 {
		wrapped[len(wrapped)-1] ^= 0xFF
		_, err = UnwrapSecret(wrapped, wrappingKey)
		if err == nil {
			t.Error("UnwrapSecret should have failed with tampered data")
		}
	}

	// Test invalid wrapping key length (AES only supports 16, 24, 32)
	_, err = WrapSecret(secret, []byte("short"))
	if err == nil {
		t.Error("WrapSecret should have failed with invalid key length")
	}
}
