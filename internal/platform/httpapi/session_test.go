package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRefreshRotatesAndLogoutRevokesOnlyThatSession(t *testing.T) {
	accounts, pool := newTestStack(t)
	register(t, accounts, "one@example.com")
	first := login(t, accounts, "one@example.com")
	second := login(t, accounts, "one@example.com")

	rotated := postAuth(accounts, "/v1/auth/refresh", "", first)
	if rotated.Code != http.StatusOK || len(rotated.Result().Cookies()) != 1 || !rotated.Result().Cookies()[0].HttpOnly {
		t.Fatalf("refresh = %d cookies = %#v body = %s", rotated.Code, rotated.Result().Cookies(), rotated.Body.String())
	}
	next := rotated.Result().Cookies()[0].Value
	if next == first || bytes.Contains(rotated.Body.Bytes(), []byte(next)) {
		t.Fatalf("refresh did not rotate or leaked the cookie: %s", rotated.Body.String())
	}
	userID := userIDFrom(t, rotated.Body.Bytes())

	replay := postAuth(accounts, "/v1/auth/refresh", "", first)
	if replay.Code != http.StatusUnauthorized {
		t.Fatalf("immediate replay = %d %s", replay.Code, replay.Body.String())
	}
	if active := countActiveSessions(t, pool, userID); active != 2 {
		t.Fatalf("replay inside the race window changed live sessions, active = %d", active)
	}

	if _, err := pool.Exec(context.Background(), `
		UPDATE refresh_sessions
		SET revoked_at = now() - interval '30 seconds'
		WHERE user_id = $1 AND revoked_at IS NOT NULL
	`, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `
		UPDATE refresh_sessions
		SET created_at = now() - interval '30 seconds'
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID); err != nil {
		t.Fatal(err)
	}
	reused := postAuth(accounts, "/v1/auth/refresh", "", first)
	if reused.Code != http.StatusUnauthorized {
		t.Fatalf("reuse = %d %s", reused.Code, reused.Body.String())
	}
	assertErrorCode(t, reused.Body.Bytes(), "INVALID_REFRESH")
	if active := countActiveSessions(t, pool, userID); active != 1 {
		t.Fatalf("reuse affected the other device, active = %d", active)
	}
	if secondStill := postAuth(accounts, "/v1/auth/refresh", "", second); secondStill.Code != http.StatusOK {
		t.Fatalf("second device after reuse = %d %s", secondStill.Code, secondStill.Body.String())
	}

	register(t, accounts, "two@example.com")
	fresh := login(t, accounts, "two@example.com")
	unknown := postAuth(accounts, "/v1/auth/refresh", "", "not-a-real-refresh-token")
	if unknown.Code != http.StatusUnauthorized {
		t.Fatalf("unknown = %d %s", unknown.Code, unknown.Body.String())
	}
	if kept := postAuth(accounts, "/v1/auth/refresh", "", fresh); kept.Code != http.StatusOK {
		t.Fatalf("unknown cookie affected another account: %d %s", kept.Code, kept.Body.String())
	}

	register(t, accounts, "three@example.com")
	other := login(t, accounts, "three@example.com")
	primary := login(t, accounts, "three@example.com")
	logout := postAuth(accounts, "/v1/auth/logout", "", primary)
	if logout.Code != http.StatusNoContent {
		t.Fatalf("logout = %d %s", logout.Code, logout.Body.String())
	}
	if cleared := logout.Result().Cookies(); len(cleared) != 1 || cleared[0].Name != "gotube_refresh" || cleared[0].MaxAge > 0 {
		t.Fatalf("logout cookie = %#v", cleared)
	}
	if gone := postAuth(accounts, "/v1/auth/refresh", "", primary); gone.Code != http.StatusUnauthorized {
		t.Fatalf("logged out refresh = %d", gone.Code)
	}
	if stays := postAuth(accounts, "/v1/auth/refresh", "", other); stays.Code != http.StatusOK {
		t.Fatalf("other device = %d %s", stays.Code, stays.Body.String())
	}
	if empty := postAuth(accounts, "/v1/auth/logout", "", ""); empty.Code != http.StatusNoContent {
		t.Fatalf("logout without cookie = %d", empty.Code)
	}

	register(t, accounts, "race@example.com")
	raceCookie := login(t, accounts, "race@example.com")
	codes := make([]int, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func(i int) {
			defer wg.Done()
			codes[i] = postAuth(accounts, "/v1/auth/refresh", "", raceCookie).Code
		}(i)
	}
	wg.Wait()
	sort.Ints(codes)
	if codes[0] != http.StatusOK || codes[1] != http.StatusUnauthorized {
		t.Fatalf("race codes = %v", codes)
	}
	var raceUser string
	if err := pool.QueryRow(context.Background(), `SELECT id::text FROM users WHERE email_normalized = 'race@example.com'`).Scan(&raceUser); err != nil {
		t.Fatal(err)
	}
	if active := countActiveSessions(t, pool, raceUser); active != 1 {
		t.Fatalf("race left %d active sessions", active)
	}
}

func register(t *testing.T, accounts Accounts, email string) {
	t.Helper()
	rec := postAuth(accounts, "/v1/auth/register", `{"email":"`+email+`","password":"correct-horse"}`, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("register %s = %d %s", email, rec.Code, rec.Body.String())
	}
}

func login(t *testing.T, accounts Accounts, email string) string {
	t.Helper()
	rec := postAuth(accounts, "/v1/auth/login", `{"email":"`+email+`","password":"correct-horse"}`, "")
	if rec.Code != http.StatusOK || len(rec.Result().Cookies()) != 1 {
		t.Fatalf("login %s = %d %s", email, rec.Code, rec.Body.String())
	}
	return rec.Result().Cookies()[0].Value
}

func postAuth(accounts Accounts, path string, body string, cookie string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "gotube_refresh", Value: cookie})
	}
	rec := httptest.NewRecorder()
	NewMux(nil, accounts, nil).ServeHTTP(rec, req)
	return rec
}

func userIDFrom(t *testing.T, body []byte) string {
	t.Helper()
	var payload struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.User.ID == "" {
		t.Fatalf("user id missing: %s", body)
	}
	return payload.User.ID
}

func countActiveSessions(t *testing.T, pool *pgxpool.Pool, userID string) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(), `
		SELECT count(*) FROM refresh_sessions
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
