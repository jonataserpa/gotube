package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/gotube/internal/identity"
	"example.com/gotube/internal/platform/postgres"
)

const testJWTSecret = "test-secret-at-least-32-bytes-long"

func TestRegisterCreatesAccountWithoutSession(t *testing.T) {
	accounts := newTestAccounts(t)
	rec := postJSON(accounts, "/v1/auth/register", `{"email":"Person@Example.com","password":"correct-horse"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("register set a cookie: %v", rec.Result().Cookies())
	}
	var created struct {
		ID           string `json:"id"`
		Email        string `json:"email"`
		PasswordHash string `json:"password_hash"`
		AccessToken  string `json:"access_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Email != "person@example.com" || created.PasswordHash != "" || created.AccessToken != "" {
		t.Fatalf("created = %#v", created)
	}
	if strings.Contains(rec.Body.String(), "argon2") || strings.Contains(rec.Body.String(), "correct-horse") {
		t.Fatalf("response leaked a secret: %s", rec.Body.String())
	}

	conflict := postJSON(accounts, "/v1/auth/register", `{"email":"  person@example.com  ","password":"correct-horse"}`)
	if conflict.Code != http.StatusConflict {
		t.Fatalf("conflict status = %d body = %s", conflict.Code, conflict.Body.String())
	}
	assertErrorCode(t, conflict.Body.Bytes(), "EMAIL_ALREADY_REGISTERED")

	short := postJSON(accounts, "/v1/auth/register", `{"email":"other@example.com","password":"1234567"}`)
	if short.Code != http.StatusBadRequest {
		t.Fatalf("short status = %d body = %s", short.Code, short.Body.String())
	}
	assertErrorCode(t, short.Body.Bytes(), "VALIDATION_ERROR")

	unknown := postJSON(accounts, "/v1/auth/register", `{"email":"third@example.com","password":"correct-horse","extra":true}`)
	if unknown.Code != http.StatusBadRequest {
		t.Fatalf("unknown status = %d body = %s", unknown.Code, unknown.Body.String())
	}
	assertErrorCode(t, unknown.Body.Bytes(), "VALIDATION_ERROR")

	down := httptest.NewRecorder()
	NewMux(nil, nil, nil).ServeHTTP(down, httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString(`{"email":"down@example.com","password":"correct-horse"}`)))
	if down.Code != http.StatusServiceUnavailable {
		t.Fatalf("down status = %d body = %s", down.Code, down.Body.String())
	}
	assertErrorCode(t, down.Body.Bytes(), "DATABASE_UNAVAILABLE")
	if strings.Contains(down.Body.String(), "criada") || strings.Contains(down.Body.String(), `"id"`) {
		t.Fatalf("503 confirmed an account: %s", down.Body.String())
	}
	health := httptest.NewRecorder()
	NewMux(nil, nil, nil).ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"status":"ok"`) {
		t.Fatalf("health = %d %s", health.Code, health.Body.String())
	}
}

func newTestAccounts(t *testing.T) *identity.Service {
	t.Helper()
	accounts, _ := newTestStack(t)
	return accounts
}

func newTestStack(t *testing.T) (*identity.Service, *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	pool := postgres.OpenCatalogTestDB(t)
	root, err := postgres.ModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := postgres.ApplyMigrations(ctx, pool, filepath.Join(root, "db", "migrations")); err != nil {
		t.Fatal(err)
	}
	accounts := identity.NewService(pool, identity.Settings{JWTSecret: testJWTSecret})
	return accounts, pool
}

func postJSON(accounts *identity.Service, path string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	NewMux(nil, accounts, nil).ServeHTTP(rec, req)
	return rec
}
