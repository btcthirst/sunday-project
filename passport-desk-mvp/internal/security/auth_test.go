package security

import (
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretP@ssword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hashing failed: %v", err)
	}

	if hash == password {
		t.Error("hash should not match password")
	}

	if !VerifyPassword(hash, password) {
		t.Error("verification failed for correct password")
	}

	if VerifyPassword(hash, "wrongpassword") {
		t.Error("verification succeeded for incorrect password")
	}
}

func TestHashedPasswordsUnique(t *testing.T) {
	password := "SamePassword"

	h1, _ := HashPassword(password)
	h2, _ := HashPassword(password)

	if h1 == h2 {
		t.Error("bcrypt hashes should be unique due to random salt")
	}
}
