package catalog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrUnavailable means the catalog database could not serve the list.
var ErrUnavailable = errors.New("catálogo indisponível")

const listSQL = `
SELECT v.id::text,
       v.title,
       v.duration_seconds,
       v.thumbnail_url,
       c.name,
       c.handle,
       v.published_at
FROM videos AS v
JOIN channels AS c ON c.id = v.channel_id
WHERE v.visibility = 'PUBLIC'
  AND v.publication_status = 'PUBLISHED'
  AND (
    $1::bool = false
    OR v.published_at < $2::timestamptz
    OR (v.published_at = $2::timestamptz AND v.id < $3::uuid)
  )
  AND ($5::bool = false OR v.channel_id = $6::uuid)
ORDER BY v.published_at DESC, v.id DESC
LIMIT $4
`

// Querier is the read surface the list needs. The pool implements it.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Store lists public published videos. It does not migrate.
type Store struct {
	db Querier
}

// NewStore binds the list to a query surface.
func NewStore(db Querier) *Store {
	return &Store{db: db}
}

type listedVideo struct {
	item        Item
	publishedAt time.Time
}

// ListPublic returns one keyset page of public published videos.
// A nil position starts at the newest item. limit is the page size.
func (s *Store) ListPublic(ctx context.Context, limit int, after *Position, channelID string) (Page, error) {
	if s == nil || s.db == nil {
		return Page{}, ErrUnavailable
	}
	rows, err := s.query(ctx, limit+1, after, channelID)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()

	collected, err := collectVideos(rows, limit+1)
	if err != nil {
		return Page{}, err
	}
	if err := rows.Err(); err != nil {
		return Page{}, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	return pageFrom(collected, limit), nil
}

func (s *Store) query(ctx context.Context, fetch int, after *Position, channelID string) (pgx.Rows, error) {
	hasCursor := after != nil
	var publishedAt any
	var id any
	if hasCursor {
		publishedAt = after.PublishedAt.UTC()
		id = after.ID
	}
	var channelArg any
	if channelID != "" {
		channelArg = channelID
	}
	rows, err := s.db.Query(ctx, listSQL, hasCursor, publishedAt, id, fetch, channelID != "", channelArg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	return rows, nil
}

func collectVideos(rows pgx.Rows, capacity int) ([]listedVideo, error) {
	collected := make([]listedVideo, 0, capacity)
	for rows.Next() {
		var entry listedVideo
		if err := rows.Scan(
			&entry.item.ID,
			&entry.item.Title,
			&entry.item.DurationSeconds,
			&entry.item.ThumbnailURL,
			&entry.item.Channel.Name,
			&entry.item.Channel.Handle,
			&entry.publishedAt,
		); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrUnavailable, err)
		}
		collected = append(collected, entry)
	}
	return collected, nil
}

func pageFrom(collected []listedVideo, limit int) Page {
	end := len(collected)
	more := end > limit
	if more {
		end = limit
	}
	page := Page{Items: make([]Item, 0, end)}
	for _, entry := range collected[:end] {
		page.Items = append(page.Items, entry.item)
	}
	if more {
		last := collected[end-1]
		cursor := EncodeCursor(last.publishedAt, last.item.ID)
		page.NextCursor = &cursor
	}
	return page
}
