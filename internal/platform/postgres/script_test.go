package postgres_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/gotube/internal/platform/postgres"
)

func TestDevSeedIsLabeledAndSeparateFromMigration(t *testing.T) {
	root, err := postgres.ModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	seed, err := os.ReadFile(filepath.Join(root, "db", "seed", "dev.sql"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(seed)
	if !strings.Contains(text, "dado de desenvolvimento") {
		t.Fatal("seed is not labeled as development data")
	}
	if !strings.Contains(text, "PUBLIC") || !strings.Contains(text, "PUBLISHED") {
		t.Fatal("seed has no public published video")
	}
	if !strings.Contains(text, "DRAFT") && !strings.Contains(text, "PRIVATE") {
		t.Fatal("seed has no video excluded by the public list")
	}
	migrate, err := os.ReadFile(filepath.Join(root, "cmd", "migrate", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(migrate), "dev.sql") || strings.Contains(string(migrate), "ApplySeed") {
		t.Fatal("migration command also applies the seed")
	}
	seedCmd, err := os.ReadFile(filepath.Join(root, "cmd", "seed", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(seedCmd), "ApplyMigrations") {
		t.Fatal("seed command also applies migrations")
	}
	api, err := os.ReadFile(filepath.Join(root, "cmd", "api", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(api), "ApplyMigrations") || strings.Contains(string(api), "db/migrations") {
		t.Fatal("API boot applies migrations")
	}
	if strings.Contains(text, "password") || strings.Contains(text, "INSERT INTO users") {
		t.Fatal("seed creates an account or stores a password")
	}
}

func TestMigrateAndSeedOnEmptyDatabase(t *testing.T) {
	ctx := context.Background()
	pool := postgres.OpenCatalogTestDB(t)
	root, err := postgres.ModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := postgres.ApplyMigrations(ctx, pool, filepath.Join(root, "db", "migrations")); err != nil {
		t.Fatal(err)
	}
	var tables int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name IN ('channels', 'videos')
	`).Scan(&tables); err != nil {
		t.Fatal(err)
	}
	if tables != 2 {
		t.Fatalf("tables = %d", tables)
	}
	if err := postgres.ApplySeed(ctx, pool, filepath.Join(root, "db", "seed", "dev.sql")); err != nil {
		t.Fatal(err)
	}
	var publicCount, hiddenCount int
	if err := pool.QueryRow(ctx, `
		SELECT
			count(*) FILTER (WHERE visibility = 'PUBLIC' AND publication_status = 'PUBLISHED'),
			count(*) FILTER (WHERE NOT (visibility = 'PUBLIC' AND publication_status = 'PUBLISHED'))
		FROM videos
	`).Scan(&publicCount, &hiddenCount); err != nil {
		t.Fatal(err)
	}
	if publicCount < 1 || hiddenCount < 1 {
		t.Fatalf("public = %d hidden = %d", publicCount, hiddenCount)
	}
	var identityTables, displayName, description, ownerColumn, categoryTables, labWithoutOwner int
	if err := pool.QueryRow(ctx, `
		SELECT
			(
				SELECT count(*) FROM information_schema.tables
				WHERE table_schema = 'public'
					AND table_name IN ('users', 'refresh_sessions', 'password_resets')
			),
			(
				SELECT count(*) FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'display_name'
			),
			(
				SELECT count(*) FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'channels' AND column_name = 'description'
			),
			(
				SELECT count(*) FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'channels' AND column_name = 'owner_id'
			),
			(
				SELECT count(*) FROM information_schema.tables
				WHERE table_schema = 'public'
					AND table_name IN ('categories', 'channel_categories')
			),
			(
				SELECT count(*) FROM channels
				WHERE handle = 'laboratorio' AND owner_id IS NULL
			)
	`).Scan(&identityTables, &displayName, &description, &ownerColumn, &categoryTables, &labWithoutOwner); err != nil {
		t.Fatal(err)
	}
	if identityTables != 3 || displayName != 1 || description != 1 || ownerColumn != 1 || categoryTables != 2 || labWithoutOwner != 1 {
		t.Fatalf("identity=%d display_name=%d description=%d owner_id=%d categories=%d lab=%d", identityTables, displayName, description, ownerColumn, categoryTables, labWithoutOwner)
	}
	var processingColumn, uploadTables, readyPublished, waitingDraft int
	if err := pool.QueryRow(ctx, `
		SELECT
			(
				SELECT count(*) FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'videos' AND column_name = 'processing_status'
			),
			(
				SELECT count(*) FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name IN ('upload_sessions', 'outbox')
			),
			(
				SELECT count(*) FROM videos
				WHERE publication_status = 'PUBLISHED' AND processing_status = 'READY'
			),
			(
				SELECT count(*) FROM videos
				WHERE title = 'Rascunho de exemplo' AND processing_status = 'WAITING_UPLOAD'
			)
	`).Scan(&processingColumn, &uploadTables, &readyPublished, &waitingDraft); err != nil {
		t.Fatal(err)
	}
	if processingColumn != 1 || uploadTables != 2 || readyPublished < 1 || waitingDraft != 1 || labWithoutOwner != 1 {
		t.Fatalf("processing=%d upload_tables=%d ready=%d draft=%d lab=%d", processingColumn, uploadTables, readyPublished, waitingDraft, labWithoutOwner)
	}
	var generationDefault, assetTables, claimColumn, jobColumns, assetConstraint, jobConstraint, seedJobs int
	var generationDefaultText string
	if err := pool.QueryRow(ctx, `
		SELECT
			(
				SELECT count(*) FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'videos' AND column_name = 'generation'
			),
			(
				SELECT column_default FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'videos' AND column_name = 'generation'
			),
			(
				SELECT count(*) FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name IN ('video_assets', 'processing_jobs')
			),
			(
				SELECT count(*) FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'outbox' AND column_name = 'claim_until'
			),
			(
				SELECT count(*) FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'processing_jobs'
					AND column_name IN ('lease_until', 'fencing_token')
			),
			(
				SELECT count(*) FROM pg_constraint
				WHERE conname = 'video_assets_generation_key'
			),
			(
				SELECT count(*) FROM pg_constraint
				WHERE conname = 'processing_jobs_generation_key'
			),
			(SELECT count(*) FROM processing_jobs)
	`).Scan(&generationDefault, &generationDefaultText, &assetTables, &claimColumn, &jobColumns, &assetConstraint, &jobConstraint, &seedJobs); err != nil {
		t.Fatal(err)
	}
	if generationDefault != 1 || !strings.Contains(generationDefaultText, "1") || assetTables != 2 || claimColumn != 1 || jobColumns != 2 || assetConstraint != 1 || jobConstraint != 1 || seedJobs != 0 {
		t.Fatalf("generation=%d default=%s tables=%d claim=%d job_cols=%d assets_key=%d jobs_key=%d seed_jobs=%d", generationDefault, generationDefaultText, assetTables, claimColumn, jobColumns, assetConstraint, jobConstraint, seedJobs)
	}
}
