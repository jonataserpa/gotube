package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory  = 19 * 1024
	argonTime    = 2
	argonThreads = 1
	argonKeyLen  = 32
	saltLen      = 16
)

// HashPassword returns an Argon2id PHC string using the OWASP parameter set.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	sum := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return formatPHC(salt, sum), nil
}

// VerifyPassword reports whether password matches an Argon2id PHC string.
func VerifyPassword(encoded string, password string) bool {
	salt, sum, ok := parsePHC(encoded)
	if !ok {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return subtle.ConstantTimeCompare(sum, got) == 1
}

func formatPHC(salt []byte, sum []byte) string {
	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum),
	)
}

func parsePHC(encoded string) ([]byte, []byte, bool) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return nil, nil, false
	}
	memory, timeCost, threads, ok := parseArgonParams(parts[3])
	if !ok || memory != argonMemory || timeCost != argonTime || threads != argonThreads {
		return nil, nil, false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, false
	}
	sum, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(sum) != argonKeyLen {
		return nil, nil, false
	}
	return salt, sum, true
}

func parseArgonParams(raw string) (uint32, uint32, uint8, bool) {
	var memory, timeCost uint32
	var threads uint8
	for _, field := range strings.Split(raw, ",") {
		key, value, found := strings.Cut(field, "=")
		if !found {
			return 0, 0, 0, false
		}
		number, err := strconv.Atoi(value)
		if err != nil || number < 1 {
			return 0, 0, 0, false
		}
		switch key {
		case "m":
			memory = uint32(number)
		case "t":
			timeCost = uint32(number)
		case "p":
			threads = uint8(number)
		default:
			return 0, 0, 0, false
		}
	}
	if memory == 0 || timeCost == 0 || threads == 0 {
		return 0, 0, 0, false
	}
	return memory, timeCost, threads, true
}
