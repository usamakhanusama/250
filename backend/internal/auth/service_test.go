package auth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"sust-cse-portal/internal/config"
)

func newTestService() *Service {
	cfg := &config.Config{
		JWTSecret:        "test_access_secret_key_that_is_long_enough",
		JWTRefreshSecret: "test_refresh_secret_key_that_is_long_enough",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 24 * time.Hour,
		MaxLoginAttempts: 3,
		LockoutDuration:  15 * time.Minute,
	}
	return NewService(NewMemoryStore(), cfg)
}

func validSignup() SignupRequest {
	return SignupRequest{
		RegNo: "2023331045", FullName: "Test Student", Email: "Test@Student.sust.edu",
		Password: "Secret123", ConfirmPassword: "Secret123",
	}
}

func TestValidateRegNo(t *testing.T) {
	good := []string{"2023331000", "2023331020", "2023331999"}
	bad := []string{"", "2022331020", "2023332020", "202333102", "20233310200", "2023331abc", " 2023331020"}
	for _, r := range good {
		if ValidateRegNo(r) != nil {
			t.Errorf("%q should be valid", r)
		}
	}
	for _, r := range bad {
		if ValidateRegNo(r) == nil {
			t.Errorf("%q should be rejected", r)
		}
	}
}

func TestSignupThenLogin(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	res, err := svc.Signup(ctx, validSignup())
	if err != nil {
		t.Fatalf("signup failed: %v", err)
	}
	if res.AccessToken == "" || res.User.Role != "student" || res.User.Student.RegNo != "2023331045" {
		t.Fatalf("unexpected signup response: %+v", res)
	}
	if res.User.Email != "test@student.sust.edu" {
		t.Errorf("email should be normalised to lower case, got %s", res.User.Email)
	}

	// The JSON sent to the browser must never contain the password or its hash.
	body, _ := json.Marshal(res)
	if strings.Contains(string(body), "Secret123") || strings.Contains(strings.ToLower(string(body)), "password") || strings.Contains(string(body), "$2a$") {
		t.Fatalf("response leaks password data: %s", body)
	}

	// Login works with reg no and with email
	for _, login := range []string{"2023331045", "test@student.sust.edu"} {
		if _, err := svc.Login(ctx, LoginRequest{Login: login, Password: "Secret123"}); err != nil {
			t.Errorf("login with %q failed: %v", login, err)
		}
	}

	// The JWT carries the reg no
	claims, err := ValidateToken(res.AccessToken, svc.cfg.JWTSecret)
	if err != nil || claims.RegNo != "2023331045" {
		t.Fatalf("expected reg_no in JWT, got %+v (%v)", claims, err)
	}
}

func TestSignupRejections(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	_, _ = svc.Signup(ctx, validSignup())

	cases := map[string]struct {
		mutate func(*SignupRequest)
		want   error
	}{
		"other batch":     {func(r *SignupRequest) { r.RegNo = "2022331045" }, ErrInvalidRegNo},
		"other dept":      {func(r *SignupRequest) { r.RegNo = "2023332045" }, ErrInvalidRegNo},
		"weak password":   {func(r *SignupRequest) { r.Password, r.ConfirmPassword = "password", "password" }, ErrWeakPassword},
		"mismatch":        {func(r *SignupRequest) { r.ConfirmPassword = "Different1" }, ErrPasswordsDiffer},
		"duplicate reg":   {func(r *SignupRequest) { r.Email = "new@student.sust.edu" }, ErrRegNoExists},
		"duplicate email": {func(r *SignupRequest) { r.RegNo = "2023331046" }, ErrEmailExists},
		"short name":      {func(r *SignupRequest) { r.RegNo, r.Email, r.FullName = "2023331047", "x@y.z", "Al" }, ErrInvalidName},
	}
	for name, c := range cases {
		req := validSignup()
		c.mutate(&req)
		if _, err := svc.Signup(ctx, req); !errors.Is(err, c.want) {
			t.Errorf("%s: expected %v, got %v", name, c.want, err)
		}
	}
}

func TestLoginErrorsAreGenericAndLockout(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	_, _ = svc.Signup(ctx, validSignup())

	// Unknown user and wrong password give the SAME error (no user enumeration)
	_, errUnknown := svc.Login(ctx, LoginRequest{Login: "2023331999", Password: "whatever1"})
	_, errWrong := svc.Login(ctx, LoginRequest{Login: "2023331045", Password: "wrong1234"})
	if !errors.Is(errUnknown, ErrInvalidCredentials) || !errors.Is(errWrong, ErrInvalidCredentials) {
		t.Fatalf("expected generic invalid credentials, got %v / %v", errUnknown, errWrong)
	}

	// MaxLoginAttempts = 3: the 3rd wrong password locks the account
	_, _ = svc.Login(ctx, LoginRequest{Login: "2023331045", Password: "wrong1234"})
	_, err := svc.Login(ctx, LoginRequest{Login: "2023331045", Password: "wrong1234"})
	if !errors.Is(err, ErrAccountLocked) {
		t.Fatalf("expected account lock, got %v", err)
	}
	// Even the correct password is refused while locked
	if _, err := svc.Login(ctx, LoginRequest{Login: "2023331045", Password: "Secret123"}); !errors.Is(err, ErrAccountLocked) {
		t.Fatalf("locked account must refuse correct password, got %v", err)
	}
}

func TestRefreshRotationAndLogout(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()
	res, _ := svc.Signup(ctx, validSignup())

	newRes, err := svc.Refresh(ctx, res.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	// The old refresh token was rotated out and cannot be reused
	if _, err := svc.Refresh(ctx, res.RefreshToken); err == nil {
		t.Fatal("reusing an old refresh token must fail")
	}
	// After logout the new one is revoked too
	_ = svc.Logout(ctx, newRes.RefreshToken)
	if _, err := svc.Refresh(ctx, newRes.RefreshToken); err == nil {
		t.Fatal("refresh after logout must fail")
	}
}
