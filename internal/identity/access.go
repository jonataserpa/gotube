package identity

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTTL       = 15 * time.Minute
	accessExpiresIn = 900
	refreshTTL      = 30 * 24 * time.Hour
	tokenIssuer     = "gotube"
	tokenAudience   = "gotube-web"
)

type accessClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

// IssueAccessToken signs a 15-minute HS256 token. Other algorithms are rejected on parse.
func IssueAccessToken(secret string, account Account, now time.Time) (string, error) {
	claims := accessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   account.ID,
			Issuer:    tokenIssuer,
			Audience:  jwt.ClaimStrings{tokenAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
		},
		Email: account.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseAccessToken checks signature, issuer, audience and expiry.
func ParseAccessToken(secret string, raw string) (Account, error) {
	claims := &accessClaims{}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithAudience(tokenAudience),
	)
	parsed, err := parser.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid || claims.Subject == "" || claims.Email == "" {
		return Account{}, ErrInvalidCredentials
	}
	return Account{ID: claims.Subject, Email: claims.Email}, nil
}
