package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// GenerateDBSecret generates a random 32-byte secret for database encryption
func GenerateDBSecret() ([]byte, error) {
	secret := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, secret); err != nil {
		return nil, fmt.Errorf("failed to generate db secret: %w", err)
	}
	return secret, nil
}

// GenerateRecoveryKey generates a random string to be used as a master recovery key
func GenerateRecoveryKey() (string, error) {
	bytes := make([]byte, 16) // 128 bits of entropy
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", fmt.Errorf("failed to generate recovery key: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// WrapSecret encrypts a secret (e.g. DBSecret) with a wrapping key (e.g. derived from password)
func WrapSecret(secret, wrappingKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(wrappingKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return gcm.Seal(nonce, nonce, secret, nil), nil
}

// UnwrapSecret decrypts a secret (e.g. DBSecret) using a wrapping key
func UnwrapSecret(wrappedSecret, wrappingKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(wrappingKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(wrappedSecret) < gcm.NonceSize() {
		return nil, fmt.Errorf("wrapped secret too short")
	}

	nonce := wrappedSecret[:gcm.NonceSize()]
	ciphertext := wrappedSecret[gcm.NonceSize():]

	secret, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to unwrap secret: %w", err)
	}

	return secret, nil
}
