package config

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	defaultReadHeaderTimeout = 2 * time.Second
	defaultReadTimeout       = 5 * time.Second
	defaultWriteTimeout      = 10 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
)

// Config is the process configuration required before the API listens.
type Config struct {
	Addr              string
	DatabaseURL       string
	JWTSecret         string
	CookieSecure      bool
	ExposeResetToken  bool
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	StorageEndpoint   string
	StorageRegion     string
	StorageBucket     string
	StorageAccessKey  string
	StorageSecretKey  string
	StorageMaxBytes   int64
	StorageCORSOrigin string
}

// Load reads environment values through getenv. Missing timeouts keep the
// defaults. An empty address or a non-positive explicit timeout is an error.
func Load(getenv func(string) string) (Config, error) {
	if getenv == nil {
		return Config{}, errors.New("getenv é obrigatório")
	}
	addr := strings.TrimSpace(getenv("API_ADDR"))
	if err := validateAddr(addr); err != nil {
		return Config{}, err
	}
	cookieSecure, err := parseBool(getenv, "AUTH_COOKIE_SECURE")
	if err != nil {
		return Config{}, err
	}
	exposeReset, err := parseBool(getenv, "AUTH_EXPOSE_RESET_TOKEN")
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Addr:              addr,
		DatabaseURL:       strings.TrimSpace(getenv("DATABASE_URL")),
		JWTSecret:         strings.TrimSpace(getenv("AUTH_JWT_SECRET")),
		CookieSecure:      cookieSecure,
		ExposeResetToken:  exposeReset,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		ReadTimeout:       defaultReadTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
		ShutdownTimeout:   defaultShutdownTimeout,
	}
	fields := []struct {
		key  string
		dest *time.Duration
	}{
		{"HTTP_READ_HEADER_TIMEOUT", &cfg.ReadHeaderTimeout},
		{"HTTP_READ_TIMEOUT", &cfg.ReadTimeout},
		{"HTTP_WRITE_TIMEOUT", &cfg.WriteTimeout},
		{"HTTP_IDLE_TIMEOUT", &cfg.IdleTimeout},
		{"HTTP_SHUTDOWN_TIMEOUT", &cfg.ShutdownTimeout},
	}
	for _, field := range fields {
		if err := applyTimeout(getenv, field.key, field.dest); err != nil {
			return Config{}, err
		}
	}
	if err := applyStorage(getenv, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func applyStorage(getenv func(string) string, cfg *Config) error {
	cfg.StorageEndpoint = strings.TrimSpace(getenv("STORAGE_ENDPOINT"))
	cfg.StorageBucket = strings.TrimSpace(getenv("STORAGE_BUCKET"))
	cfg.StorageAccessKey = strings.TrimSpace(getenv("STORAGE_ACCESS_KEY"))
	cfg.StorageSecretKey = strings.TrimSpace(getenv("STORAGE_SECRET_KEY"))
	cfg.StorageRegion = strings.TrimSpace(getenv("STORAGE_REGION"))
	cfg.StorageCORSOrigin = strings.TrimSpace(getenv("STORAGE_CORS_ORIGIN"))
	if cfg.StorageRegion == "" {
		cfg.StorageRegion = "us-east-1"
	}
	configured := 0
	for _, value := range []string{cfg.StorageEndpoint, cfg.StorageBucket, cfg.StorageAccessKey, cfg.StorageSecretKey} {
		if value != "" {
			configured++
		}
	}
	if configured == 0 {
		return nil
	}
	if configured != 4 {
		return errors.New("STORAGE_ENDPOINT, STORAGE_BUCKET, STORAGE_ACCESS_KEY e STORAGE_SECRET_KEY devem ser definidos juntos")
	}
	rawMax := strings.TrimSpace(getenv("STORAGE_MAX_BYTES"))
	if rawMax == "" {
		cfg.StorageMaxBytes = 512 * 1024 * 1024
		return nil
	}
	maxBytes, err := strconv.ParseInt(rawMax, 10, 64)
	if err != nil || maxBytes < 1 {
		return errors.New("STORAGE_MAX_BYTES deve ser um inteiro positivo")
	}
	cfg.StorageMaxBytes = maxBytes
	return nil
}

func validateAddr(addr string) error {
	if addr == "" {
		return errors.New("API_ADDR é obrigatório")
	}
	if strings.Contains(addr, "://") {
		return errors.New("API_ADDR deve ser host:port, não uma URL")
	}
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("API_ADDR: %w", err)
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return errors.New("API_ADDR precisa de uma porta entre 1 e 65535")
	}
	return nil
}

func parseBool(getenv func(string) string, key string) (bool, error) {
	raw := strings.TrimSpace(getenv(key))
	if raw == "" || raw == "false" {
		return false, nil
	}
	if raw == "true" {
		return true, nil
	}
	return false, fmt.Errorf("%s deve ser true ou false", key)
}

func applyTimeout(getenv func(string) string, key string, dest *time.Duration) error {
	raw := strings.TrimSpace(getenv(key))
	if raw == "" {
		return nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fmt.Errorf("%s deve ser uma duração positiva", key)
	}
	*dest = value
	return nil
}
