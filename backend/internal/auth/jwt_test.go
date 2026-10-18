package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	accessSecret := "test_access_secret_key_12345"
	refreshSecret := "test_refresh_secret_key_67890"
	userID := "user-1234-uuid"
	role := "student"

	tokens, hash, err := GenerateTokenPair(userID, role, accessSecret, refreshSecret, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("expected no error generating tokens, got: %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected non-empty access and refresh tokens")
	}

	if hash == "" {
		t.Fatal("expected non-empty refresh token hash")
	}

	// Validate access token
	claims, err := ValidateToken(tokens.AccessToken, accessSecret)
	if err != nil {
		t.Fatalf("failed validating valid access token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}

	if claims.Role != role {
		t.Errorf("expected role %s, got %s", role, claims.Role)
	}

	// Validate with wrong secret must fail
	_, err = ValidateToken(tokens.AccessToken, "wrong_secret")
	if err == nil {
		t.Fatal("expected validation to fail with wrong secret, but succeeded")
	}
}

func TestJWTAlgorithmValidation(t *testing.T) {
	// Attempt token forgery using 'none' algorithm
	claims := CustomClaims{
		UserID: "attacker",
		Role:   "admin",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := ValidateToken(tokenStr, "any_secret")
	if err == nil {
		t.Fatal("security violation: token signed with 'none' alg must be rejected by keyfunc")
	}
}

func TestHashTokenConsistency(t *testing.T) {
	token := "sample_refresh_token_string"
	hash1 := HashToken(token)
	hash2 := HashToken(token)

	if hash1 != hash2 {
		t.Fatalf("SHA-256 hash must be deterministic: %s != %s", hash1, hash2)
	}
}
