package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/gotube/internal/account"
	"example.com/gotube/internal/catalog"
	"example.com/gotube/internal/platform/postgres"
)

func TestCreateDraftRequiresChannelAndStaysOutOfPublicList(t *testing.T) {
	accounts, pool := newTestStack(t)
	root, err := postgres.ModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := postgres.ApplySeed(context.Background(), pool, filepath.Join(root, "db", "seed", "dev.sql")); err != nil {
		t.Fatal(err)
	}
	directory := account.NewService(pool, testJWTSecret)
	videos := catalog.NewStore(pool)
	owner := loginAccessToken(t, accounts, "owner@example.com", "correct-horse")

	labBefore := countVideos(t, pool, `SELECT count(*) FROM videos WHERE channel_id = '00000000-0000-4000-8000-000000000001'`)
	ownedBefore := countVideos(t, pool, `SELECT count(*) FROM videos v JOIN channels c ON c.id = v.channel_id WHERE c.owner_id IS NOT NULL`)

	missing := authedJSON(accounts, directory, http.MethodPost, "/v1/videos", owner, `{"title":"Aula de Go"}`)
	if missing.Code != http.StatusConflict {
		t.Fatalf("without channel = %d %s", missing.Code, missing.Body.String())
	}
	assertErrorCode(t, missing.Body.Bytes(), "CHANNEL_REQUIRED")
	if countVideos(t, pool, `SELECT count(*) FROM videos v JOIN channels c ON c.id = v.channel_id WHERE c.owner_id IS NOT NULL`) != ownedBefore {
		t.Fatal("conflict created a video")
	}

	createdChannel := authedJSON(accounts, directory, http.MethodPost, "/v1/channels", owner, `{"handle":"ada-channel","name":"Canal Ada"}`)
	if createdChannel.Code != http.StatusCreated {
		t.Fatalf("channel = %d %s", createdChannel.Code, createdChannel.Body.String())
	}

	blank := authedJSON(accounts, directory, http.MethodPost, "/v1/videos", owner, `{"title":"   "}`)
	if blank.Code != http.StatusBadRequest {
		t.Fatalf("blank = %d %s", blank.Code, blank.Body.String())
	}
	unknown := authedJSON(accounts, directory, http.MethodPost, "/v1/videos", owner, `{"title":"Aula","object_key":"client/key"}`)
	if unknown.Code != http.StatusBadRequest {
		t.Fatalf("unknown = %d %s", unknown.Code, unknown.Body.String())
	}
	if countVideos(t, pool, `SELECT count(*) FROM videos v JOIN channels c ON c.id = v.channel_id WHERE c.handle = 'ada-channel'`) != 0 {
		t.Fatal("invalid body created a video")
	}

	anonymous := authedJSON(accounts, directory, http.MethodPost, "/v1/videos", "", `{"title":"Aula de Go"}`)
	if anonymous.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous = %d %s", anonymous.Code, anonymous.Body.String())
	}

	created := authedJSON(accounts, directory, http.MethodPost, "/v1/videos", owner, `{"title":"  Aula de Go  "}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("draft = %d %s", created.Code, created.Body.String())
	}
	body := created.Body.String()
	for _, part := range []string{`"publication_status":"DRAFT"`, `"visibility":"PRIVATE"`, `"processing_status":"WAITING_UPLOAD"`, `"duration_seconds":0`} {
		if !strings.Contains(body, part) {
			t.Fatalf("draft missing %s: %s", part, body)
		}
	}
	if strings.Contains(body, "@") || strings.Contains(body, "object_key") {
		t.Fatalf("draft leaked private data: %s", body)
	}

	list := httptest.NewRecorder()
	NewMux(videos, accounts, directory).ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/v1/videos", nil))
	if list.Code != http.StatusOK || strings.Contains(list.Body.String(), "Aula de Go") {
		t.Fatalf("public list = %d %s", list.Code, list.Body.String())
	}
	if countVideos(t, pool, `SELECT count(*) FROM videos WHERE channel_id = '00000000-0000-4000-8000-000000000001'`) != labBefore {
		t.Fatal("draft was attached to the lab channel")
	}

	unavailable := httptest.NewRecorder()
	NewMux(nil, nil, nil).ServeHTTP(unavailable, httptest.NewRequest(http.MethodPost, "/v1/videos", strings.NewReader(`{"title":"Aula"}`)))
	if unavailable.Code != http.StatusServiceUnavailable || strings.Contains(unavailable.Body.String(), `"id"`) {
		t.Fatalf("unavailable = %d %s", unavailable.Code, unavailable.Body.String())
	}
	health := httptest.NewRecorder()
	NewMux(nil, nil, nil).ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health = %d", health.Code)
	}
}

func countVideos(t *testing.T, pool *pgxpool.Pool, query string) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(), query).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
