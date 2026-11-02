package catalog_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"example.com/gotube/internal/catalog"
	"example.com/gotube/internal/platform/postgres"
)

func TestPublicListOmitsHiddenVideosAndPagesTies(t *testing.T) {
	ctx := context.Background()
	pool := postgres.OpenCatalogTestDB(t)
	root := moduleRoot(t)
	if err := postgres.ApplyMigrations(ctx, pool, filepath.Join(root, "db", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO channels (id, handle, name)
		VALUES ('00000000-0000-4000-8000-0000000000aa', 'canal', 'Canal Público');
		INSERT INTO videos (id, channel_id, title, duration_seconds, thumbnail_url, visibility, publication_status, published_at)
		VALUES
			('00000000-0000-4000-8000-0000000000b2', '00000000-0000-4000-8000-0000000000aa', 'Empate maior', 10, NULL, 'PUBLIC', 'PUBLISHED', '2026-03-01T12:00:00Z'),
			('00000000-0000-4000-8000-0000000000b1', '00000000-0000-4000-8000-0000000000aa', 'Empate menor', 11, 'https://example.com/thumb.jpg', 'PUBLIC', 'PUBLISHED', '2026-03-01T12:00:00Z'),
			('00000000-0000-4000-8000-0000000000c1', '00000000-0000-4000-8000-0000000000aa', 'Rascunho', 1, NULL, 'PUBLIC', 'DRAFT', NULL),
			('00000000-0000-4000-8000-0000000000c2', '00000000-0000-4000-8000-0000000000aa', 'Privado', 1, NULL, 'PRIVATE', 'PUBLISHED', '2026-03-02T12:00:00Z'),
			('00000000-0000-4000-8000-0000000000c3', '00000000-0000-4000-8000-0000000000aa', 'Não listado', 1, NULL, 'UNLISTED', 'PUBLISHED', '2026-03-02T12:00:00Z'),
			('00000000-0000-4000-8000-0000000000c4', '00000000-0000-4000-8000-0000000000aa', 'Bloqueado', 1, NULL, 'PUBLIC', 'BLOCKED', NULL);
	`); err != nil {
		t.Fatal(err)
	}

	store := catalog.NewStore(pool)
	first, err := store.ListPublic(ctx, 1, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Items) != 1 || first.Items[0].ID != "00000000-0000-4000-8000-0000000000b2" {
		t.Fatalf("first page = %#v", first.Items)
	}
	if first.NextCursor == nil {
		t.Fatal("expected a cursor after the first tied video")
	}
	if first.Items[0].ThumbnailURL != nil {
		t.Fatal("first video has no thumbnail")
	}
	if first.Items[0].Channel.Name != "Canal Público" || first.Items[0].Channel.Handle != "canal" {
		t.Fatalf("channel = %#v", first.Items[0].Channel)
	}

	second, err := store.ListPublic(ctx, 1, mustDecode(t, *first.NextCursor), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Items) != 1 || second.Items[0].ID != "00000000-0000-4000-8000-0000000000b1" {
		t.Fatalf("second page = %#v", second.Items)
	}
	if second.Items[0].ThumbnailURL == nil || *second.Items[0].ThumbnailURL != "https://example.com/thumb.jpg" {
		t.Fatalf("thumbnail = %#v", second.Items[0].ThumbnailURL)
	}
	if second.NextCursor != nil {
		t.Fatalf("cursor repeated a page: %s", *second.NextCursor)
	}
}

func mustDecode(t *testing.T, raw string) *catalog.Position {
	t.Helper()
	position, err := catalog.DecodeCursor(raw)
	if err != nil {
		t.Fatal(err)
	}
	return &position
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	root, err := postgres.ModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCursorRoundTripKeepsTieOrder(t *testing.T) {
	published := time.Date(2026, 3, 1, 12, 0, 0, 123456000, time.UTC)
	raw := catalog.EncodeCursor(published, "00000000-0000-4000-8000-0000000000b2")
	decoded, err := catalog.DecodeCursor(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !decoded.PublishedAt.Equal(published) || decoded.ID != "00000000-0000-4000-8000-0000000000b2" {
		t.Fatalf("decoded = %#v", decoded)
	}
	if _, err := catalog.DecodeCursor("não-é-cursor"); err == nil {
		t.Fatal("malformed cursor was accepted")
	}
}
