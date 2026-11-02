package httpapi

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"example.com/gotube/internal/platform/config"
)

func TestServeStopsAcceptingAfterCancel(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, testConfig(), ln, nil, nil, nil)
	}()

	url := "http://" + ln.Addr().String() + "/healthz"
	if err := waitForHealth(url); err != nil {
		cancel()
		t.Fatal(err)
	}
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown exceeded the configured deadline")
	}

	conn, err := net.DialTimeout("tcp", ln.Addr().String(), 200*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		t.Fatal("listener still accepted a connection")
	}
}

func testConfig() config.Config {
	return config.Config{
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       time.Second,
		WriteTimeout:      time.Second,
		IdleTimeout:       time.Second,
		ShutdownTimeout:   time.Second,
	}
}

func waitForHealth(url string) error {
	deadline := time.Now().Add(2 * time.Second)
	var last error
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			last = errStatus(resp.StatusCode)
		} else {
			last = err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return last
}

type errStatus int

func (e errStatus) Error() string { return http.StatusText(int(e)) }
