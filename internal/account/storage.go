package account

import (
	"context"
	"time"
)

// ObjectStore signs one PUT and reports whether the object is present.
type ObjectStore interface {
	PresignPut(ctx context.Context, key string, expiresAt time.Time) (string, error)
	Head(ctx context.Context, key string) (size int64, found bool, err error)
}

// UseStorage binds the object store used by upload sessions.
func (s *Service) UseStorage(store ObjectStore, maxBytes int64) {
	s.store = store
	s.maxBytes = maxBytes
}
