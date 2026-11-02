package processing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"

	"github.com/jackc/pgx/v5"
	"time"
)

type claimedJob struct {
	id         string
	videoID    string
	generation int
	token      int64
	attempts   int
	owned      bool
	busy       bool
}

func (w *Worker) claim(ctx context.Context, message uploadedMessage) (claimedJob, error) {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return claimedJob{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	job := claimedJob{videoID: message.VideoID, generation: message.Generation}
	var status string
	var leaseUntil *time.Time
	scanErr := tx.QueryRow(ctx, `
		SELECT id::text, status, fencing_token, attempts, lease_until
		FROM processing_jobs
		WHERE video_id = $1 AND generation = $2
		FOR UPDATE
	`, message.VideoID, message.Generation).Scan(&job.id, &status, &job.token, &job.attempts, &leaseUntil)
	if scanErr != nil && !errors.Is(scanErr, pgx.ErrNoRows) {
		return claimedJob{}, scanErr
	}
	if status == "READY" || status == "FAILED" {
		return job, nil
	}
	if status == "PROCESSING" && leaseUntil != nil && leaseUntil.After(time.Now()) {
		job.busy = true
		return job, nil
	}
	seconds := int(w.lease.Seconds())
	if errors.Is(scanErr, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, `
			INSERT INTO processing_jobs (id, video_id, generation, status, lease_until, fencing_token, attempts)
			VALUES (gen_random_uuid(), $1, $2, 'PROCESSING', now() + make_interval(secs => $3), 1, 1)
			RETURNING id::text, fencing_token, attempts
		`, message.VideoID, message.Generation, seconds).Scan(&job.id, &job.token, &job.attempts)
	} else {
		err = tx.QueryRow(ctx, `
			UPDATE processing_jobs
			SET status = 'PROCESSING',
				lease_until = now() + make_interval(secs => $2),
				fencing_token = fencing_token + 1,
				attempts = attempts + 1
			WHERE id = $1
			RETURNING id::text, fencing_token, attempts
		`, job.id, seconds).Scan(&job.id, &job.token, &job.attempts)
	}
	if err != nil {
		return claimedJob{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE videos
		SET processing_status = 'PROCESSING'
		WHERE id = $1 AND generation = $2 AND processing_status IN ('QUEUED', 'PROCESSING', 'FAILED')
	`, message.VideoID, message.Generation); err != nil {
		return claimedJob{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return claimedJob{}, err
	}
	job.owned = true
	return job, nil
}

func (w *Worker) release(ctx context.Context, job claimedJob) error {
	_, err := w.pool.Exec(ctx, `
		UPDATE processing_jobs
		SET status = 'QUEUED', lease_until = NULL
		WHERE id = $1 AND fencing_token = $2 AND status = 'PROCESSING'
	`, job.id, job.token)
	return err
}

func (w *Worker) fail(ctx context.Context, job claimedJob, code string, message string) error {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `
		UPDATE processing_jobs
		SET status = 'FAILED', lease_until = NULL, error_code = $3, error_message = $4
		WHERE id = $1 AND fencing_token = $2 AND status = 'PROCESSING'
	`, job.id, job.token, code, message)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return nil
	}
	if _, err := tx.Exec(ctx, `
		UPDATE videos
		SET processing_status = 'FAILED'
		WHERE id = $1 AND generation = $2 AND processing_status <> 'READY'
	`, job.videoID, job.generation); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type storedAsset struct {
	kind        string
	key         string
	contentType string
	path        string
	checksum    string
}

func (w *Worker) finish(ctx context.Context, job claimedJob, duration int, assets []storedAsset) (bool, error) {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `
		UPDATE processing_jobs
		SET status = 'READY', lease_until = NULL, error_code = NULL, error_message = NULL
		WHERE id = $1 AND fencing_token = $2 AND generation = $3 AND status = 'PROCESSING'
	`, job.id, job.token, job.generation)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() != 1 {
		return false, nil
	}
	var generation int
	if err := tx.QueryRow(ctx, `SELECT generation FROM videos WHERE id = $1 FOR UPDATE`, job.videoID).Scan(&generation); err != nil {
		return false, err
	}
	if generation != job.generation {
		return false, nil
	}
	for _, asset := range assets {
		if _, err := tx.Exec(ctx, `
			INSERT INTO video_assets (id, video_id, generation, kind, object_key, checksum)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
		`, job.videoID, job.generation, asset.kind, asset.key, asset.checksum); err != nil {
			return false, err
		}
	}
	tag, err = tx.Exec(ctx, `
		UPDATE videos
		SET processing_status = 'READY', duration_seconds = $3
		WHERE id = $1 AND generation = $2
			AND publication_status = 'DRAFT' AND visibility = 'PRIVATE'
	`, job.videoID, job.generation, duration)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() != 1 {
		return false, nil
	}
	payload, err := json.Marshal(map[string]any{"video_id": job.videoID, "generation": job.generation})
	if err != nil {
		return false, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO outbox (id, aggregate_id, type, version, payload)
		VALUES (gen_random_uuid(), $1, 'video.ready.v1', 1, $2)
	`, job.videoID, payload); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func fileChecksum(path string) (string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}
