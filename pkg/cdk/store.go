package cdk

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

type PgStore struct {
	db *sql.DB
	sq sq.StatementBuilderType
}

func NewPgStore(db *sql.DB) (*PgStore, error) {
	return &PgStore{
		db: db,
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func (s *PgStore) OpenResourceRecord(ctx context.Context, res tqsdk.Resource) (string, error) {
	id := uuid.NewString()
	query, args, err := s.sq.Insert("resources").Columns("id", "spaceKey", "key", "kind", "payload", "status", "createdAt").
		Values(id, res.SpaceKey, res.Key, res.Kind, res.Payload, "open", time.Now().UTC()).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("cdk.OpenResourceRecord: failed to build query: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return "", fmt.Errorf("PgStore.OpenResourceRecord: failed to insert: %w", err)
	}

	return id, nil
}

func (s *PgStore) GetOpenResources(ctx context.Context, spaceID string, key string) ([]SavedResource, error) {
	query, args, err := s.sq.Select("id", "space_id", "key", "kind", "payload").
		From("resources").
		Where(sq.Eq{"space_id": spaceID, "key": key, "status": "open"}).
		OrderBy("createdAt DESC").ToSql()

	if err != nil {
		return nil, fmt.Errorf("PgStore.getOpenResource: failed to build query: %w", err)
	}
	var resources []SavedResource
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("PgStore.getOpenResource: failed to select: %w", err)
	}

	for rows.Next() {
		var resource SavedResource
		err = rows.Scan(&resource.ID, &resource.SpaceKey, &resource.Key, &resource.Kind, &resource.Payload)
		if err != nil {
			return nil, fmt.Errorf("PgStore.getOpenResource: failed to scan: %w", err)
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func (s *PgStore) MarkResourceAsDone(ctx context.Context, id string) error {
	return s.markResource(ctx, id, "done")
}

func (s *PgStore) MarkResourceAsReverted(ctx context.Context, id string) error {
	return s.markResource(ctx, id, "reverted")
}

func (s *PgStore) markResource(ctx context.Context, id string, status string) error {
	query, args, err := s.sq.Update("resources").
		Set("status", status).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("PgStore.markResourceAs-%s: failed to build query: %w", status, err)
	}
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("PgStore.markResourceAs-%s: failed to update: %w", status, err)
	}
	return nil
}
