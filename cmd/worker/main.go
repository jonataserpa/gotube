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
	"example.com/gotube/internal/platform/storage"
	"example.com/gotube/internal/processing"
)

func main() {
	if err := run(os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "worker: %v\n", err)
		os.Exit(1)
	}
}

func run(getenv func(string) string) error {
	cfg, err := config.LoadWorker(getenv)
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
	objects, err := storage.New(cfg.StorageEndpoint, cfg.StorageRegion, cfg.StorageBucket, cfg.StorageAccessKey, cfg.StorageSecretKey)
	if err != nil {
		return err
	}
	worker, err := processing.NewWorker(pool, objects, processing.NewFFmpeg(cfg.FFmpegBin, cfg.FFprobeBin), cfg.Lease, cfg.MaxAttempts)
	if err != nil {
		return err
	}
	broker, err := queue.Dial(cfg.RabbitURL)
	if err != nil {
		return err
	}
	defer broker.Close()
	slog.Info("worker listening")
	go recoverLeases(ctx, worker)
	return broker.Consume(ctx, worker.Handle)
}

func recoverLeases(ctx context.Context, worker *processing.Worker) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		if err := worker.RecoverExpired(ctx); err != nil && ctx.Err() == nil {
			slog.Error("worker lease", "err", err.Error())
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
