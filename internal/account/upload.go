package account

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const uploadSessionTTL = 15 * time.Minute

// UploadSession is a server-issued permission to PUT one object.
type UploadSession struct {
	ID        string    `json:"id"`
	UploadURL string    `json:"upload_url"`
	ExpiresAt time.Time `json:"expires_at"`
	MaxBytes  int64     `json:"max_bytes"`
}

// UploadResult is the processing state after a confirmed object.
type UploadResult struct {
	ProcessingStatus string `json:"processing_status"`
}

type uploadedEvent struct {
	VideoID         string `json:"video_id"`
	Generation      int    `json:"generation"`
	UploadSessionID string `json:"upload_session_id"`
}

// CreateUploadSession authorizes one PUT for a draft the token account owns.
func (s *Service) CreateUploadSession(ctx context.Context, rawToken string, videoID string, declaredSize int64) (UploadSession, error) {
	owner, err := s.authenticate(rawToken)
	if err != nil {
		return UploadSession{}, err
	}
	if s.pool == nil || s.store == nil || s.maxBytes < 1 {
		return UploadSession{}, ErrUnavailable
	}
	if declaredSize < 1 || declaredSize > s.maxBytes {
		return UploadSession{}, ErrValidation
	}
	if err := s.authorizeWaitingDraft(ctx, owner.ID, videoID); err != nil {
		return UploadSession{}, err
	}
	if _, err := s.pool.Exec(ctx, `
		DELETE FROM upload_sessions
		WHERE video_id = $1 AND completed_at IS NULL AND expires_at <= now()
	`, videoID); err != nil {
		return UploadSession{}, wrapDB(err)
	}
	expiresAt := time.Now().Add(uploadSessionTTL).UTC()
	session, err := s.insertUploadSession(ctx, videoID, owner.ID, declaredSize, expiresAt)
	if err != nil {
		return UploadSession{}, err
	}
	url, err := s.store.PresignPut(ctx, session.key, session.ExpiresAt)
	if err != nil {
		_, _ = s.pool.Exec(ctx, `DELETE FROM upload_sessions WHERE id = $1 AND completed_at IS NULL`, session.ID)
		return UploadSession{}, errors.Join(ErrUnavailable, err)
	}
	session.UploadURL = url
	session.MaxBytes = declaredSize
	return session.UploadSession, nil
}

// CompleteUploadSession checks the object and records one upload event.
func (s *Service) CompleteUploadSession(ctx context.Context, rawToken string, sessionID string) (UploadResult, error) {
	owner, err := s.authenticate(rawToken)
	if err != nil {
		return UploadResult{}, err
	}
	if s.pool == nil || s.store == nil {
		return UploadResult{}, ErrUnavailable
	}
	row, err := s.loadSession(ctx, sessionID)
	if err != nil {
		return UploadResult{}, err
	}
	if row.ownerID != owner.ID {
		return UploadResult{}, ErrVideoForbidden
	}
	if row.completed {
		return UploadResult{ProcessingStatus: "QUEUED"}, nil
	}
	if time.Now().After(row.expiresAt) || row.processing != "WAITING_UPLOAD" {
		return UploadResult{}, ErrUploadRejected
	}
	size, found, err := s.store.Head(ctx, row.key)
	if err != nil {
		return UploadResult{}, errors.Join(ErrUnavailable, err)
	}
	if !found || size < 1 || size > row.declaredSize {
		return UploadResult{}, ErrUploadRejected
	}
	return s.commitUpload(ctx, row)
}

type pendingSession struct {
	UploadSession
	key          string
	ownerID      string
	videoID      string
	declaredSize int64
	expiresAt    time.Time
	processing   string
	completed    bool
}

func (s *Service) authorizeWaitingDraft(ctx context.Context, ownerID string, videoID string) error {
	var channelOwner *string
	var status string
	err := s.pool.QueryRow(ctx, `
		SELECT c.owner_id::text, v.processing_status
		FROM videos v
		JOIN channels c ON c.id = v.channel_id
		WHERE v.id = $1
	`, videoID).Scan(&channelOwner, &status)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return ErrVideoNotFound
	}
	if err != nil {
		return wrapDB(err)
	}
	if channelOwner == nil || *channelOwner != ownerID {
		return ErrVideoForbidden
	}
	if status != "WAITING_UPLOAD" {
		return ErrNotWaiting
	}
	return nil
}

func (s *Service) insertUploadSession(ctx context.Context, videoID string, ownerID string, declaredSize int64, expiresAt time.Time) (pendingSession, error) {
	var session pendingSession
	err := s.pool.QueryRow(ctx, `
		WITH created AS (
			SELECT gen_random_uuid() AS id, $1::uuid AS video_id
		)
		INSERT INTO upload_sessions (id, video_id, owner_id, object_key, declared_size, expires_at)
		SELECT id, video_id, $2::uuid, 'videos/' || video_id::text || '/source/' || id::text, $3, $4
		FROM created
		RETURNING id::text, object_key, expires_at
	`, videoID, ownerID, declaredSize, expiresAt).Scan(&session.ID, &session.key, &session.ExpiresAt)
	if isUniqueViolation(err) {
		return s.reloadOpenSession(ctx, videoID)
	}
	if err != nil {
		return pendingSession{}, wrapDB(err)
	}
	session.declaredSize = declaredSize
	return session, nil
}

func (s *Service) reloadOpenSession(ctx context.Context, videoID string) (pendingSession, error) {
	var session pendingSession
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, object_key, declared_size, expires_at
		FROM upload_sessions
		WHERE video_id = $1 AND completed_at IS NULL AND expires_at > now()
	`, videoID).Scan(&session.ID, &session.key, &session.declaredSize, &session.ExpiresAt)
	if err != nil {
		return pendingSession{}, wrapDB(err)
	}
	return session, nil
}

func (s *Service) loadSession(ctx context.Context, sessionID string) (pendingSession, error) {
	var row pendingSession
	var completedAt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT s.id::text, s.video_id::text, s.owner_id::text, s.object_key, s.declared_size, s.expires_at, s.completed_at, v.processing_status
		FROM upload_sessions s
		JOIN videos v ON v.id = s.video_id
		WHERE s.id = $1
	`, sessionID).Scan(&row.ID, &row.videoID, &row.ownerID, &row.key, &row.declaredSize, &row.expiresAt, &completedAt, &row.processing)
	if errors.Is(err, pgx.ErrNoRows) || isInvalidUUID(err) {
		return pendingSession{}, ErrUploadNotFound
	}
	if err != nil {
		return pendingSession{}, wrapDB(err)
	}
	row.completed = completedAt != nil
	return row, nil
}

func (s *Service) commitUpload(ctx context.Context, row pendingSession) (UploadResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return UploadResult{}, errors.Join(ErrUnavailable, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var completedAt *time.Time
	err = tx.QueryRow(ctx, `SELECT completed_at FROM upload_sessions WHERE id = $1 FOR UPDATE`, row.ID).Scan(&completedAt)
	if err != nil {
		return UploadResult{}, wrapDB(err)
	}
	if completedAt != nil {
		return UploadResult{ProcessingStatus: "QUEUED"}, nil
	}
	payload, err := json.Marshal(uploadedEvent{VideoID: row.videoID, Generation: 1, UploadSessionID: row.ID})
	if err != nil {
		return UploadResult{}, err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE videos SET processing_status = 'QUEUED'
		WHERE id = $1 AND processing_status = 'WAITING_UPLOAD'
	`, row.videoID)
	if err != nil {
		return UploadResult{}, wrapDB(err)
	}
	if tag.RowsAffected() != 1 {
		return UploadResult{}, ErrUploadRejected
	}
	if _, err := tx.Exec(ctx, `UPDATE upload_sessions SET completed_at = now() WHERE id = $1`, row.ID); err != nil {
		return UploadResult{}, wrapDB(err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO outbox (id, aggregate_id, type, version, payload)
		VALUES (gen_random_uuid(), $1, 'video.uploaded.v1', 1, $2)
	`, row.videoID, payload); err != nil {
		return UploadResult{}, wrapDB(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return UploadResult{}, errors.Join(ErrUnavailable, err)
	}
	return UploadResult{ProcessingStatus: "QUEUED"}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
