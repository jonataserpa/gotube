package processing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// OutcomeAck confirms the delivery. The same body must not be retried.
	OutcomeAck = Outcome(iota)
	// OutcomeRetry returns the delivery to the queue.
	OutcomeRetry
)

// Outcome tells the consumer whether the message is durable.
type Outcome int

// Objects moves bytes between the bucket and a local file.
type Objects interface {
	Download(ctx context.Context, key string, dest string) error
	Upload(ctx context.Context, key string, path string, contentType string) error
}

// Probe is the inspection the worker is allowed to keep.
type Probe struct {
	DurationSeconds int
	Height          int
	HasAudio        bool
	HasVideo        bool
}

// Runner inspects and renders one source file. Implementations must not join arguments into a shell.
type Runner interface {
	Probe(ctx context.Context, sourcePath string) (Probe, error)
	Render(ctx context.Context, sourcePath string, outputDir string, probe Probe) error
}

// Worker turns one uploaded object into a thumbnail and one HLS rendition.
type Worker struct {
	pool        *pgxpool.Pool
	objects     Objects
	media       Runner
	lease       time.Duration
	maxAttempts int
}

// NewWorker binds the pipeline. maxAttempts counts runs of the same generation.
func NewWorker(pool *pgxpool.Pool, objects Objects, media Runner, lease time.Duration, maxAttempts int) (*Worker, error) {
	if pool == nil || objects == nil || media == nil {
		return nil, errors.New("banco, objetos e mídia são obrigatórios")
	}
	if lease <= 0 {
		lease = 2 * time.Minute
	}
	if maxAttempts < 1 {
		maxAttempts = 3
	}
	return &Worker{pool: pool, objects: objects, media: media, lease: lease, maxAttempts: maxAttempts}, nil
}

type uploadedMessage struct {
	VideoID         string `json:"video_id"`
	Generation      int    `json:"generation"`
	UploadSessionID string `json:"upload_session_id"`
}

type videoRow struct {
	generation  int
	processing  string
	publication string
	visibility  string
}

// Handle applies one delivery. Invalid JSON is acknowledged and not retried.
func (w *Worker) Handle(ctx context.Context, body []byte) (Outcome, error) {
	var message uploadedMessage
	if err := json.Unmarshal(body, &message); err != nil || message.VideoID == "" || message.Generation < 1 || message.UploadSessionID == "" {
		return OutcomeAck, nil
	}
	video, err := w.loadVideo(ctx, message.VideoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return OutcomeAck, nil
	}
	if err != nil {
		return OutcomeRetry, err
	}
	if video.generation != message.Generation || video.processing == "READY" {
		return OutcomeAck, nil
	}
	job, err := w.claim(ctx, message)
	if err != nil {
		return OutcomeRetry, err
	}
	if !job.owned {
		if job.busy {
			return OutcomeRetry, nil
		}
		return OutcomeAck, nil
	}
	applied, err := w.pipeline(ctx, message, job)
	if err != nil {
		return w.failOrRetry(ctx, job, err)
	}
	if !applied {
		return OutcomeRetry, nil
	}
	return OutcomeAck, nil
}

func (w *Worker) loadVideo(ctx context.Context, videoID string) (videoRow, error) {
	var video videoRow
	err := w.pool.QueryRow(ctx, `
		SELECT generation, processing_status, publication_status, visibility
		FROM videos
		WHERE id = $1
	`, videoID).Scan(&video.generation, &video.processing, &video.publication, &video.visibility)
	if err != nil {
		return videoRow{}, err
	}
	return video, nil
}

func (w *Worker) failOrRetry(ctx context.Context, job claimedJob, cause error) (Outcome, error) {
	if !permanent(cause) && job.attempts < w.maxAttempts {
		if err := w.release(ctx, job); err != nil {
			return OutcomeRetry, err
		}
		return OutcomeRetry, nil
	}
	code, message := safeFailure(cause)
	if err := w.fail(ctx, job, code, message); err != nil {
		return OutcomeRetry, err
	}
	return OutcomeAck, nil
}

func permanent(err error) bool {
	return errors.Is(err, errNoVideo) || errors.Is(err, errInvalidOutput)
}

func safeFailure(err error) (string, string) {
	if errors.Is(err, errNoVideo) {
		return "MEDIA_INVALID", "o arquivo não tem vídeo"
	}
	return "PROCESSING_FAILED", "não foi possível processar o vídeo"
}

var (
	errNoVideo       = errors.New("sem vídeo")
	errInvalidOutput = errors.New("saída inválida")
)

// RecoverExpired returns an abandoned PROCESSING lease to the queue, or fails it at the attempt cap.
func (w *Worker) RecoverExpired(ctx context.Context) error {
	if _, err := w.pool.Exec(ctx, `
		WITH expired AS (
			UPDATE processing_jobs
			SET status = 'FAILED',
				lease_until = NULL,
				error_code = 'PROCESSING_FAILED',
				error_message = 'não foi possível processar o vídeo'
			WHERE status = 'PROCESSING'
				AND lease_until IS NOT NULL
				AND lease_until <= now()
				AND attempts >= $1
			RETURNING video_id, generation
		)
		UPDATE videos
		SET processing_status = 'FAILED'
		FROM expired
		WHERE videos.id = expired.video_id
			AND videos.generation = expired.generation
			AND videos.processing_status <> 'READY'
	`, w.maxAttempts); err != nil {
		return fmt.Errorf("encerrar lease: %w", err)
	}
	if _, err := w.pool.Exec(ctx, `
		UPDATE processing_jobs
		SET status = 'QUEUED', lease_until = NULL
		WHERE status = 'PROCESSING'
			AND lease_until IS NOT NULL
			AND lease_until <= now()
			AND attempts < $1
	`, w.maxAttempts); err != nil {
		return fmt.Errorf("devolver lease: %w", err)
	}
	return nil
}
