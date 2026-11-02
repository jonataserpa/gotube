package identity

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Account is the public identity returned after registration or login.
type Account struct {
	ID    string
	Email string
}

// Settings configures signing and laboratory behavior. An empty JWTSecret
// makes every account route report that signing is unavailable.
type Settings struct {
	JWTSecret        string
	CookieSecure     bool
	ExposeResetToken bool
	Now              func() time.Time
}

// Service persists accounts and sessions.
type Service struct {
	pool      *pgxpool.Pool
	secret    string
	secure    bool
	expose    bool
	now       func() time.Time
	dummyOnce sync.Once
	dummy     string
}

// NewService binds identity operations to a PostgreSQL pool.
func NewService(pool *pgxpool.Pool, settings Settings) *Service {
	now := settings.Now
	if now == nil {
		now = time.Now
	}
	return &Service{
		pool:   pool,
		secret: strings.TrimSpace(settings.JWTSecret),
		secure: settings.CookieSecure,
		expose: settings.ExposeResetToken,
		now:    now,
	}
}

// Register stores a new active account and does not open a session.
func (s *Service) Register(ctx context.Context, email string, password string) (Account, error) {
	if err := s.ready(); err != nil {
		return Account{}, err
	}
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return Account{}, err
	}
	if err := ValidatePassword(password); err != nil {
		return Account{}, err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return Account{}, err
	}
	var id string
	err = s.pool.QueryRow(ctx, `
		INSERT INTO users (id, email_normalized, password_hash, status, created_at)
		VALUES (gen_random_uuid(), $1, $2, 'ACTIVE', $3)
		RETURNING id::text
	`, normalized, hash, s.now().UTC()).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Account{}, ErrEmailTaken
		}
		return Account{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	return Account{ID: id, Email: normalized}, nil
}

func (s *Service) ready() error {
	if s == nil || s.pool == nil {
		return ErrUnavailable
	}
	if s.secret == "" {
		return ErrAuthUnavailable
	}
	return nil
}
