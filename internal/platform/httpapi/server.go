package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"example.com/gotube/internal/platform/config"
)

const maxHeaderBytes = 1 << 20

// Serve owns the listener until ctx is cancelled or Accept fails.
// Cancellation stops new connections and waits up to ShutdownTimeout.
func Serve(ctx context.Context, cfg config.Config, ln net.Listener, videos PublicVideos, accounts Accounts, directory Directory) error {
	if ln == nil {
		return errors.New("listener é obrigatório")
	}
	server := &http.Server{
		Handler:           NewMux(videos, accounts, directory),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}
	errCh := make(chan error, 1)
	go func() {
		err := server.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return <-errCh
}
