package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadRejectsMissingAddress(t *testing.T) {
	_, err := Load(func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "API_ADDR") {
		t.Fatalf("expected API_ADDR error, got %v", err)
	}
}

func TestLoadRejectsURLAndNonPositiveTimeout(t *testing.T) {
	_, err := Load(envWith("API_ADDR", "http://127.0.0.1:8080"))
	if err == nil || !strings.Contains(err.Error(), "API_ADDR") {
		t.Fatalf("expected URL rejection, got %v", err)
	}

	_, err = Load(envWith("API_ADDR", ":8080", "HTTP_SHUTDOWN_TIMEOUT", "0s"))
	if err == nil || !strings.Contains(err.Error(), "HTTP_SHUTDOWN_TIMEOUT") {
		t.Fatalf("expected timeout error, got %v", err)
	}

	_, err = Load(envWith("API_ADDR", ":8080", "HTTP_READ_TIMEOUT", "nope"))
	if err == nil || !strings.Contains(err.Error(), "HTTP_READ_TIMEOUT") {
		t.Fatalf("expected invalid duration error, got %v", err)
	}
}

func TestLoadUsesDefaultsAndExplicitTimeout(t *testing.T) {
	cfg, err := Load(envWith("API_ADDR", "127.0.0.1:8080"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != "127.0.0.1:8080" {
		t.Fatalf("addr = %q", cfg.Addr)
	}
	if cfg.ReadHeaderTimeout != 2*time.Second || cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}

	cfg, err = Load(envWith("API_ADDR", ":8080", "HTTP_IDLE_TIMEOUT", "30s"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.IdleTimeout != 30*time.Second {
		t.Fatalf("idle = %s", cfg.IdleTimeout)
	}
}

func envWith(pairs ...string) func(string) string {
	values := map[string]string{}
	for i := 0; i < len(pairs); i += 2 {
		values[pairs[i]] = pairs[i+1]
	}
	return func(key string) string { return values[key] }
}
