package processing

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/gotube/internal/platform/postgres"
)

func openProcessingDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := postgres.OpenCatalogTestDB(t)
	root, err := postgres.ModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := postgres.ApplyMigrations(context.Background(), pool, filepath.Join(root, "db", "migrations")); err != nil {
		t.Fatal(err)
	}
	return pool
}

type queuedVideo struct {
	videoID   string
	sessionID string
	objectKey string
}

func insertQueuedVideo(t *testing.T, pool *pgxpool.Pool) queuedVideo {
	t.Helper()
	ctx := context.Background()
	var ownerID, channelID, videoID, sessionID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (id, email_normalized, password_hash, status, created_at)
		VALUES (gen_random_uuid(), 'owner-' || gen_random_uuid()::text || '@example.com', 'hash', 'ACTIVE', now())
		RETURNING id::text
	`).Scan(&ownerID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO channels (id, handle, name, owner_id)
		VALUES (gen_random_uuid(), 'canal-' || left(gen_random_uuid()::text, 8), 'Canal', $1)
		RETURNING id::text
	`, ownerID).Scan(&channelID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO videos (id, channel_id, title, duration_seconds, visibility, publication_status, processing_status, generation)
		VALUES (gen_random_uuid(), $1, 'Aula', 0, 'PRIVATE', 'DRAFT', 'QUEUED', 1)
		RETURNING id::text
	`, channelID).Scan(&videoID); err != nil {
		t.Fatal(err)
	}
	objectKey := "videos/" + videoID + "/source/session"
	if err := pool.QueryRow(ctx, `
		INSERT INTO upload_sessions (id, video_id, owner_id, object_key, declared_size, expires_at, completed_at)
		VALUES (gen_random_uuid(), $1, $2, $3, 100, now() + interval '15 minutes', now())
		RETURNING id::text
	`, videoID, ownerID, objectKey).Scan(&sessionID); err != nil {
		t.Fatal(err)
	}
	return queuedVideo{videoID: videoID, sessionID: sessionID, objectKey: objectKey}
}

func countQuery(t *testing.T, pool *pgxpool.Pool, query string, args ...any) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func videoFields(t *testing.T, pool *pgxpool.Pool, videoID string) (string, string, string, int, int) {
	t.Helper()
	var processing, publication, visibility string
	var duration, generation int
	err := pool.QueryRow(context.Background(), `
		SELECT processing_status, publication_status, visibility, duration_seconds, generation
		FROM videos WHERE id = $1
	`, videoID).Scan(&processing, &publication, &visibility, &duration, &generation)
	if err != nil {
		t.Fatal(err)
	}
	return processing, publication, visibility, duration, generation
}
