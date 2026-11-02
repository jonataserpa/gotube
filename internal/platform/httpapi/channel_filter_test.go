package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"example.com/gotube/internal/catalog"
	"example.com/gotube/internal/platform/postgres"
)

func TestPublicListFiltersByChannelWithoutRelaxingVisibility(t *testing.T) {
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
		INSERT INTO channels (id, handle, name) VALUES
			('00000000-0000-4000-8000-0000000000a1', 'um', 'Um'),
			('00000000-0000-4000-8000-0000000000a2', 'dois', 'Dois');
		INSERT INTO videos (id, channel_id, title, duration_seconds, visibility, publication_status, published_at) VALUES
			('00000000-0000-4000-8000-0000000000b1', '00000000-0000-4000-8000-0000000000a1', 'Publicado Um', 5, 'PUBLIC', 'PUBLISHED', now()),
			('00000000-0000-4000-8000-0000000000b2', '00000000-0000-4000-8000-0000000000a1', 'Rascunho Um', 5, 'PUBLIC', 'DRAFT', NULL),
			('00000000-0000-4000-8000-0000000000b3', '00000000-0000-4000-8000-0000000000a2', 'Publicado Dois', 5, 'PUBLIC', 'PUBLISHED', now());
	`); err != nil {
		t.Fatal(err)
	}
	store := catalog.NewStore(pool)

	filtered := httptest.NewRecorder()
	NewMux(store, nil, nil).ServeHTTP(filtered, httptest.NewRequest(http.MethodGet, "/v1/videos?channel_id=00000000-0000-4000-8000-0000000000a1", nil))
	if filtered.Code != http.StatusOK {
		t.Fatalf("filtered = %d %s", filtered.Code, filtered.Body.String())
	}
	var page struct {
		Items []catalog.Item `json:"items"`
	}
	if err := json.Unmarshal(filtered.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].Title != "Publicado Um" {
		t.Fatalf("page = %#v", page.Items)
	}

	empty := httptest.NewRecorder()
	NewMux(store, nil, nil).ServeHTTP(empty, httptest.NewRequest(http.MethodGet, "/v1/videos?channel_id=00000000-0000-4000-8000-0000000000a9", nil))
	if empty.Code != http.StatusOK {
		t.Fatalf("empty = %d %s", empty.Code, empty.Body.String())
	}
	var emptyPage struct {
		Items []catalog.Item `json:"items"`
	}
	if err := json.Unmarshal(empty.Body.Bytes(), &emptyPage); err != nil {
		t.Fatal(err)
	}
	if len(emptyPage.Items) != 0 {
		t.Fatalf("unknown channel returned %#v", emptyPage.Items)
	}

	bad := httptest.NewRecorder()
	NewMux(store, nil, nil).ServeHTTP(bad, httptest.NewRequest(http.MethodGet, "/v1/videos?channel_id=nao-e-uuid", nil))
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("bad = %d %s", bad.Code, bad.Body.String())
	}
	assertErrorCode(t, bad.Body.Bytes(), "INVALID_CHANNEL")
}
