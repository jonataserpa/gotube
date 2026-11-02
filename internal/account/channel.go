package account

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var handlePattern = regexp.MustCompile(`^[a-z0-9-]{3,30}$`)

// Category is an active label that can be shown with a channel.
type Category struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// Channel is the public metadata of one channel.
type Channel struct {
	ID          string     `json:"id"`
	Handle      string     `json:"handle"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Categories  []Category `json:"categories"`
}

// OwnChannel states whether the authenticated account has a channel.
type OwnChannel struct {
	Channel *Channel `json:"channel"`
}

// ChannelInput is the editable public text of a new channel.
type ChannelInput struct {
	Handle      string
	Name        string
	Description string
	CategoryIDs []string
}

// ChannelPatch updates only the fields the client sent.
type ChannelPatch struct {
	Handle      *string
	Name        *string
	Description *string
	CategoryIDs *[]string
}

// ReadOwnChannel returns the account channel or an explicit empty result.
func (s *Service) ReadOwnChannel(ctx context.Context, rawToken string) (OwnChannel, error) {
	account, err := s.authenticate(rawToken)
	if err != nil {
		return OwnChannel{}, err
	}
	if s.pool == nil {
		return OwnChannel{}, ErrUnavailable
	}
	channel, err := s.channelByOwner(ctx, account.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return OwnChannel{}, nil
	}
	if err != nil {
		return OwnChannel{}, err
	}
	return OwnChannel{Channel: &channel}, nil
}

// CreateChannel inserts the single channel owned by the token account.
func (s *Service) CreateChannel(ctx context.Context, rawToken string, input ChannelInput) (Channel, error) {
	account, err := s.authenticate(rawToken)
	if err != nil {
		return Channel{}, err
	}
	normalized, err := normalizeChannel(input)
	if err != nil {
		return Channel{}, err
	}
	if s.pool == nil {
		return Channel{}, ErrUnavailable
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Channel{}, errors.Join(ErrUnavailable, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := assertActiveCategories(ctx, tx, normalized.CategoryIDs); err != nil {
		return Channel{}, err
	}
	var channel Channel
	err = tx.QueryRow(ctx, `
		INSERT INTO channels (id, handle, name, description, owner_id)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
		RETURNING id::text, handle, name, description
	`, normalized.Handle, normalized.Name, normalized.Description, account.ID).Scan(
		&channel.ID, &channel.Handle, &channel.Name, &channel.Description,
	)
	if err != nil {
		return Channel{}, mapChannelWrite(err)
	}
	if err := replaceCategories(ctx, tx, channel.ID, normalized.CategoryIDs); err != nil {
		return Channel{}, err
	}
	channel.Categories, err = loadCategories(ctx, tx, channel.ID)
	if err != nil {
		return Channel{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Channel{}, errors.Join(ErrUnavailable, err)
	}
	return channel, nil
}

// UpdateChannel changes the channel only when the token account owns it.
func (s *Service) UpdateChannel(ctx context.Context, rawToken string, id string, patch ChannelPatch) (Channel, error) {
	account, err := s.authenticate(rawToken)
	if err != nil {
		return Channel{}, err
	}
	if s.pool == nil {
		return Channel{}, ErrUnavailable
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Channel{}, errors.Join(ErrUnavailable, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var owner *string
	var current Channel
	err = tx.QueryRow(ctx, `
		SELECT owner_id::text, id::text, handle, name, description
		FROM channels
		WHERE id = $1
	`, id).Scan(&owner, &current.ID, &current.Handle, &current.Name, &current.Description)
	if errors.Is(err, pgx.ErrNoRows) {
		return Channel{}, ErrNotFound
	}
	if err != nil {
		if isInvalidUUID(err) {
			return Channel{}, ErrNotFound
		}
		return Channel{}, errors.Join(ErrUnavailable, err)
	}
	if owner == nil || *owner != account.ID {
		return Channel{}, ErrForbidden
	}
	next := ChannelInput{Handle: current.Handle, Name: current.Name, Description: current.Description}
	if patch.Handle != nil {
		next.Handle = *patch.Handle
	}
	if patch.Name != nil {
		next.Name = *patch.Name
	}
	if patch.Description != nil {
		next.Description = *patch.Description
	}
	normalized, err := normalizeChannel(next)
	if err != nil {
		return Channel{}, err
	}
	var channel Channel
	err = tx.QueryRow(ctx, `
		UPDATE channels
		SET handle = $2, name = $3, description = $4
		WHERE id = $1 AND owner_id = $5
		RETURNING id::text, handle, name, description
	`, id, normalized.Handle, normalized.Name, normalized.Description, account.ID).Scan(
		&channel.ID, &channel.Handle, &channel.Name, &channel.Description,
	)
	if err != nil {
		return Channel{}, mapChannelWrite(err)
	}
	if patch.CategoryIDs != nil {
		if err := assertActiveCategories(ctx, tx, *patch.CategoryIDs); err != nil {
			return Channel{}, err
		}
		if err := replaceCategories(ctx, tx, channel.ID, *patch.CategoryIDs); err != nil {
			return Channel{}, err
		}
	}
	channel.Categories, err = loadCategories(ctx, tx, channel.ID)
	if err != nil {
		return Channel{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Channel{}, errors.Join(ErrUnavailable, err)
	}
	return channel, nil
}

// ReadPublicChannel returns channel metadata without account data.
func (s *Service) ReadPublicChannel(ctx context.Context, handle string) (Channel, error) {
	if s.pool == nil {
		return Channel{}, ErrUnavailable
	}
	normalized := strings.ToLower(strings.TrimSpace(handle))
	var channel Channel
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, handle, name, description
		FROM channels
		WHERE handle = $1
	`, normalized).Scan(&channel.ID, &channel.Handle, &channel.Name, &channel.Description)
	if errors.Is(err, pgx.ErrNoRows) {
		return Channel{}, ErrNotFound
	}
	if err != nil {
		return Channel{}, errors.Join(ErrUnavailable, err)
	}
	channel.Categories, err = loadCategories(ctx, s.pool, channel.ID)
	if err != nil {
		return Channel{}, err
	}
	return channel, nil
}

func (s *Service) channelByOwner(ctx context.Context, ownerID string) (Channel, error) {
	var channel Channel
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, handle, name, description
		FROM channels
		WHERE owner_id = $1
	`, ownerID).Scan(&channel.ID, &channel.Handle, &channel.Name, &channel.Description)
	if err != nil {
		return Channel{}, err
	}
	channel.Categories, err = loadCategories(ctx, s.pool, channel.ID)
	if err != nil {
		return Channel{}, err
	}
	return channel, nil
}

func normalizeChannel(input ChannelInput) (ChannelInput, error) {
	handle := strings.ToLower(strings.TrimSpace(input.Handle))
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)
	if !handlePattern.MatchString(handle) || strings.Trim(handle, "-") == "" {
		return ChannelInput{}, ErrValidation
	}
	if name == "" || utf8.RuneCountInString(name) > 80 || utf8.RuneCountInString(description) > 500 {
		return ChannelInput{}, ErrValidation
	}
	return ChannelInput{Handle: handle, Name: name, Description: description, CategoryIDs: input.CategoryIDs}, nil
}

func mapChannelWrite(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "owner") {
			return ErrChannelExists
		}
		return ErrHandleTaken
	}
	if isInvalidUUID(err) {
		return ErrNotFound
	}
	return errors.Join(ErrUnavailable, err)
}

func isInvalidUUID(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "22P02"
}
