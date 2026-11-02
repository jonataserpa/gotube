package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRunDoesNotListenWhenConfigIsInvalid(t *testing.T) {
	done := make(chan error, 1)
	go func() {
		done <- run(func(key string) string {
			switch key {
			case "API_ADDR":
				return "127.0.0.1:18081"
			case "HTTP_SHUTDOWN_TIMEOUT":
				return "0s"
			default:
				return ""
			}
		})
	}()

	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "HTTP_SHUTDOWN_TIMEOUT") {
			t.Fatalf("expected configuration error, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("run blocked; the listener may have started with invalid configuration")
	}
}

func TestAPIServesHealthWithoutQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- serve(ctx, func(key string) string {
			switch key {
			case "API_ADDR":
				return "127.0.0.1:18087"
			case "HTTP_SHUTDOWN_TIMEOUT":
				return "200ms"
			default:
				return ""
			}
		})
	}()
	var body string
	deadline := time.Now().Add(2 * time.Second)
	for {
		response, err := http.Get("http://127.0.0.1:18087/healthz")
		if err == nil {
			raw, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			body = string(raw)
			if response.StatusCode == http.StatusOK {
				break
			}
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatalf("health without queue: %v body=%s", err, body)
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("api did not stop")
	}
	if !strings.Contains(body, `"status":"ok"`) {
		t.Fatalf("health = %s", body)
	}
}
