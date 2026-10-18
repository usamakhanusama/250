package auth

import (
	"testing"
)

func TestHashPasswordAndCheck(t *testing.T) {
	password := "SecretPassword@123"

	// 1. Hash password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if hash == password {
		t.Fatal("password hash must not match plaintext password")
	}

	// 2. Validate correct password
	if !CheckPassword(hash, password) {
		t.Fatal("CheckPassword failed for correct password")
	}

	// 3. Validate wrong password
	if CheckPassword(hash, "WrongPassword@123") {
		t.Fatal("CheckPassword should return false for incorrect password")
	}

	// 4. Validate short password rejection
	_, err = HashPassword("123")
	if err == nil {
		t.Fatal("expected error when hashing password shorter than 6 characters")
	}
}
