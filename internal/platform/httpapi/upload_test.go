package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/gotube/internal/account"
)

func TestUploadSessionAuthorizesOwnerAndCompletesOnce(t *testing.T) {
	accounts, pool := newTestStack(t)
	store := &fakeObjectStore{objects: map[string]int64{}}
	directory := account.NewService(pool, testJWTSecret)
	directory.UseStorage(store, 1024)
	owner := loginAccessToken(t, accounts, "owner@example.com", "correct-horse")
	other := loginAccessToken(t, accounts, "other@example.com", "correct-horse")

	channel := authedJSON(accounts, directory, http.MethodPost, "/v1/channels", owner, `{"handle":"ada-channel","name":"Canal Ada"}`)
	if channel.Code != http.StatusCreated {
		t.Fatalf("channel = %d %s", channel.Code, channel.Body.String())
	}
	draft := authedJSON(accounts, directory, http.MethodPost, "/v1/videos", owner, `{"title":"Aula de Go"}`)
	if draft.Code != http.StatusCreated {
		t.Fatalf("draft = %d %s", draft.Code, draft.Body.String())
	}
	videoID := jsonString(t, draft.Body.Bytes(), "id")

	withKey := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+videoID+"/upload-sessions", owner, `{"declared_size":100,"object_key":"client/key"}`)
	if withKey.Code != http.StatusBadRequest {
		t.Fatalf("client key = %d %s", withKey.Code, withKey.Body.String())
	}
	if countVideos(t, pool, `SELECT count(*) FROM upload_sessions WHERE object_key = 'client/key'`) != 0 {
		t.Fatal("client key was stored")
	}

	forbidden := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+videoID+"/upload-sessions", other, `{"declared_size":100}`)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("other account = %d %s", forbidden.Code, forbidden.Body.String())
	}
	assertErrorCode(t, forbidden.Body.Bytes(), "VIDEO_FORBIDDEN")

	unknown := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/00000000-0000-4000-8000-000000000099/upload-sessions", owner, `{"declared_size":100}`)
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("unknown = %d %s", unknown.Code, unknown.Body.String())
	}

	tooBig := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+videoID+"/upload-sessions", owner, `{"declared_size":2048}`)
	if tooBig.Code != http.StatusBadRequest {
		t.Fatalf("declared size = %d %s", tooBig.Code, tooBig.Body.String())
	}

	session := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+videoID+"/upload-sessions", owner, `{"declared_size":100}`)
	if session.Code != http.StatusCreated {
		t.Fatalf("session = %d %s", session.Code, session.Body.String())
	}
	sessionBody := session.Body.String()
	if !strings.Contains(sessionBody, "videos/"+videoID+"/source/") || strings.Contains(sessionBody, "client/key") {
		t.Fatalf("authorization = %s", sessionBody)
	}
	sessionID := jsonString(t, session.Body.Bytes(), "id")

	missing := authedJSON(accounts, directory, http.MethodPost, "/v1/upload-sessions/"+sessionID+"/complete", owner, `{}`)
	if missing.Code != http.StatusConflict || videoStatus(t, pool, videoID) != "WAITING_UPLOAD" {
		t.Fatalf("missing object = %d status=%s body=%s", missing.Code, videoStatus(t, pool, videoID), missing.Body.String())
	}

	store.objects[serverKey(sessionBody)] = 101
	oversized := authedJSON(accounts, directory, http.MethodPost, "/v1/upload-sessions/"+sessionID+"/complete", owner, `{}`)
	if oversized.Code != http.StatusConflict || videoStatus(t, pool, videoID) != "WAITING_UPLOAD" {
		t.Fatalf("oversized = %d status=%s", oversized.Code, videoStatus(t, pool, videoID))
	}

	store.objects[serverKey(sessionBody)] = 100
	done := authedJSON(accounts, directory, http.MethodPost, "/v1/upload-sessions/"+sessionID+"/complete", owner, `{}`)
	if done.Code != http.StatusAccepted || videoStatus(t, pool, videoID) != "QUEUED" {
		t.Fatalf("complete = %d status=%s body=%s", done.Code, videoStatus(t, pool, videoID), done.Body.String())
	}
	if uploadEvents(t, pool, videoID) != 1 {
		t.Fatalf("events = %d", uploadEvents(t, pool, videoID))
	}
	payload := outboxPayload(t, pool, videoID)
	if strings.Contains(payload, "http") || strings.Contains(payload, "object_key") || !strings.Contains(payload, `"generation": 1`) {
		t.Fatalf("payload = %s", payload)
	}

	again := authedJSON(accounts, directory, http.MethodPost, "/v1/upload-sessions/"+sessionID+"/complete", owner, `{}`)
	if again.Code != http.StatusAccepted || uploadEvents(t, pool, videoID) != 1 || videoStatus(t, pool, videoID) != "QUEUED" {
		t.Fatalf("second = %d events=%d status=%s", again.Code, uploadEvents(t, pool, videoID), videoStatus(t, pool, videoID))
	}

	afterQueue := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+videoID+"/upload-sessions", owner, `{"declared_size":100}`)
	if afterQueue.Code != http.StatusConflict {
		t.Fatalf("queued video = %d %s", afterQueue.Code, afterQueue.Body.String())
	}

	second := authedJSON(accounts, directory, http.MethodPost, "/v1/videos", owner, `{"title":"Outra aula"}`)
	secondID := jsonString(t, second.Body.Bytes(), "id")
	opened := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+secondID+"/upload-sessions", owner, `{"declared_size":50}`)
	if opened.Code != http.StatusCreated {
		t.Fatalf("second session = %d %s", opened.Code, opened.Body.String())
	}
	store.fail = true
	failed := authedJSON(accounts, directory, http.MethodPost, "/v1/upload-sessions/"+jsonString(t, opened.Body.Bytes(), "id")+"/complete", owner, `{}`)
	if failed.Code != http.StatusServiceUnavailable || strings.Contains(failed.Body.String(), "QUEUED") || videoStatus(t, pool, secondID) != "WAITING_UPLOAD" {
		t.Fatalf("storage down = %d status=%s body=%s", failed.Code, videoStatus(t, pool, secondID), failed.Body.String())
	}

	down := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/videos/"+videoID+"/upload-sessions", strings.NewReader(`{"declared_size":100}`))
	NewMux(nil, nil, nil).ServeHTTP(down, req)
	if down.Code != http.StatusServiceUnavailable {
		t.Fatalf("no database = %d", down.Code)
	}
	health := httptest.NewRecorder()
	NewMux(nil, nil, nil).ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health = %d", health.Code)
	}
}

type fakeObjectStore struct {
	objects map[string]int64
	fail    bool
}

func (f *fakeObjectStore) PresignPut(_ context.Context, key string, expiresAt time.Time) (string, error) {
	if expiresAt.Before(time.Now()) {
		return "", context.DeadlineExceeded
	}
	return "https://storage.example/" + key, nil
}

func (f *fakeObjectStore) Head(_ context.Context, key string) (int64, bool, error) {
	if f.fail {
		return 0, false, context.DeadlineExceeded
	}
	size, ok := f.objects[key]
	return size, ok, nil
}

func TestExpiredUploadSessionCanBeReplaced(t *testing.T) {
	accounts, pool := newTestStack(t)
	store := &fakeObjectStore{objects: map[string]int64{}}
	directory := account.NewService(pool, testJWTSecret)
	directory.UseStorage(store, 1024)
	owner := loginAccessToken(t, accounts, "owner@example.com", "correct-horse")
	if channel := authedJSON(accounts, directory, http.MethodPost, "/v1/channels", owner, `{"handle":"ada-channel","name":"Canal Ada"}`); channel.Code != http.StatusCreated {
		t.Fatal(channel.Body.String())
	}
	draft := authedJSON(accounts, directory, http.MethodPost, "/v1/videos", owner, `{"title":"Aula curta"}`)
	videoID := jsonString(t, draft.Body.Bytes(), "id")
	opened := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+videoID+"/upload-sessions", owner, `{"declared_size":40}`)
	sessionID := jsonString(t, opened.Body.Bytes(), "id")
	if _, err := pool.Exec(context.Background(), `UPDATE upload_sessions SET expires_at = now() - interval '1 minute' WHERE id = $1`, sessionID); err != nil {
		t.Fatal(err)
	}
	expired := authedJSON(accounts, directory, http.MethodPost, "/v1/upload-sessions/"+sessionID+"/complete", owner, `{}`)
	if expired.Code != http.StatusConflict || videoStatus(t, pool, videoID) != "WAITING_UPLOAD" {
		t.Fatalf("expired = %d status=%s", expired.Code, videoStatus(t, pool, videoID))
	}
	replaced := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+videoID+"/upload-sessions", owner, `{"declared_size":40}`)
	if replaced.Code != http.StatusCreated || jsonString(t, replaced.Body.Bytes(), "id") == sessionID {
		t.Fatalf("replacement = %d %s", replaced.Code, replaced.Body.String())
	}
}

func jsonString(t *testing.T, body []byte, key string) string {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	value, _ := payload[key].(string)
	if value == "" {
		t.Fatalf("missing %s in %s", key, body)
	}
	return value
}

func serverKey(body string) string {
	const marker = "https://storage.example/"
	index := strings.Index(body, marker)
	if index < 0 {
		return ""
	}
	rest := body[index+len(marker):]
	end := strings.IndexAny(rest, `"`)
	if end < 0 {
		return rest
	}
	return rest[:end]
}

func videoStatus(t *testing.T, pool *pgxpool.Pool, videoID string) string {
	t.Helper()
	var status string
	if err := pool.QueryRow(context.Background(), `SELECT processing_status FROM videos WHERE id = $1`, videoID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	return status
}

func uploadEvents(t *testing.T, pool *pgxpool.Pool, videoID string) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM outbox WHERE type = 'video.uploaded.v1' AND aggregate_id = $1`, videoID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func outboxPayload(t *testing.T, pool *pgxpool.Pool, videoID string) string {
	t.Helper()
	var payload string
	if err := pool.QueryRow(context.Background(), `SELECT payload::text FROM outbox WHERE aggregate_id = $1`, videoID).Scan(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}
