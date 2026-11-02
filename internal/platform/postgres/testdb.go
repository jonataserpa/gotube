package postgres

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const catalogTestDatabase = "gotube_catalog_test"

// OpenCatalogTestDB returns an empty database reserved for catalog tests.
// DATABASE_URL is required. The caller applies migrations itself.
func OpenCatalogTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	raw := os.Getenv("DATABASE_URL")
	if raw == "" {
		t.Fatal("DATABASE_URL é obrigatório para o teste de catálogo")
	}
	ctx := context.Background()
	admin, err := Open(ctx, raw)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := admin.Acquire(ctx)
	if err != nil {
		admin.Close()
		t.Fatal(err)
	}
	if _, err := lock.Exec(ctx, "SELECT pg_advisory_lock($1)", int64(84201)); err != nil {
		lock.Release()
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = lock.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", int64(84201))
		lock.Release()
		admin.Close()
	})

	var exists bool
	if err := admin.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", catalogTestDatabase).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if !exists {
		if _, err := admin.Exec(ctx, "CREATE DATABASE "+catalogTestDatabase); err != nil {
			t.Fatal(err)
		}
	}

	testURL, err := swapDatabase(raw, catalogTestDatabase)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := Open(ctx, testURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if _, err := pool.Exec(ctx, "DROP SCHEMA IF EXISTS public CASCADE"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "CREATE SCHEMA public"); err != nil {
		t.Fatal(err)
	}
	return pool
}

func swapDatabase(raw string, name string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("DATABASE_URL: %w", err)
	}
	parsed.Path = "/" + name
	return parsed.String(), nil
}
