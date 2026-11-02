package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/gotube/internal/platform/config"
	"example.com/gotube/internal/platform/postgres"
	"example.com/gotube/internal/platform/queue"
	"example.com/gotube/internal/processing"
)

func main() {
	if err := run(os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "outbox: %v\n", err)
		os.Exit(1)
	}
}

func run(getenv func(string) string) error {
	cfg, err := config.LoadRelay(getenv)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	pool, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	broker, err := queue.Dial(cfg.RabbitURL)
	if err != nil {
		return err
	}
	defer broker.Close()
	relay, err := processing.NewRelay(pool, broker, 30*time.Second)
	if err != nil {
		return err
	}
	slog.Info("outbox relay listening")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		if _, err := relay.PublishOnce(ctx); err != nil && ctx.Err() == nil {
			slog.Error("outbox publish", "err", err.Error())
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
