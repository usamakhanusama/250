package auth

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// BatchRegPrefix: this portal is only for SUST CSE batch 2023 (session 2023-24).
// SUST reg no format: 2023 (year) + 331 (CSE dept code) + 3-digit roll.
const BatchRegPrefix = "2023331"

// regNoPattern = "2023331" followed by exactly 3 digits, e.g. 2023331020.
var regNoPattern = regexp.MustCompile(`^` + BatchRegPrefix + `\d{3}$`)

var (
	ErrInvalidRegNo    = errors.New("registration number must be a CSE 2023 batch number (2023331XXX, 10 digits)")
	ErrWeakPassword    = errors.New("password must be at least 8 characters and contain at least one letter and one digit")
	ErrInvalidName     = errors.New("full name must be between 3 and 100 characters")
	ErrPasswordsDiffer = errors.New("password and confirm password do not match")
)

// ValidateRegNo returns nil only for a valid 2023 batch CSE registration number.
func ValidateRegNo(regNo string) error {
	if !regNoPattern.MatchString(regNo) {
		return ErrInvalidRegNo
	}
	return nil
}

// ValidatePasswordStrength enforces a simple password policy:
// at least 8 characters, at least one letter and at least one digit.
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 || len(password) > 72 { // bcrypt only uses the first 72 bytes
		return ErrWeakPassword
	}
	hasLetter, hasDigit := false, false
	for _, ch := range password {
		if unicode.IsLetter(ch) {
			hasLetter = true
		}
		if unicode.IsDigit(ch) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return ErrWeakPassword
	}
	return nil
}

// normalizeSignup trims spaces and lower-cases the email so
// "Afif@Student.sust.edu " and "afif@student.sust.edu" count as the same account.
func normalizeSignup(req *SignupRequest) {
	req.RegNo = strings.TrimSpace(req.RegNo)
	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
}
