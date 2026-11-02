package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

const raceWindow = 10 * time.Second

// Refresh rotates a live refresh token. A token already rotated inside the
// race window is rejected without revoking the successor. After that window,
// presenting it revokes the family.
func (s *Service) Refresh(ctx context.Context, raw string) (Session, error) {
	if err := s.ready(); err != nil {
		return Session{}, err
	}
	if raw == "" {
		return Session{}, ErrInvalidRefresh
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Session{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		id       string
		userID   string
		familyID string
		expires  time.Time
		revoked  *time.Time
	)
	err = tx.QueryRow(ctx, `
		SELECT id::text, user_id::text, family_id::text, expires_at, revoked_at
		FROM refresh_sessions
		WHERE token_hash = $1
		FOR UPDATE
	`, hashToken(raw)).Scan(&id, &userID, &familyID, &expires, &revoked)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrInvalidRefresh
	}
	if err != nil {
		return Session{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if revoked != nil {
		if err := s.reuse(ctx, tx, id, familyID); err != nil {
			return Session{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return Session{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
		}
		return Session{}, ErrInvalidRefresh
	}
	if !expires.After(s.now()) {
		return Session{}, ErrInvalidRefresh
	}

	var email, status string
	if err := tx.QueryRow(ctx, `
		SELECT email_normalized, status FROM users WHERE id = $1
	`, userID).Scan(&email, &status); err != nil {
		return Session{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if status != "ACTIVE" {
		return Session{}, ErrInvalidRefresh
	}
	session, err := s.rotate(ctx, tx, Account{ID: userID, Email: email}, familyID, id)
	if err != nil {
		return Session{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	return session, nil
}

func (s *Service) reuse(ctx context.Context, tx pgx.Tx, sessionID string, familyID string) error {
	var successor time.Time
	err := tx.QueryRow(ctx, `
		SELECT created_at
		FROM refresh_sessions
		WHERE family_id = $1 AND id::text <> $2
		ORDER BY created_at DESC
		LIMIT 1
	`, familyID, sessionID).Scan(&successor)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if err == nil && s.now().Sub(successor) <= raceWindow {
		return nil
	}
	if _, err := tx.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = $2
		WHERE family_id = $1 AND revoked_at IS NULL
	`, familyID, s.now().UTC()); err != nil {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	return nil
}

func (s *Service) rotate(ctx context.Context, tx pgx.Tx, account Account, familyID string, currentID string) (Session, error) {
	raw, hashed, err := newSecretToken()
	if err != nil {
		return Session{}, err
	}
	now := s.now().UTC()
	access, err := IssueAccessToken(s.secret, account, now)
	if err != nil {
		return Session{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE refresh_sessions SET revoked_at = $2 WHERE id = $1
	`, currentID, now); err != nil {
		return Session{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO refresh_sessions (id, user_id, token_hash, family_id, expires_at, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
	`, account.ID, hashed, familyID, now.Add(refreshTTL), now); err != nil {
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

// Logout revokes only the presented refresh session.
func (s *Service) Logout(ctx context.Context, raw string) error {
	if err := s.ready(); err != nil {
		return err
	}
	if raw == "" {
		return nil
	}
	if _, err := s.pool.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = $2
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, hashToken(raw), s.now().UTC()); err != nil {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	return nil
}

// SecureCookie reports whether the refresh cookie must set Secure.
func (s *Service) SecureCookie() bool {
	return s.secure
}
