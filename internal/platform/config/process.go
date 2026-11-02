package config

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// Relay is the outbox publisher process. It is not required for the API to boot.
type Relay struct {
	DatabaseURL string
	RabbitURL   string
}

// Worker is the media process. It is not required for the API to boot.
type Worker struct {
	DatabaseURL      string
	RabbitURL        string
	StorageEndpoint  string
	StorageRegion    string
	StorageBucket    string
	StorageAccessKey string
	StorageSecretKey string
	FFmpegBin        string
	FFprobeBin       string
	MaxAttempts      int
	Lease            time.Duration
}

// LoadRelay reads the database and the queue. Both are required for this process.
func LoadRelay(getenv func(string) string) (Relay, error) {
	cfg := Relay{
		DatabaseURL: strings.TrimSpace(getenv("DATABASE_URL")),
		RabbitURL:   strings.TrimSpace(getenv("RABBITMQ_URL")),
	}
	if cfg.DatabaseURL == "" || cfg.RabbitURL == "" {
		return Relay{}, errors.New("DATABASE_URL e RABBITMQ_URL são obrigatórios para o relay")
	}
	return cfg, nil
}

// LoadWorker reads the queue, the bucket and the FFmpeg binaries.
func LoadWorker(getenv func(string) string) (Worker, error) {
	cfg := Worker{
		DatabaseURL:      strings.TrimSpace(getenv("DATABASE_URL")),
		RabbitURL:        strings.TrimSpace(getenv("RABBITMQ_URL")),
		StorageEndpoint:  strings.TrimSpace(getenv("STORAGE_ENDPOINT")),
		StorageBucket:    strings.TrimSpace(getenv("STORAGE_BUCKET")),
		StorageAccessKey: strings.TrimSpace(getenv("STORAGE_ACCESS_KEY")),
		StorageSecretKey: strings.TrimSpace(getenv("STORAGE_SECRET_KEY")),
		StorageRegion:    strings.TrimSpace(getenv("STORAGE_REGION")),
		FFmpegBin:        strings.TrimSpace(getenv("FFMPEG_BIN")),
		FFprobeBin:       strings.TrimSpace(getenv("FFPROBE_BIN")),
		MaxAttempts:      3,
		Lease:            2 * time.Minute,
	}
	if cfg.StorageRegion == "" {
		cfg.StorageRegion = "us-east-1"
	}
	if cfg.FFmpegBin == "" {
		cfg.FFmpegBin = "ffmpeg"
	}
	if cfg.FFprobeBin == "" {
		cfg.FFprobeBin = "ffprobe"
	}
	if cfg.DatabaseURL == "" || cfg.RabbitURL == "" {
		return Worker{}, errors.New("DATABASE_URL e RABBITMQ_URL são obrigatórios para o worker")
	}
	if cfg.StorageEndpoint == "" || cfg.StorageBucket == "" || cfg.StorageAccessKey == "" || cfg.StorageSecretKey == "" {
		return Worker{}, errors.New("STORAGE_ENDPOINT, STORAGE_BUCKET, STORAGE_ACCESS_KEY e STORAGE_SECRET_KEY são obrigatórios para o worker")
	}
	if raw := strings.TrimSpace(getenv("PROCESSING_MAX_ATTEMPTS")); raw != "" {
		attempts, err := strconv.Atoi(raw)
		if err != nil || attempts < 1 {
			return Worker{}, errors.New("PROCESSING_MAX_ATTEMPTS deve ser um inteiro positivo")
		}
		cfg.MaxAttempts = attempts
	}
	if raw := strings.TrimSpace(getenv("PROCESSING_LEASE")); raw != "" {
		lease, err := time.ParseDuration(raw)
		if err != nil || lease <= 0 {
			return Worker{}, errors.New("PROCESSING_LEASE deve ser uma duração positiva")
		}
		cfg.Lease = lease
	}
	return cfg, nil
}
