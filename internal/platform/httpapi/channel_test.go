package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"example.com/gotube/internal/account"
	"example.com/gotube/internal/platform/postgres"
)

func TestChannelCreateEditAndPublicRead(t *testing.T) {
	accounts, pool := newTestStack(t)
	root, err := postgres.ModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := postgres.ApplySeed(context.Background(), pool, filepath.Join(root, "db", "seed", "dev.sql")); err != nil {
		t.Fatal(err)
	}
	directory := account.NewService(pool, testJWTSecret)
	owner := loginAccessToken(t, accounts, "owner@example.com", "correct-horse")
	other := loginAccessToken(t, accounts, "other@example.com", "correct-horse")

	absent := authedJSON(accounts, directory, http.MethodGet, "/v1/me/channel", owner, "")
	if absent.Code != http.StatusOK || !strings.Contains(absent.Body.String(), `"channel":null`) {
		t.Fatalf("absent = %d %s", absent.Code, absent.Body.String())
	}

	invalid := authedJSON(accounts, directory, http.MethodPost, "/v1/channels", owner, `{"handle":"A B","name":"Canal"}`)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid = %d %s", invalid.Code, invalid.Body.String())
	}
	stillAbsent := authedJSON(accounts, directory, http.MethodGet, "/v1/me/channel", owner, "")
	if !strings.Contains(stillAbsent.Body.String(), `"channel":null`) {
		t.Fatalf("invalid handle created a channel: %s", stillAbsent.Body.String())
	}

	created := authedJSON(accounts, directory, http.MethodPost, "/v1/channels", owner, `{"handle":" Ada-Channel ","name":" Canal Ada ","description":"Aulas"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create = %d %s", created.Code, created.Body.String())
	}
	if strings.Contains(created.Body.String(), "owner@example.com") {
		t.Fatalf("create leaked email: %s", created.Body.String())
	}
	owned := authedJSON(accounts, directory, http.MethodGet, "/v1/me/channel", owner, "")
	if !strings.Contains(owned.Body.String(), `"handle":"ada-channel"`) || !strings.Contains(owned.Body.String(), `"name":"Canal Ada"`) {
		t.Fatalf("owned = %s", owned.Body.String())
	}

	again := authedJSON(accounts, directory, http.MethodPost, "/v1/channels", owner, `{"handle":"outro","name":"Outro"}`)
	if again.Code != http.StatusConflict {
		t.Fatalf("second = %d %s", again.Code, again.Body.String())
	}
	assertErrorCode(t, again.Body.Bytes(), "CHANNEL_ALREADY_EXISTS")

	taken := authedJSON(accounts, directory, http.MethodPost, "/v1/channels", other, `{"handle":"  ADA-CHANNEL  ","name":"Copia"}`)
	if taken.Code != http.StatusConflict {
		t.Fatalf("taken = %d %s", taken.Code, taken.Body.String())
	}
	assertErrorCode(t, taken.Body.Bytes(), "HANDLE_ALREADY_REGISTERED")

	public := httptest.NewRecorder()
	NewMux(nil, accounts, directory).ServeHTTP(public, httptest.NewRequest(http.MethodGet, "/v1/channels/ada-channel", nil))
	if public.Code != http.StatusOK || strings.Contains(public.Body.String(), "@") {
		t.Fatalf("public = %d %s", public.Code, public.Body.String())
	}

	id := channelID(t, owned.Body.String())
	renamed := authedJSON(accounts, directory, http.MethodPatch, "/v1/channels/"+id, owner, `{"name":"Canal Novo"}`)
	if renamed.Code != http.StatusOK {
		t.Fatalf("rename = %d %s", renamed.Code, renamed.Body.String())
	}
	after := httptest.NewRecorder()
	NewMux(nil, accounts, directory).ServeHTTP(after, httptest.NewRequest(http.MethodGet, "/v1/channels/ada-channel", nil))
	if !strings.Contains(after.Body.String(), `"name":"Canal Novo"`) {
		t.Fatalf("public name = %s", after.Body.String())
	}

	forbidden := authedJSON(accounts, directory, http.MethodPatch, "/v1/channels/"+id, other, `{"name":"Invadido"}`)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("other = %d %s", forbidden.Code, forbidden.Body.String())
	}
	assertErrorCode(t, forbidden.Body.Bytes(), "CHANNEL_FORBIDDEN")

	lab := authedJSON(accounts, directory, http.MethodPatch, "/v1/channels/00000000-0000-4000-8000-000000000001", owner, `{"name":"Tomado"}`)
	if lab.Code != http.StatusNotFound && lab.Code != http.StatusForbidden {
		t.Fatalf("lab status = %d %s", lab.Code, lab.Body.String())
	}
	if lab.Code != http.StatusForbidden {
		t.Fatalf("lab = %d %s", lab.Code, lab.Body.String())
	}

	missing := authedJSON(accounts, directory, http.MethodPatch, "/v1/channels/00000000-0000-4000-8000-000000000099", owner, `{"name":"Nao"}`)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing = %d %s", missing.Code, missing.Body.String())
	}

	down := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/channels", bytes.NewBufferString(`{"handle":"novo","name":"Novo"}`))
	req.Header.Set("Authorization", "Bearer "+owner)
	NewMux(nil, nil, nil).ServeHTTP(down, req)
	if down.Code != http.StatusServiceUnavailable {
		t.Fatalf("down = %d %s", down.Code, down.Body.String())
	}
	if strings.Contains(down.Body.String(), `"handle"`) {
		t.Fatalf("503 created a channel: %s", down.Body.String())
	}
}

func channelID(t *testing.T, body string) string {
	t.Helper()
	const marker = `"id":"`
	start := strings.Index(body, marker)
	if start < 0 {
		t.Fatalf("no id in %s", body)
	}
	rest := body[start+len(marker):]
	end := strings.Index(rest, `"`)
	if end < 0 {
		t.Fatalf("id not closed in %s", body)
	}
	return rest[:end]
}
