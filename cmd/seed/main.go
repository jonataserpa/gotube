package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"example.com/gotube/internal/platform/postgres"
)

func main() {
	if err := run(os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "seed: %v\n", err)
		os.Exit(1)
	}
}

func run(getenv func(string) string) error {
	ctx := context.Background()
	pool, err := postgres.Open(ctx, getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer pool.Close()
	root, err := postgres.ModuleRoot()
	if err != nil {
		return err
	}
	return postgres.ApplySeed(ctx, pool, filepath.Join(root, "db", "seed", "dev.sql"))
}
