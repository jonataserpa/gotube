package account

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/gotube/internal/identity"
)

// Service reads and updates the authenticated account's public profile and channel.
type Service struct {
	pool     *pgxpool.Pool
	secret   string
	store    ObjectStore
	maxBytes int64
}

// NewService binds profile and channel commands to the account database.
func NewService(pool *pgxpool.Pool, jwtSecret string) *Service {
	return &Service{pool: pool, secret: jwtSecret}
}

// Profile is the account the access token belongs to.
type Profile struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
}

// ReadProfile returns the account named by a valid access token.
func (s *Service) ReadProfile(ctx context.Context, rawToken string) (Profile, error) {
	account, err := s.authenticate(rawToken)
	if err != nil {
		return Profile{}, err
	}
	return s.profileByID(ctx, account.ID)
}

// UpdateProfile stores a trimmed display name of 1 to 80 runes.
func (s *Service) UpdateProfile(ctx context.Context, rawToken string, displayName string) (Profile, error) {
	account, err := s.authenticate(rawToken)
	if err != nil {
		return Profile{}, err
	}
	name, err := normalizeDisplayName(displayName)
	if err != nil {
		return Profile{}, err
	}
	if s.pool == nil {
		return Profile{}, ErrUnavailable
	}
	var profile Profile
	err = s.pool.QueryRow(ctx, `
		UPDATE users
		SET display_name = $2
		WHERE id = $1
		RETURNING id::text, email_normalized, display_name
	`, account.ID, name).Scan(&profile.ID, &profile.Email, &profile.DisplayName)
	if err != nil {
		return Profile{}, wrapDB(err)
	}
	return profile, nil
}

func (s *Service) profileByID(ctx context.Context, id string) (Profile, error) {
	if s.pool == nil {
		return Profile{}, ErrUnavailable
	}
	var profile Profile
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, email_normalized, display_name
		FROM users
		WHERE id = $1
	`, id).Scan(&profile.ID, &profile.Email, &profile.DisplayName)
	if err != nil {
		return Profile{}, wrapDB(err)
	}
	return profile, nil
}

func (s *Service) authenticate(rawToken string) (identity.Account, error) {
	if strings.TrimSpace(rawToken) == "" || s.secret == "" {
		return identity.Account{}, ErrUnauthenticated
	}
	account, err := identity.ParseAccessToken(s.secret, rawToken)
	if err != nil {
		return identity.Account{}, ErrUnauthenticated
	}
	return account, nil
}

func normalizeDisplayName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	count := utf8.RuneCountInString(name)
	if count < 1 || count > 80 {
		return "", ErrValidation
	}
	return name, nil
}

func wrapDB(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUnauthenticated
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return err
	}
	return errors.Join(ErrUnavailable, err)
}
