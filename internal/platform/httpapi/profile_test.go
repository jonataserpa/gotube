package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/gotube/internal/account"
	"example.com/gotube/internal/identity"
)

func TestProfileReadsAndUpdatesDisplayName(t *testing.T) {
	accounts, pool := newTestStack(t)
	directory := account.NewService(pool, testJWTSecret)
	token := loginAccessToken(t, accounts, "ada@example.com", "correct-horse")

	read := authedJSON(accounts, directory, http.MethodGet, "/v1/me", token, "")
	if read.Code != http.StatusOK {
		t.Fatalf("read status = %d body = %s", read.Code, read.Body.String())
	}
	var profile struct {
		ID           string  `json:"id"`
		Email        string  `json:"email"`
		DisplayName  *string `json:"display_name"`
		PasswordHash string  `json:"password_hash"`
	}
	if err := json.Unmarshal(read.Body.Bytes(), &profile); err != nil {
		t.Fatal(err)
	}
	if profile.ID == "" || profile.Email != "ada@example.com" || profile.DisplayName != nil || profile.PasswordHash != "" {
		t.Fatalf("profile = %#v", profile)
	}
	if strings.Contains(read.Body.String(), "argon2") || strings.Contains(read.Body.String(), "correct-horse") {
		t.Fatalf("profile leaked a secret: %s", read.Body.String())
	}

	missing := authedJSON(accounts, directory, http.MethodGet, "/v1/me", "", "")
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("missing status = %d body = %s", missing.Code, missing.Body.String())
	}
	assertErrorCode(t, missing.Body.Bytes(), "UNAUTHENTICATED")
	if strings.Contains(missing.Body.String(), "ada@example.com") {
		t.Fatalf("401 leaked email: %s", missing.Body.String())
	}

	blank := authedJSON(accounts, directory, http.MethodPatch, "/v1/me", token, `{"display_name":"   "}`)
	if blank.Code != http.StatusBadRequest {
		t.Fatalf("blank status = %d body = %s", blank.Code, blank.Body.String())
	}
	assertErrorCode(t, blank.Body.Bytes(), "VALIDATION_ERROR")

	unknown := authedJSON(accounts, directory, http.MethodPatch, "/v1/me", token, `{"display_name":"Ada","email":"other@example.com"}`)
	if unknown.Code != http.StatusBadRequest {
		t.Fatalf("unknown status = %d body = %s", unknown.Code, unknown.Body.String())
	}

	updated := authedJSON(accounts, directory, http.MethodPatch, "/v1/me", token, `{"display_name":" Ada Lovelace "}`)
	if updated.Code != http.StatusOK {
		t.Fatalf("update status = %d body = %s", updated.Code, updated.Body.String())
	}
	again := authedJSON(accounts, directory, http.MethodGet, "/v1/me", token, "")
	if !strings.Contains(again.Body.String(), `"display_name":"Ada Lovelace"`) || !strings.Contains(again.Body.String(), `"email":"ada@example.com"`) {
		t.Fatalf("after update = %s", again.Body.String())
	}

	down := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v1/me", bytes.NewBufferString(`{"display_name":"Ada"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	NewMux(nil, nil, nil).ServeHTTP(down, req)
	if down.Code != http.StatusServiceUnavailable {
		t.Fatalf("down status = %d body = %s", down.Code, down.Body.String())
	}
	assertErrorCode(t, down.Body.Bytes(), "DATABASE_UNAVAILABLE")
	if strings.Contains(down.Body.String(), "alterad") || strings.Contains(down.Body.String(), "Ada") {
		t.Fatalf("503 claimed an update: %s", down.Body.String())
	}
	health := httptest.NewRecorder()
	NewMux(nil, nil, nil).ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health = %d %s", health.Code, health.Body.String())
	}
}

func loginAccessToken(t *testing.T, accounts *identity.Service, email, password string) string {
	t.Helper()
	rec := postJSON(accounts, "/v1/auth/register", `{"email":"`+email+`","password":"`+password+`"}`)
	if rec.Code != http.StatusCreated && rec.Code != http.StatusConflict {
		t.Fatalf("register status = %d body = %s", rec.Code, rec.Body.String())
	}
	logged := postJSON(accounts, "/v1/auth/login", `{"email":"`+email+`","password":"`+password+`"}`)
	if logged.Code != http.StatusOK {
		t.Fatalf("login status = %d body = %s", logged.Code, logged.Body.String())
	}
	var session struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(logged.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.AccessToken == "" {
		t.Fatal("login did not return an access token")
	}
	return session.AccessToken
}

func authedJSON(accounts *identity.Service, directory Directory, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	NewMux(nil, accounts, directory).ServeHTTP(rec, req)
	return rec
}
