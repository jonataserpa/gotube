package catalog

import "time"

// Item is the public video DTO. It has no account or processing fields.
type Item struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	DurationSeconds int     `json:"duration_seconds"`
	ThumbnailURL    *string `json:"thumbnail_url"`
	Channel         Channel `json:"channel"`
}

// Channel is the public name and handle shown on a catalog card.
type Channel struct {
	Name   string `json:"name"`
	Handle string `json:"handle"`
}

// Page is one keyset page. NextCursor is nil when the page is the last.
type Page struct {
	Items      []Item  `json:"items"`
	NextCursor *string `json:"next_cursor"`
}

// Position is the last item of a page, decoded from an opaque cursor.
type Position struct {
	PublishedAt time.Time
	ID          string
}
