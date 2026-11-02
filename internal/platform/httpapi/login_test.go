package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/gotube/internal/identity"
)

func TestLoginIssuesAccessTokenAndHttpOnlyRefresh(t *testing.T) {
	accounts, pool := newTestStack(t)
	if rec := postJSON(accounts, "/v1/auth/register", `{"email":"Member@Example.com","password":"correct-horse"}`); rec.Code != http.StatusCreated {
		t.Fatalf("register status = %d body = %s", rec.Code, rec.Body.String())
	}

	rec := postJSON(accounts, "/v1/auth/login", `{"email":" member@example.com ","password":"correct-horse"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "gotube_refresh" || !cookies[0].HttpOnly || cookies[0].Value == "" {
		t.Fatalf("cookie = %#v", cookies)
	}
	if strings.Contains(rec.Body.String(), cookies[0].Value) {
		t.Fatalf("refresh leaked into JSON: %s", rec.Body.String())
	}
	var session struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		User        struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	if session.TokenType != "Bearer" || session.ExpiresIn != 900 || session.User.Email != "member@example.com" || session.User.ID == "" {
		t.Fatalf("session = %#v", session)
	}
	account, err := identity.ParseAccessToken(testJWTSecret, session.AccessToken)
	if err != nil || account.Email != "member@example.com" || account.ID != session.User.ID {
		t.Fatalf("token = %#v err = %v", account, err)
	}

	unknown := postJSON(accounts, "/v1/auth/login", `{"email":"missing@example.com","password":"correct-horse"}`)
	wrong := postJSON(accounts, "/v1/auth/login", `{"email":"member@example.com","password":"wrong-password"}`)
	if _, err := pool.Exec(context.Background(), `UPDATE users SET status = 'SUSPENDED' WHERE email_normalized = $1`, "member@example.com"); err != nil {
		t.Fatal(err)
	}
	inactive := postJSON(accounts, "/v1/auth/login", `{"email":"member@example.com","password":"correct-horse"}`)
	for _, rec := range []*httptest.ResponseRecorder{unknown, wrong, inactive} {
		if rec.Code != http.StatusUnauthorized || len(rec.Result().Cookies()) != 0 {
			t.Fatalf("status = %d cookies = %d body = %s", rec.Code, len(rec.Result().Cookies()), rec.Body.String())
		}
		assertErrorCode(t, rec.Body.Bytes(), "INVALID_CREDENTIALS")
	}
	if unknown.Body.String() != wrong.Body.String() || unknown.Body.String() != inactive.Body.String() {
		t.Fatalf("credential errors differ:\n%s\n%s\n%s", unknown.Body.String(), wrong.Body.String(), inactive.Body.String())
	}
}
