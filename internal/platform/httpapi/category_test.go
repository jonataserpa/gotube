package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"example.com/gotube/internal/account"
	"example.com/gotube/internal/platform/postgres"
)

func TestCategoriesStayActiveAndOwned(t *testing.T) {
	accounts, pool := newTestStack(t)
	root, err := postgres.ModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := postgres.ApplySeed(context.Background(), pool, filepath.Join(root, "db", "seed", "dev.sql")); err != nil {
		t.Fatal(err)
	}
	directory := account.NewService(pool, testJWTSecret)
	owner := loginAccessToken(t, accounts, "cat-owner@example.com", "correct-horse")
	other := loginAccessToken(t, accounts, "cat-other@example.com", "correct-horse")

	list := httptest.NewRecorder()
	NewMux(nil, accounts, directory).ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/v1/categories", nil))
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"slug":"programacao"`) || !strings.Contains(list.Body.String(), `"slug":"go"`) || strings.Contains(list.Body.String(), "arquivada") {
		t.Fatalf("list = %d %s", list.Code, list.Body.String())
	}

	created := authedJSON(accounts, directory, http.MethodPost, "/v1/channels", owner, `{
		"handle":"categorias","name":"Com categorias",
		"category_ids":["00000000-0000-4000-8000-000000000101","00000000-0000-4000-8000-000000000102"]
	}`)
	if created.Code != http.StatusCreated || !strings.Contains(created.Body.String(), `"slug":"go"`) || !strings.Contains(created.Body.String(), `"slug":"programacao"`) {
		t.Fatalf("create = %d %s", created.Code, created.Body.String())
	}
	id := channelID(t, created.Body.String())

	rejected := authedJSON(accounts, directory, http.MethodPatch, "/v1/channels/"+id, owner, `{
		"category_ids":["00000000-0000-4000-8000-000000000103"]
	}`)
	if rejected.Code != http.StatusBadRequest {
		t.Fatalf("inactive = %d %s", rejected.Code, rejected.Body.String())
	}
	public := httptest.NewRecorder()
	NewMux(nil, accounts, directory).ServeHTTP(public, httptest.NewRequest(http.MethodGet, "/v1/channels/categorias", nil))
	if !strings.Contains(public.Body.String(), `"slug":"go"`) || strings.Contains(public.Body.String(), "arquivada") {
		t.Fatalf("set changed after rejection: %s", public.Body.String())
	}

	alien := authedJSON(accounts, directory, http.MethodPatch, "/v1/channels/"+id, other, `{
		"category_ids":["00000000-0000-4000-8000-000000000101"]
	}`)
	if alien.Code != http.StatusForbidden {
		t.Fatalf("alien = %d %s", alien.Code, alien.Body.String())
	}

	if _, err := pool.Exec(context.Background(), `UPDATE categories SET active = FALSE WHERE slug = 'go'`); err != nil {
		t.Fatal(err)
	}
	hidden := httptest.NewRecorder()
	NewMux(nil, accounts, directory).ServeHTTP(hidden, httptest.NewRequest(http.MethodGet, "/v1/channels/categorias", nil))
	if strings.Contains(hidden.Body.String(), `"slug":"go"`) || !strings.Contains(hidden.Body.String(), `"slug":"programacao"`) {
		t.Fatalf("inactive still public: %s", hidden.Body.String())
	}

	down := httptest.NewRecorder()
	NewMux(nil, nil, nil).ServeHTTP(down, httptest.NewRequest(http.MethodGet, "/v1/categories", nil))
	if down.Code != http.StatusServiceUnavailable || strings.Contains(down.Body.String(), `"slug"`) {
		t.Fatalf("down = %d %s", down.Code, down.Body.String())
	}
	assertErrorCode(t, down.Body.Bytes(), "DATABASE_UNAVAILABLE")
	health := httptest.NewRecorder()
	NewMux(nil, nil, nil).ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health = %d %s", health.Code, health.Body.String())
	}
}
