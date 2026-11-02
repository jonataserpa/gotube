package account

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

const maxDraftTitleRunes = 120

// Draft is a private video that is not yet in the public catalog.
type Draft struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	DurationSeconds   int    `json:"duration_seconds"`
	Visibility        string `json:"visibility"`
	PublicationStatus string `json:"publication_status"`
	ProcessingStatus  string `json:"processing_status"`
}

// CreateDraft inserts a private unpublished video on the token account's channel.
func (s *Service) CreateDraft(ctx context.Context, rawToken string, title string) (Draft, error) {
	owner, err := s.authenticate(rawToken)
	if err != nil {
		return Draft{}, err
	}
	title = strings.TrimSpace(title)
	if title == "" || utf8.RuneCountInString(title) > maxDraftTitleRunes {
		return Draft{}, ErrValidation
	}
	if s.pool == nil {
		return Draft{}, ErrUnavailable
	}
	var draft Draft
	err = s.pool.QueryRow(ctx, `
		INSERT INTO videos (id, channel_id, title, duration_seconds, visibility, publication_status, processing_status)
		SELECT gen_random_uuid(), c.id, $1, 0, 'PRIVATE', 'DRAFT', 'WAITING_UPLOAD'
		FROM channels c
		WHERE c.owner_id = $2
		RETURNING id::text, title, duration_seconds, visibility, publication_status, processing_status
	`, title, owner.ID).Scan(
		&draft.ID, &draft.Title, &draft.DurationSeconds, &draft.Visibility, &draft.PublicationStatus, &draft.ProcessingStatus,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Draft{}, ErrChannelRequired
	}
	if err != nil {
		return Draft{}, wrapDB(err)
	}
	return draft, nil
}
