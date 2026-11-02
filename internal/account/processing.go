package account

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ProcessingError is the safe failure the owner can read.
type ProcessingError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ProcessingView is the owner's processing state. It never includes an object key.
type ProcessingView struct {
	ProcessingStatus string           `json:"processing_status"`
	Error            *ProcessingError `json:"error,omitempty"`
}

// ReadProcessing returns the processing state of a video owned by the token account.
func (s *Service) ReadProcessing(ctx context.Context, rawToken string, videoID string) (ProcessingView, error) {
	owner, err := s.authenticate(rawToken)
	if err != nil {
		return ProcessingView{}, err
	}
	if s.pool == nil {
		return ProcessingView{}, ErrUnavailable
	}
	var view ProcessingView
	var channelOwner *string
	var code, message *string
	err = s.pool.QueryRow(ctx, `
		SELECT v.processing_status, c.owner_id::text, j.error_code, j.error_message
		FROM videos v
		JOIN channels c ON c.id = v.channel_id
		LEFT JOIN processing_jobs j ON j.video_id = v.id AND j.generation = v.generation
		WHERE v.id = $1
	`, videoID).Scan(&view.ProcessingStatus, &channelOwner, &code, &message)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return ProcessingView{}, ErrVideoNotFound
	}
	if err != nil {
		return ProcessingView{}, wrapDB(err)
	}
	if channelOwner == nil || *channelOwner != owner.ID {
		return ProcessingView{}, ErrVideoForbidden
	}
	if view.ProcessingStatus == "FAILED" && code != nil && message != nil && *code != "" && *message != "" {
		view.Error = &ProcessingError{Code: *code, Message: *message}
	}
	return view, nil
}

// RetryProcessing queues a new generation when the owner's video has failed.
func (s *Service) RetryProcessing(ctx context.Context, rawToken string, videoID string) (UploadResult, error) {
	owner, err := s.authenticate(rawToken)
	if err != nil {
		return UploadResult{}, err
	}
	if s.pool == nil {
		return UploadResult{}, ErrUnavailable
	}
	if err := s.authorizeVideo(ctx, owner.ID, videoID); err != nil {
		return UploadResult{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return UploadResult{}, errors.Join(ErrUnavailable, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var generation int
	var sessionID string
	err = tx.QueryRow(ctx, `
		UPDATE videos
		SET processing_status = 'QUEUED', generation = generation + 1
		WHERE id = $1 AND processing_status = 'FAILED'
		RETURNING generation
	`, videoID).Scan(&generation)
	if errors.Is(err, pgx.ErrNoRows) {
		return UploadResult{}, ErrNotFailed
	}
	if err != nil {
		return UploadResult{}, wrapDB(err)
	}
	err = tx.QueryRow(ctx, `
		SELECT id::text
		FROM upload_sessions
		WHERE video_id = $1 AND completed_at IS NOT NULL
		ORDER BY completed_at DESC
		LIMIT 1
	`, videoID).Scan(&sessionID)
	if err != nil {
		return UploadResult{}, wrapDB(err)
	}
	payload, err := json.Marshal(uploadedEvent{VideoID: videoID, Generation: generation, UploadSessionID: sessionID})
	if err != nil {
		return UploadResult{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO outbox (id, aggregate_id, type, version, payload)
		VALUES (gen_random_uuid(), $1, 'video.uploaded.v1', 1, $2)
	`, videoID, payload); err != nil {
		return UploadResult{}, wrapDB(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return UploadResult{}, errors.Join(ErrUnavailable, err)
	}
	return UploadResult{ProcessingStatus: "QUEUED"}, nil
}

func (s *Service) authorizeVideo(ctx context.Context, ownerID string, videoID string) error {
	var channelOwner *string
	err := s.pool.QueryRow(ctx, `
		SELECT c.owner_id::text
		FROM videos v
		JOIN channels c ON c.id = v.channel_id
		WHERE v.id = $1
	`, videoID).Scan(&channelOwner)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return ErrVideoNotFound
	}
	if err != nil {
		return wrapDB(err)
	}
	if channelOwner == nil || *channelOwner != ownerID {
		return ErrVideoForbidden
	}
	return nil
}
