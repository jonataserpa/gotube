package httpapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAPIDocumentsHealthVideosAndAccountRoutes(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(moduleRoot(t), "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, required := range []string{
		"\n  /healthz:\n",
		"\n  /v1/videos:\n",
		"\n  /v1/auth/register:\n",
		"\n  /v1/auth/login:\n",
		"\n  /v1/auth/refresh:\n",
		"\n  /v1/auth/logout:\n",
		"\n  /v1/auth/password/forgot:\n",
		"\n  /v1/auth/password/reset:\n",
		"INVALID_LIMIT",
		"INVALID_CURSOR",
		"DATABASE_UNAVAILABLE",
		"AUTH_UNAVAILABLE",
		"VALIDATION_ERROR",
		"EMAIL_ALREADY_REGISTERED",
		"INVALID_CREDENTIALS",
		"INVALID_REFRESH",
		"INVALID_RESET_TOKEN",
		"\n  /v1/me:\n",
		"\n  /v1/me/channel:\n",
		"\n  /v1/channels:\n",
		"\n  /v1/channels/{handle}:\n",
		"\n  /v1/categories:\n",
		"UNAUTHENTICATED",
		"INVALID_CHANNEL",
		"CHANNEL_FORBIDDEN",
		"CHANNEL_NOT_FOUND",
		"CHANNEL_ALREADY_EXISTS",
		"HANDLE_ALREADY_REGISTERED",
		"createDraft",
		"\n  /v1/videos/{id}/upload-sessions:\n",
		"\n  /v1/upload-sessions/{id}/complete:\n",
		"CHANNEL_REQUIRED",
		"VIDEO_FORBIDDEN",
		"VIDEO_NOT_FOUND",
		"VIDEO_NOT_WAITING",
		"UPLOAD_REJECTED",
		"UPLOAD_SESSION_NOT_FOUND",
		"WAITING_UPLOAD",
		"\n  /v1/videos/{id}/processing:\n",
		"\n  /v1/videos/{id}/retry-processing:\n",
		"PROCESSING_NOT_FAILED",
		"readProcessing",
		"retryProcessing",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("OpenAPI missing %s", required)
		}
	}
	for _, forbidden := range []string{"/watch", "/playback", "/admin"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("OpenAPI documents unavailable endpoint %s", forbidden)
		}
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
