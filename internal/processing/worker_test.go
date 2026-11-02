package processing

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/gotube/internal/platform/postgres"
)

func TestWorkerStoresOneRenditionWithoutPublishing(t *testing.T) {
	pool := openProcessingDB(t)
	video := insertQueuedVideo(t, pool)
	objects := newMemObjects(video.objectKey)
	worker := newTestWorker(t, pool, objects, fakeMedia{probe: Probe{DurationSeconds: 12, Height: 720, HasVideo: true, HasAudio: true}}, 3)

	outcome, err := worker.Handle(context.Background(), message(video, 1))
	if err != nil || outcome != OutcomeAck {
		t.Fatalf("outcome=%d err=%v", outcome, err)
	}
	processing, publication, visibility, duration, generation := videoFields(t, pool, video.videoID)
	if processing != "READY" || publication != "DRAFT" || visibility != "PRIVATE" || duration != 12 || generation != 1 {
		t.Fatalf("state=%s publication=%s visibility=%s duration=%d generation=%d", processing, publication, visibility, duration, generation)
	}
	if countQuery(t, pool, `SELECT count(*) FROM video_assets WHERE video_id = $1 AND generation = 1`, video.videoID) != 3 {
		t.Fatal("expected playlist, segment and thumbnail")
	}
	if countQuery(t, pool, `SELECT count(*) FROM outbox WHERE aggregate_id = $1 AND type = 'video.ready.v1'`, video.videoID) != 1 {
		t.Fatal("expected one ready event")
	}
	if !strings.HasPrefix(objects.order[0], "videos/"+video.videoID+"/generations/1/hls/segment") || !strings.HasSuffix(objects.order[1], "index.m3u8") || !strings.HasSuffix(objects.order[2], "thumbnail.jpg") {
		t.Fatalf("upload order = %v", objects.order)
	}

	again, err := worker.Handle(context.Background(), message(video, 1))
	if err != nil || again != OutcomeAck {
		t.Fatalf("redelivery outcome=%d err=%v", again, err)
	}
	if countQuery(t, pool, `SELECT count(*) FROM video_assets WHERE video_id = $1`, video.videoID) != 3 {
		t.Fatal("redelivery created another asset")
	}
	if countQuery(t, pool, `SELECT count(*) FROM outbox WHERE aggregate_id = $1 AND type = 'video.ready.v1'`, video.videoID) != 1 {
		t.Fatal("redelivery created another ready event")
	}
}

func TestWorkerFailsWhenProbeFindsNoVideo(t *testing.T) {
	pool := openProcessingDB(t)
	video := insertQueuedVideo(t, pool)
	objects := newMemObjects(video.objectKey)
	worker := newTestWorker(t, pool, objects, fakeMedia{probe: Probe{HasVideo: false}}, 3)

	outcome, err := worker.Handle(context.Background(), message(video, 1))
	if err != nil || outcome != OutcomeAck {
		t.Fatalf("outcome=%d err=%v", outcome, err)
	}
	processing, _, _, _, _ := videoFields(t, pool, video.videoID)
	if processing != "FAILED" {
		t.Fatalf("status = %s", processing)
	}
	if len(objects.order) != 0 || countQuery(t, pool, `SELECT count(*) FROM video_assets WHERE video_id = $1`, video.videoID) != 0 {
		t.Fatal("playlist was stored without a video track")
	}
	assertSafeError(t, pool, video.videoID)
}

func TestStaleFenceDoesNotChangeTheVideo(t *testing.T) {
	pool := openProcessingDB(t)
	video := insertQueuedVideo(t, pool)
	objects := newMemObjects(video.objectKey)
	media := fakeMedia{probe: Probe{DurationSeconds: 4, Height: 240, HasVideo: true}, before: func() {
		if _, err := pool.Exec(context.Background(), `UPDATE processing_jobs SET fencing_token = fencing_token + 1 WHERE video_id = $1`, video.videoID); err != nil {
			t.Fatal(err)
		}
	}}
	worker := newTestWorker(t, pool, objects, media, 3)
	outcome, err := worker.Handle(context.Background(), message(video, 1))
	if err != nil || outcome != OutcomeRetry {
		t.Fatalf("outcome=%d err=%v", outcome, err)
	}
	processing, publication, _, _, generation := videoFields(t, pool, video.videoID)
	if processing == "READY" || publication != "DRAFT" || generation != 1 {
		t.Fatalf("status=%s publication=%s generation=%d", processing, publication, generation)
	}
	if countQuery(t, pool, `SELECT count(*) FROM video_assets WHERE video_id = $1`, video.videoID) != 0 {
		t.Fatal("stale fence stored assets")
	}
	if countQuery(t, pool, `SELECT count(*) FROM outbox WHERE aggregate_id = $1 AND type = 'video.ready.v1'`, video.videoID) != 0 {
		t.Fatal("stale fence stored a ready event")
	}
}

func TestOldGenerationDoesNotChangeTheVideo(t *testing.T) {
	pool := openProcessingDB(t)
	video := insertQueuedVideo(t, pool)
	if _, err := pool.Exec(context.Background(), `UPDATE videos SET generation = 2 WHERE id = $1`, video.videoID); err != nil {
		t.Fatal(err)
	}
	objects := newMemObjects(video.objectKey)
	worker := newTestWorker(t, pool, objects, fakeMedia{probe: Probe{DurationSeconds: 4, Height: 240, HasVideo: true}}, 3)
	outcome, err := worker.Handle(context.Background(), message(video, 1))
	if err != nil || outcome != OutcomeAck {
		t.Fatalf("outcome=%d err=%v", outcome, err)
	}
	processing, _, _, _, generation := videoFields(t, pool, video.videoID)
	if processing != "QUEUED" || generation != 2 || len(objects.order) != 0 {
		t.Fatalf("status=%s generation=%d uploads=%v", processing, generation, objects.order)
	}
}

func TestTransientFailureStopsAtTheAttemptCap(t *testing.T) {
	pool := openProcessingDB(t)
	video := insertQueuedVideo(t, pool)
	objects := newMemObjects(video.objectKey)
	worker := newTestWorker(t, pool, objects, fakeMedia{err: errors.New("storage timeout")}, 2)

	first, err := worker.Handle(context.Background(), message(video, 1))
	if err != nil || first != OutcomeRetry {
		t.Fatalf("first=%d err=%v", first, err)
	}
	processing, _, _, _, _ := videoFields(t, pool, video.videoID)
	if processing == "FAILED" {
		t.Fatal("failed before the attempt cap")
	}
	second, err := worker.Handle(context.Background(), message(video, 1))
	if err != nil || second != OutcomeAck {
		t.Fatalf("second=%d err=%v", second, err)
	}
	processing, _, _, _, _ = videoFields(t, pool, video.videoID)
	if processing != "FAILED" {
		t.Fatalf("status = %s", processing)
	}
	assertSafeError(t, pool, video.videoID)
}

func TestInvalidMessageIsAcknowledged(t *testing.T) {
	pool := openProcessingDB(t)
	worker := newTestWorker(t, pool, newMemObjects("unused"), fakeMedia{}, 3)
	outcome, err := worker.Handle(context.Background(), []byte(`{"video_id":`))
	if err != nil || outcome != OutcomeAck {
		t.Fatalf("outcome=%d err=%v", outcome, err)
	}
	if countQuery(t, pool, `SELECT count(*) FROM processing_jobs`) != 0 {
		t.Fatal("invalid message created a job")
	}
}

func TestExpiredLeaseBecomesClaimableOrFailed(t *testing.T) {
	pool := openProcessingDB(t)
	claimable := insertQueuedVideo(t, pool)
	exhausted := insertQueuedVideo(t, pool)
	insertJob(t, pool, claimable.videoID, "PROCESSING", 1, time.Now().Add(-time.Minute))
	insertJob(t, pool, exhausted.videoID, "PROCESSING", 2, time.Now().Add(-time.Minute))
	if _, err := pool.Exec(context.Background(), `UPDATE videos SET processing_status = 'PROCESSING' WHERE id IN ($1, $2)`, claimable.videoID, exhausted.videoID); err != nil {
		t.Fatal(err)
	}
	worker := newTestWorker(t, pool, newMemObjects("unused"), fakeMedia{}, 2)
	if err := worker.RecoverExpired(context.Background()); err != nil {
		t.Fatal(err)
	}
	if countQuery(t, pool, `SELECT count(*) FROM processing_jobs WHERE video_id = $1 AND status = 'QUEUED' AND lease_until IS NULL`, claimable.videoID) != 1 {
		t.Fatal("expired lease was not claimable")
	}
	processing, _, _, _, _ := videoFields(t, pool, exhausted.videoID)
	if processing != "FAILED" {
		t.Fatalf("exhausted = %s", processing)
	}
	assertSafeError(t, pool, exhausted.videoID)
}

func TestSeedReadyVideosAreNotQueued(t *testing.T) {
	pool := openProcessingDB(t)
	root, err := postgres.ModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := postgres.ApplySeed(context.Background(), pool, filepath.Join(root, "db", "seed", "dev.sql")); err != nil {
		t.Fatal(err)
	}
	worker := newTestWorker(t, pool, newMemObjects("unused"), fakeMedia{}, 3)
	if err := worker.RecoverExpired(context.Background()); err != nil {
		t.Fatal(err)
	}
	if countQuery(t, pool, `SELECT count(*) FROM processing_jobs`) != 0 {
		t.Fatal("seed videos gained a job")
	}
	if countQuery(t, pool, `SELECT count(*) FROM videos WHERE publication_status = 'PUBLISHED' AND processing_status = 'READY'`) < 1 {
		t.Fatal("published examples are no longer ready")
	}
}

func message(video queuedVideo, generation int) []byte {
	body, _ := json.Marshal(map[string]any{
		"video_id":          video.videoID,
		"generation":        generation,
		"upload_session_id": video.sessionID,
	})
	return body
}

func newTestWorker(t *testing.T, pool *pgxpool.Pool, objects Objects, media Runner, maxAttempts int) *Worker {
	t.Helper()
	worker, err := NewWorker(pool, objects, media, time.Minute, maxAttempts)
	if err != nil {
		t.Fatal(err)
	}
	return worker
}

func insertJob(t *testing.T, pool *pgxpool.Pool, videoID string, status string, attempts int, lease time.Time) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO processing_jobs (id, video_id, generation, status, lease_until, fencing_token, attempts)
		VALUES (gen_random_uuid(), $1, 1, $2, $3, 1, $4)
	`, videoID, status, lease, attempts); err != nil {
		t.Fatal(err)
	}
}

func assertSafeError(t *testing.T, pool *pgxpool.Pool, videoID string) {
	t.Helper()
	var message string
	if err := pool.QueryRow(context.Background(), `SELECT error_message FROM processing_jobs WHERE video_id = $1`, videoID).Scan(&message); err != nil {
		t.Fatal(err)
	}
	if message == "" || strings.Contains(message, "videos/") || strings.Contains(message, "object") {
		t.Fatalf("error = %s", message)
	}
}

type fakeMedia struct {
	probe  Probe
	err    error
	before func()
}

func (f fakeMedia) Probe(context.Context, string) (Probe, error) {
	if f.err != nil {
		return Probe{}, f.err
	}
	return f.probe, nil
}

func (f fakeMedia) Render(_ context.Context, _ string, outputDir string, _ Probe) error {
	if f.before != nil {
		f.before()
	}
	if err := os.WriteFile(filepath.Join(outputDir, "index.m3u8"), []byte("#EXTM3U\n#EXTINF:4,\nsegment000.ts\n"), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "segment000.ts"), []byte("segment"), 0o600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outputDir, "thumbnail.jpg"), []byte{0xFF, 0xD8, 0xFF, 0xD9}, 0o600)
}

type memObjects struct {
	source map[string][]byte
	order  []string
}

func newMemObjects(key string) *memObjects {
	return &memObjects{source: map[string][]byte{key: []byte("source-bytes")}}
}

func (m *memObjects) Download(_ context.Context, key string, dest string) error {
	body, ok := m.source[key]
	if !ok {
		return errors.New("objeto ausente")
	}
	return os.WriteFile(dest, body, 0o600)
}

func (m *memObjects) Upload(_ context.Context, key string, path string, _ string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	m.source[key] = body
	m.order = append(m.order, key)
	return nil
}
