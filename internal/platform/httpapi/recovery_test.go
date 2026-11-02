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

func TestPasswordRecoveryIsGenericAndSingleUse(t *testing.T) {
	accounts, pool := newTestStack(t)
	register(t, accounts, "recover@example.com")
	cookie := login(t, accounts, "recover@example.com")

	known := postAuth(accounts, "/v1/auth/password/forgot", `{"email":"Recover@Example.com"}`, "")
	unknown := postAuth(accounts, "/v1/auth/password/forgot", `{"email":"missing@example.com"}`, "")
	if known.Code != http.StatusAccepted || unknown.Code != http.StatusAccepted {
		t.Fatalf("known = %d %s unknown = %d %s", known.Code, known.Body.String(), unknown.Code, unknown.Body.String())
	}
	if known.Body.String() != unknown.Body.String() || strings.Contains(known.Body.String(), "reset_token") {
		t.Fatalf("forgot leaked account existence: %s vs %s", known.Body.String(), unknown.Body.String())
	}

	exposed := identity.NewService(pool, identity.Settings{JWTSecret: testJWTSecret, ExposeResetToken: true})
	hidden := postAuth(exposed, "/v1/auth/password/forgot", `{"email":"missing@example.com"}`, "")
	if hidden.Code != http.StatusAccepted || strings.Contains(hidden.Body.String(), "reset_token") {
		t.Fatalf("missing account exposed a token: %d %s", hidden.Code, hidden.Body.String())
	}
	issued := postAuth(exposed, "/v1/auth/password/forgot", `{"email":"recover@example.com"}`, "")
	token := resetTokenFrom(t, issued.Body.Bytes())

	short := postAuth(accounts, "/v1/auth/password/reset", `{"token":"`+token+`","password":"1234567"}`, "")
	if short.Code != http.StatusBadRequest {
		t.Fatalf("short reset = %d %s", short.Code, short.Body.String())
	}
	assertErrorCode(t, short.Body.Bytes(), "VALIDATION_ERROR")

	reset := postAuth(accounts, "/v1/auth/password/reset", `{"token":"`+token+`","password":"new-password"}`, "")
	if reset.Code != http.StatusNoContent {
		t.Fatalf("reset = %d %s", reset.Code, reset.Body.String())
	}
	again := postAuth(accounts, "/v1/auth/password/reset", `{"token":"`+token+`","password":"another-password"}`, "")
	if again.Code != http.StatusBadRequest {
		t.Fatalf("second reset = %d %s", again.Code, again.Body.String())
	}
	assertErrorCode(t, again.Body.Bytes(), "INVALID_RESET_TOKEN")
	if old := postAuth(accounts, "/v1/auth/login", `{"email":"recover@example.com","password":"correct-horse"}`, ""); old.Code != http.StatusUnauthorized {
		t.Fatalf("old password still works: %d", old.Code)
	}
	if kept := postAuth(accounts, "/v1/auth/login", `{"email":"recover@example.com","password":"new-password"}`, ""); kept.Code != http.StatusOK {
		t.Fatalf("new password = %d %s", kept.Code, kept.Body.String())
	}
	if reused := postAuth(accounts, "/v1/auth/refresh", "", cookie); reused.Code != http.StatusUnauthorized {
		t.Fatalf("refresh survived password reset: %d", reused.Code)
	}

	expiredToken := resetTokenFrom(t, postAuth(exposed, "/v1/auth/password/forgot", `{"email":"recover@example.com"}`, "").Body.Bytes())
	if _, err := pool.Exec(context.Background(), `
		UPDATE password_resets
		SET expires_at = now() - interval '1 minute'
		WHERE user_id = (SELECT id FROM users WHERE email_normalized = 'recover@example.com')
			AND used_at IS NULL
	`); err != nil {
		t.Fatal(err)
	}
	expired := postAuth(accounts, "/v1/auth/password/reset", `{"token":"`+expiredToken+`","password":"third-password"}`, "")
	if expired.Code != http.StatusBadRequest {
		t.Fatalf("expired = %d %s", expired.Code, expired.Body.String())
	}
	if still := postAuth(accounts, "/v1/auth/login", `{"email":"recover@example.com","password":"new-password"}`, ""); still.Code != http.StatusOK {
		t.Fatalf("expired reset changed the password: %d", still.Code)
	}

	down := httptest.NewRecorder()
	NewMux(nil, nil, nil).ServeHTTP(down, httptest.NewRequest(http.MethodPost, "/v1/auth/password/reset", strings.NewReader(`{"token":"abc","password":"new-password"}`)))
	if down.Code != http.StatusServiceUnavailable || strings.Contains(down.Body.String(), "alterad") {
		t.Fatalf("down = %d %s", down.Code, down.Body.String())
	}
	assertErrorCode(t, down.Body.Bytes(), "DATABASE_UNAVAILABLE")
}

func resetTokenFrom(t *testing.T, body []byte) string {
	t.Helper()
	var payload struct {
		ResetToken string `json:"reset_token"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.ResetToken == "" {
		t.Fatalf("reset token missing: %s", body)
	}
	return payload.ResetToken
}
