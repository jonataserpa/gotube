package account

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type categoryStore interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// ListCategories returns active categories in name order.
func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	if s.pool == nil {
		return nil, ErrUnavailable
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, slug, name
		FROM categories
		WHERE active
		ORDER BY name, id
	`)
	if err != nil {
		return nil, errors.Join(ErrUnavailable, err)
	}
	defer rows.Close()
	items := []Category{}
	for rows.Next() {
		var item Category
		if err := rows.Scan(&item.ID, &item.Slug, &item.Name); err != nil {
			return nil, errors.Join(ErrUnavailable, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Join(ErrUnavailable, err)
	}
	return items, nil
}

func assertActiveCategories(ctx context.Context, db categoryStore, ids []string) error {
	unique := uniqueIDs(ids)
	if len(unique) == 0 {
		return nil
	}
	var found int
	err := db.QueryRow(ctx, `
		SELECT count(*)
		FROM categories
		WHERE active AND id = ANY($1::uuid[])
	`, unique).Scan(&found)
	if err != nil {
		if isInvalidUUID(err) {
			return ErrValidation
		}
		return errors.Join(ErrUnavailable, err)
	}
	if found != len(unique) {
		return ErrValidation
	}
	return nil
}

func replaceCategories(ctx context.Context, db categoryStore, channelID string, ids []string) error {
	if _, err := db.Exec(ctx, `DELETE FROM channel_categories WHERE channel_id = $1`, channelID); err != nil {
		return errors.Join(ErrUnavailable, err)
	}
	for _, id := range uniqueIDs(ids) {
		if _, err := db.Exec(ctx, `
			INSERT INTO channel_categories (channel_id, category_id)
			VALUES ($1, $2)
		`, channelID, id); err != nil {
			if isInvalidUUID(err) {
				return ErrValidation
			}
			return errors.Join(ErrUnavailable, err)
		}
	}
	return nil
}

func loadCategories(ctx context.Context, db categoryStore, channelID string) ([]Category, error) {
	rows, err := db.Query(ctx, `
		SELECT c.id::text, c.slug, c.name
		FROM channel_categories AS link
		JOIN categories AS c ON c.id = link.category_id AND c.active
		WHERE link.channel_id = $1
		ORDER BY c.name, c.id
	`, channelID)
	if err != nil {
		return nil, errors.Join(ErrUnavailable, err)
	}
	defer rows.Close()
	items := []Category{}
	for rows.Next() {
		var item Category
		if err := rows.Scan(&item.ID, &item.Slug, &item.Name); err != nil {
			return nil, errors.Join(ErrUnavailable, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Join(ErrUnavailable, err)
	}
	return items, nil
}

func uniqueIDs(ids []string) []string {
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
