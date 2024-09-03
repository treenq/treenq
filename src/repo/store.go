package repo

import (
	"context"

	"github.com/jmoiron/sqlx"

	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/treenq/treenq/src/domain"
)

type Store struct {
	db *sqlx.DB
	sq sq.StatementBuilderType
}

func NewStore(db *sqlx.DB) (*Store, error) {
	sq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &Store{db: db, sq: sq}, nil
}

func (s *Store) SaveSpace(ctx context.Context, name string, region string) error {
	query, args, err := s.sq.Insert("spaces").Columns("name", "region").Values(name, region).ToSql()
	if err != nil {
		return fmt.Errorf("Store.SaveSpace: failed to build query: %w", err)
	}
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("Store.SaveSpace: failed to insert: %w", err)
	}

	return nil
}

func (s *Store) SaveResource(ctx context.Context, resource domain.Resource) error {
	query, args, err := s.sq.Insert("resources").Columns("key", "kind", "payload").Values(resource.Key, resource.Kind, resource.Payload).ToSql()
	if err != nil {
		return fmt.Errorf("Store.SaveResource: failed to build query: %w", err)
	}
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("Store.SaveResource: failed to insert: %w", err)
	}

	return nil
}
