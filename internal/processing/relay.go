package processing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Publisher confirms one durable event. A refusal leaves the outbox row unpublished.
type Publisher interface {
	Publish(ctx context.Context, eventType string, payload []byte) error
}

// Relay claims unpublished outbox rows and marks them only after the broker confirms.
type Relay struct {
	pool      *pgxpool.Pool
	publisher Publisher
	claimFor  time.Duration
}

// NewRelay binds the relay to a database and a confirming publisher.
func NewRelay(pool *pgxpool.Pool, publisher Publisher, claimFor time.Duration) (*Relay, error) {
	if pool == nil || publisher == nil {
		return nil, errors.New("banco e publicador são obrigatórios")
	}
	if claimFor <= 0 {
		claimFor = 30 * time.Second
	}
	return &Relay{pool: pool, publisher: publisher, claimFor: claimFor}, nil
}

type claimedEvent struct {
	id      string
	typ     string
	payload []byte
}

// PublishOnce claims a batch, publishes it, and marks only the confirmed rows.
func (r *Relay) PublishOnce(ctx context.Context) (int, error) {
	claimed, err := r.claim(ctx)
	if err != nil {
		return 0, err
	}
	published := 0
	var first error
	for _, event := range claimed {
		if err := r.publisher.Publish(ctx, event.typ, event.payload); err != nil {
			if first == nil {
				first = err
			}
			continue
		}
		if err := r.mark(ctx, event.id); err != nil {
			if first == nil {
				first = err
			}
			continue
		}
		published++
	}
	return published, first
}

func (r *Relay) claim(ctx context.Context) ([]claimedEvent, error) {
	rows, err := r.pool.Query(ctx, `
		WITH picked AS (
			SELECT id
			FROM outbox
			WHERE published_at IS NULL
				AND (claim_until IS NULL OR claim_until <= now())
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 8
		)
		UPDATE outbox AS pending
		SET claim_until = now() + make_interval(secs => $1)
		FROM picked
		WHERE pending.id = picked.id
		RETURNING pending.id::text, pending.type, pending.payload
	`, int(r.claimFor.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("reivindicar outbox: %w", err)
	}
	defer rows.Close()
	var claimed []claimedEvent
	for rows.Next() {
		var event claimedEvent
		if err := rows.Scan(&event.id, &event.typ, &event.payload); err != nil {
			return nil, err
		}
		claimed = append(claimed, event)
	}
	return claimed, rows.Err()
}

func (r *Relay) mark(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox
		SET published_at = now(), claim_until = NULL
		WHERE id = $1 AND published_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("marcar outbox: %w", err)
	}
	return nil
}
