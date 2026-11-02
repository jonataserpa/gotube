package identity

import "strings"

// NormalizeEmail trims and lowercases an address. Empty values and values
// without @ are rejected.
func NormalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" || strings.ContainsAny(email, " \t") || !strings.Contains(email, "@") {
		return "", ErrValidation
	}
	return email, nil
}

// ValidatePassword accepts a password of at least 8 characters.
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrValidation
	}
	return nil
}
