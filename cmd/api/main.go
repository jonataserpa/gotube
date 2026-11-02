package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"example.com/gotube/internal/account"
	"example.com/gotube/internal/catalog"
	"example.com/gotube/internal/identity"
	"example.com/gotube/internal/platform/config"
	"example.com/gotube/internal/platform/httpapi"
	"example.com/gotube/internal/platform/postgres"
	"example.com/gotube/internal/platform/storage"
)

func main() {
	if err := run(os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "api: %v\n", err)
		os.Exit(1)
	}
}

func run(getenv func(string) string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return serve(ctx, getenv)
}

func serve(ctx context.Context, getenv func(string) string) error {
	cfg, err := config.Load(getenv)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	videos, accounts, directory, closePool, err := openStores(cfg)
	if err != nil {
		_ = ln.Close()
		return err
	}
	defer closePool()
	slog.Info("api listening", "addr", ln.Addr().String())
	return httpapi.Serve(ctx, cfg, ln, videos, accounts, directory)
}

func openStores(cfg config.Config) (httpapi.PublicVideos, httpapi.Accounts, httpapi.Directory, func(), error) {
	if cfg.DatabaseURL == "" {
		return nil, nil, nil, func() {}, nil
	}
	pool, err := postgres.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	accounts := identity.NewService(pool, identity.Settings{
		JWTSecret:        cfg.JWTSecret,
		CookieSecure:     cfg.CookieSecure,
		ExposeResetToken: cfg.ExposeResetToken,
	})
	directory := account.NewService(pool, cfg.JWTSecret)
	if cfg.StorageEndpoint != "" {
		objects, err := storage.New(cfg.StorageEndpoint, cfg.StorageRegion, cfg.StorageBucket, cfg.StorageAccessKey, cfg.StorageSecretKey)
		if err != nil {
			pool.Close()
			return nil, nil, nil, nil, err
		}
		directory.UseStorage(objects, cfg.StorageMaxBytes)
		if err := objects.Ensure(context.Background(), cfg.StorageCORSOrigin); err != nil {
			slog.Warn("storage bucket was not prepared", "err", err)
		}
	}
	return catalog.NewStore(pool), accounts, directory, func() { pool.Close() }, nil
}
