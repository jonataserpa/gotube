package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"example.com/gotube/internal/account"
	"example.com/gotube/internal/identity"
)

func TestOwnerReadsProcessingWithoutObjectKeys(t *testing.T) {
	accounts, pool := newTestStack(t)
	directory := account.NewService(pool, testJWTSecret)
	owner := loginAccessToken(t, accounts, "owner@example.com", "correct-horse")
	other := loginAccessToken(t, accounts, "other@example.com", "correct-horse")
	videoID := createOwnedDraft(t, accounts, directory, owner)

	if _, err := pool.Exec(context.Background(), `UPDATE videos SET processing_status = 'FAILED' WHERE id = $1`, videoID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO processing_jobs (id, video_id, generation, status, fencing_token, attempts, error_code, error_message)
		SELECT gen_random_uuid(), id, generation, 'FAILED', 1, 1, 'MEDIA_INVALID', 'o arquivo não tem vídeo'
		FROM videos WHERE id = $1
	`, videoID); err != nil {
		t.Fatal(err)
	}

	missingToken := authedJSON(accounts, directory, http.MethodGet, "/v1/videos/"+videoID+"/processing", "", "")
	if missingToken.Code != http.StatusUnauthorized {
		t.Fatalf("token = %d %s", missingToken.Code, missingToken.Body.String())
	}
	forbidden := authedJSON(accounts, directory, http.MethodGet, "/v1/videos/"+videoID+"/processing", other, "")
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("other = %d %s", forbidden.Code, forbidden.Body.String())
	}
	assertErrorCode(t, forbidden.Body.Bytes(), "VIDEO_FORBIDDEN")
	unknown := authedJSON(accounts, directory, http.MethodGet, "/v1/videos/00000000-0000-4000-8000-000000000099/processing", owner, "")
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("unknown = %d %s", unknown.Code, unknown.Body.String())
	}

	view := authedJSON(accounts, directory, http.MethodGet, "/v1/videos/"+videoID+"/processing", owner, "")
	body := view.Body.String()
	if view.Code != http.StatusOK || !strings.Contains(body, `"processing_status":"FAILED"`) || !strings.Contains(body, "MEDIA_INVALID") {
		t.Fatalf("view = %d %s", view.Code, body)
	}
	if strings.Contains(body, "object_key") || strings.Contains(body, "http") || strings.Contains(body, "videos/") {
		t.Fatalf("view leaked a key: %s", body)
	}
}

func TestRetryOnlyRewritesAFailedVideo(t *testing.T) {
	accounts, pool := newTestStack(t)
	directory := account.NewService(pool, testJWTSecret)
	owner := loginAccessToken(t, accounts, "owner@example.com", "correct-horse")
	other := loginAccessToken(t, accounts, "other@example.com", "correct-horse")
	videoID := createOwnedDraft(t, accounts, directory, owner)
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO upload_sessions (id, video_id, owner_id, object_key, declared_size, expires_at, completed_at)
		SELECT gen_random_uuid(), v.id, c.owner_id, 'videos/' || v.id::text || '/source/session', 100, now(), now()
		FROM videos v JOIN channels c ON c.id = v.channel_id
		WHERE v.id = $1
	`, videoID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `UPDATE videos SET processing_status = 'QUEUED' WHERE id = $1`, videoID); err != nil {
		t.Fatal(err)
	}

	queued := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+videoID+"/retry-processing", owner, `{}`)
	if queued.Code != http.StatusConflict || uploadEvents(t, pool, videoID) != 0 {
		t.Fatalf("queued retry = %d events=%d body=%s", queued.Code, uploadEvents(t, pool, videoID), queued.Body.String())
	}
	assertErrorCode(t, queued.Body.Bytes(), "PROCESSING_NOT_FAILED")

	if _, err := pool.Exec(context.Background(), `UPDATE videos SET processing_status = 'FAILED' WHERE id = $1`, videoID); err != nil {
		t.Fatal(err)
	}
	forbidden := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+videoID+"/retry-processing", other, `{}`)
	if forbidden.Code != http.StatusForbidden || videoStatus(t, pool, videoID) != "FAILED" {
		t.Fatalf("other retry = %d status=%s", forbidden.Code, videoStatus(t, pool, videoID))
	}

	done := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+videoID+"/retry-processing", owner, `{}`)
	if done.Code != http.StatusAccepted || videoStatus(t, pool, videoID) != "QUEUED" || uploadEvents(t, pool, videoID) != 1 {
		t.Fatalf("retry = %d status=%s events=%d body=%s", done.Code, videoStatus(t, pool, videoID), uploadEvents(t, pool, videoID), done.Body.String())
	}
	var generation int
	var publication string
	if err := pool.QueryRow(context.Background(), `SELECT generation, publication_status FROM videos WHERE id = $1`, videoID).Scan(&generation, &publication); err != nil {
		t.Fatal(err)
	}
	if generation != 2 || publication != "DRAFT" {
		t.Fatalf("generation=%d publication=%s", generation, publication)
	}

	if _, err := pool.Exec(context.Background(), `UPDATE videos SET processing_status = 'READY' WHERE id = $1`, videoID); err != nil {
		t.Fatal(err)
	}
	ready := authedJSON(accounts, directory, http.MethodPost, "/v1/videos/"+videoID+"/retry-processing", owner, `{}`)
	if ready.Code != http.StatusConflict || uploadEvents(t, pool, videoID) != 1 {
		t.Fatalf("ready retry = %d events=%d", ready.Code, uploadEvents(t, pool, videoID))
	}
	if err := pool.QueryRow(context.Background(), `SELECT generation FROM videos WHERE id = $1`, videoID).Scan(&generation); err != nil {
		t.Fatal(err)
	}
	if generation != 2 {
		t.Fatalf("generation changed to %d", generation)
	}
}

func TestProcessingRoutesWithoutDatabaseKeepHealth(t *testing.T) {
	for _, target := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/v1/videos/00000000-0000-4000-8000-000000000099/processing"},
		{http.MethodPost, "/v1/videos/00000000-0000-4000-8000-000000000099/retry-processing"},
	} {
		rec := authedJSON(nil, nil, target.method, target.path, "token", `{}`)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s %s = %d", target.method, target.path, rec.Code)
		}
	}
	health := authedJSON(nil, nil, http.MethodGet, "/healthz", "", "")
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"status":"ok"`) {
		t.Fatalf("health = %d %s", health.Code, health.Body.String())
	}
}

func createOwnedDraft(t *testing.T, accounts *identity.Service, directory Directory, token string) string {
	t.Helper()
	channel := authedJSON(accounts, directory, http.MethodPost, "/v1/channels", token, `{"handle":"ada-proc","name":"Canal Ada"}`)
	if channel.Code != http.StatusCreated {
		t.Fatalf("channel = %d %s", channel.Code, channel.Body.String())
	}
	draft := authedJSON(accounts, directory, http.MethodPost, "/v1/videos", token, `{"title":"Aula de Go"}`)
	if draft.Code != http.StatusCreated {
		t.Fatalf("draft = %d %s", draft.Code, draft.Body.String())
	}
	return jsonString(t, draft.Body.Bytes(), "id")
}
