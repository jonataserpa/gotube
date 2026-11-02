package identity

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const resetTTL = 30 * time.Minute

// Forgot stores a single-use reset token when the account exists.
// The raw token is returned only when laboratory exposure is enabled.
func (s *Service) Forgot(ctx context.Context, email string) (string, error) {
	if err := s.ready(); err != nil {
		return "", err
	}
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return "", nil
	}
	var userID string
	err = s.pool.QueryRow(ctx, `
		SELECT id::text FROM users WHERE email_normalized = $1
	`, normalized).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	raw, hashed, err := newSecretToken()
	if err != nil {
		return "", err
	}
	now := s.now().UTC()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
		UPDATE password_resets
		SET used_at = $2
		WHERE user_id = $1 AND used_at IS NULL
	`, userID, now); err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO password_resets (id, user_id, token_hash, expires_at)
		VALUES (gen_random_uuid(), $1, $2, $3)
	`, userID, hashed, now.Add(resetTTL)); err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if !s.expose {
		return "", nil
	}
	return raw, nil
}

// Reset consumes a live token, replaces the password and revokes every refresh session.
func (s *Service) Reset(ctx context.Context, token string, password string) error {
	if err := s.ready(); err != nil {
		return err
	}
	if strings.TrimSpace(token) == "" {
		return ErrInvalidReset
	}
	if err := ValidatePassword(password); err != nil {
		return err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		resetID string
		userID  string
		expires time.Time
		used    *time.Time
	)
	err = tx.QueryRow(ctx, `
		SELECT id::text, user_id::text, expires_at, used_at
		FROM password_resets
		WHERE token_hash = $1
		FOR UPDATE
	`, hashToken(token)).Scan(&resetID, &userID, &expires, &used)
	if errors.Is(err, pgx.ErrNoRows) || used != nil || (err == nil && !expires.After(s.now())) {
		return ErrInvalidReset
	}
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	now := s.now().UTC()
	if _, err := tx.Exec(ctx, `UPDATE users SET password_hash = $2 WHERE id = $1`, userID, hash); err != nil {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if _, err := tx.Exec(ctx, `UPDATE password_resets SET used_at = $2 WHERE id = $1`, resetID, now); err != nil {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID, now); err != nil {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	return nil
}
