package identity

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
)

// Session is an authenticated account plus the tokens issued for it.
// RefreshToken is delivered only as an HttpOnly cookie by the HTTP layer.
type Session struct {
	Account      Account
	AccessToken  string
	ExpiresIn    int
	RefreshToken string
	CookieSecure bool
}

// Login checks the password and opens a refresh session.
func (s *Service) Login(ctx context.Context, email string, password string) (Session, error) {
	if err := s.ready(); err != nil {
		return Session{}, err
	}
	normalized, err := NormalizeEmail(email)
	if err != nil || ValidatePassword(password) != nil {
		_ = VerifyPassword(s.dummyHash(), password)
		return Session{}, ErrInvalidCredentials
	}
	var (
		id     string
		hash   string
		status string
	)
	err = s.pool.QueryRow(ctx, `
		SELECT id::text, password_hash, status
		FROM users
		WHERE email_normalized = $1
	`, normalized).Scan(&id, &hash, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = VerifyPassword(s.dummyHash(), password)
		return Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if !VerifyPassword(hash, password) || status != "ACTIVE" {
		return Session{}, ErrInvalidCredentials
	}
	return s.startSession(ctx, Account{ID: id, Email: normalized})
}

func (s *Service) startSession(ctx context.Context, account Account) (Session, error) {
	raw, hashed, err := newSecretToken()
	if err != nil {
		return Session{}, err
	}
	now := s.now().UTC()
	access, err := IssueAccessToken(s.secret, account, now)
	if err != nil {
		return Session{}, err
	}
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO refresh_sessions (id, user_id, token_hash, family_id, expires_at, created_at)
		VALUES (gen_random_uuid(), $1, $2, gen_random_uuid(), $3, $4)
	`, account.ID, hashed, now.Add(refreshTTL), now); err != nil {
		return Session{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	return Session{
		Account:      account,
		AccessToken:  access,
		ExpiresIn:    accessExpiresIn,
		RefreshToken: raw,
		CookieSecure: s.secure,
	}, nil
}

func (s *Service) dummyHash() string {
	s.dummyOnce.Do(func() {
		s.dummy, _ = HashPassword("not-a-real-password")
	})
	return s.dummy
}
