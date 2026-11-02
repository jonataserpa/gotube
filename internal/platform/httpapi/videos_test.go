package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"example.com/gotube/internal/catalog"
	"example.com/gotube/internal/platform/postgres"
)

func TestInvalidLimitAndCursorDoNotReturnAPage(t *testing.T) {
	store := catalog.NewStore(unavailableDB{})
	for _, target := range []string{"/v1/videos?limit=51", "/v1/videos?limit=0", "/v1/videos?limit=abc", "/v1/videos?cursor=nao-e-cursor"} {
		req := httptest.NewRequest(http.MethodGet, target, nil)
		rec := httptest.NewRecorder()
		NewMux(store, nil, nil).ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d body = %s", target, rec.Code, rec.Body.String())
		}
		want := "INVALID_LIMIT"
		if strings.Contains(target, "cursor") {
			want = "INVALID_CURSOR"
		}
		assertErrorCode(t, rec.Body.Bytes(), want)
		if jsonHasItems(rec.Body.Bytes()) {
			t.Fatalf("%s returned a page: %s", target, rec.Body.String())
		}
	}
}

func TestDatabaseDownIsNotAnEmptyCatalog(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/videos", nil)
	rec := httptest.NewRecorder()
	NewMux(catalog.NewStore(unavailableDB{}), nil, nil).ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	if jsonHasItems(rec.Body.Bytes()) {
		t.Fatalf("503 body was a page: %s", rec.Body.String())
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error.Code != "DATABASE_UNAVAILABLE" || payload.Status == "ok" {
		t.Fatalf("payload = %#v", payload)
	}

	health := httptest.NewRecorder()
	NewMux(catalog.NewStore(unavailableDB{}), nil, nil).ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK || !json.Valid(health.Body.Bytes()) {
		t.Fatalf("health = %d %s", health.Code, health.Body.String())
	}
}

func TestPublicListThroughHTTP(t *testing.T) {
	ctx := context.Background()
	pool := postgres.OpenCatalogTestDB(t)
	root, err := postgres.ModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := postgres.ApplyMigrations(ctx, pool, filepath.Join(root, "db", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO channels (id, handle, name)
		VALUES ('00000000-0000-4000-8000-0000000000aa', 'canal', 'Canal Público');
	`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 13; i++ {
		if _, err := pool.Exec(ctx, `
			INSERT INTO videos (id, channel_id, title, duration_seconds, visibility, publication_status, published_at)
			VALUES (gen_random_uuid(), '00000000-0000-4000-8000-0000000000aa', 'Público', 5, 'PUBLIC', 'PUBLISHED', now() - make_interval(secs => $1))
		`, i); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/v1/videos", nil)
	rec := httptest.NewRecorder()
	NewMux(catalog.NewStore(pool), nil, nil).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var page struct {
		Items      []catalog.Item `json:"items"`
		NextCursor *string        `json:"next_cursor"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 12 || page.NextCursor == nil {
		t.Fatalf("default page = %d items cursor %#v", len(page.Items), page.NextCursor)
	}
	for _, item := range page.Items {
		if item.Channel.Name == "" || item.Title == "" {
			t.Fatalf("item = %#v", item)
		}
	}
}

type unavailableDB struct{}

func (unavailableDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("connection refused")
}

func assertErrorCode(t *testing.T, body []byte, want string) {
	t.Helper()
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error.Code != want {
		t.Fatalf("code = %s, want %s", payload.Error.Code, want)
	}
}

func jsonHasItems(body []byte) bool {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}
	_, ok := payload["items"]
	return ok
}
