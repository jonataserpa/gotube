package processing

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRelayMarksOnlyAfterConfirm(t *testing.T) {
	pool := openProcessingDB(t)
	video := insertQueuedVideo(t, pool)
	insertOutbox(t, pool, video)
	publisher := &memoryPublisher{}
	relay := newTestRelay(t, pool, publisher)

	published, err := relay.PublishOnce(context.Background())
	if err != nil || published != 1 {
		t.Fatalf("published=%d err=%v", published, err)
	}
	body := string(publisher.bodies[0])
	if !strings.Contains(body, `"video_id"`) || !strings.Contains(body, `"generation"`) || strings.Contains(body, "http") || strings.Contains(body, "object_key") {
		t.Fatalf("payload = %s", body)
	}
	if countQuery(t, pool, `SELECT count(*) FROM outbox WHERE aggregate_id = $1 AND published_at IS NOT NULL`, video.videoID) != 1 {
		t.Fatal("row was not marked")
	}

	if _, err := pool.Exec(context.Background(), `UPDATE outbox SET published_at = NULL, claim_until = now() - interval '1 minute' WHERE aggregate_id = $1`, video.videoID); err != nil {
		t.Fatal(err)
	}
	again, err := relay.PublishOnce(context.Background())
	if err != nil || again != 1 || len(publisher.bodies) != 2 {
		t.Fatalf("republish=%d messages=%d err=%v", again, len(publisher.bodies), err)
	}
	if countQuery(t, pool, `SELECT count(*) FROM outbox WHERE aggregate_id = $1`, video.videoID) != 1 {
		t.Fatal("republish deleted the row")
	}
}

func TestRelayKeepsRowWhenBrokerRefuses(t *testing.T) {
	pool := openProcessingDB(t)
	video := insertQueuedVideo(t, pool)
	insertOutbox(t, pool, video)
	relay := newTestRelay(t, pool, &memoryPublisher{err: errors.New("fila indisponível")})

	published, err := relay.PublishOnce(context.Background())
	if err == nil || published != 0 {
		t.Fatalf("published=%d err=%v", published, err)
	}
	if countQuery(t, pool, `SELECT count(*) FROM outbox WHERE aggregate_id = $1 AND published_at IS NULL`, video.videoID) != 1 {
		t.Fatal("refused publish marked the row")
	}
}

func TestTwoRelaysDoNotClaimTheSameRow(t *testing.T) {
	pool := openProcessingDB(t)
	video := insertQueuedVideo(t, pool)
	insertOutbox(t, pool, video)
	started := make(chan struct{})
	release := make(chan struct{})
	firstPub := &memoryPublisher{started: started, release: release}
	secondPub := &memoryPublisher{}
	first := newTestRelay(t, pool, firstPub)
	second := newTestRelay(t, pool, secondPub)

	done := make(chan error, 1)
	go func() {
		_, err := first.PublishOnce(context.Background())
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("first relay did not claim")
	}
	published, err := second.PublishOnce(context.Background())
	if err != nil || published != 0 || len(secondPub.bodies) != 0 {
		t.Fatalf("second published=%d err=%v", published, err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if len(firstPub.bodies) != 1 {
		t.Fatalf("first messages = %d", len(firstPub.bodies))
	}
}

func insertOutbox(t *testing.T, pool *pgxpool.Pool, video queuedVideo) {
	t.Helper()
	payload := `{"video_id":"` + video.videoID + `","generation":1,"upload_session_id":"` + video.sessionID + `"}`
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO outbox (id, aggregate_id, type, version, payload)
		VALUES (gen_random_uuid(), $1, 'video.uploaded.v1', 1, $2::jsonb)
	`, video.videoID, payload); err != nil {
		t.Fatal(err)
	}
}

func newTestRelay(t *testing.T, pool *pgxpool.Pool, publisher Publisher) *Relay {
	t.Helper()
	relay, err := NewRelay(pool, publisher, 30*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	return relay
}

type memoryPublisher struct {
	mu      sync.Mutex
	bodies  [][]byte
	err     error
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (p *memoryPublisher) Publish(_ context.Context, _ string, payload []byte) error {
	if p.started != nil {
		p.once.Do(func() { close(p.started) })
		<-p.release
	}
	if p.err != nil {
		return p.err
	}
	p.mu.Lock()
	p.bodies = append(p.bodies, append([]byte(nil), payload...))
	p.mu.Unlock()
	return nil
}
