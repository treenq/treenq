package repo

import (
	"context"
	"database/sql"

	"fmt"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"

	"github.com/treenq/treenq/src/domain"
)

type Store struct {
	db *sql.DB
	sq sq.StatementBuilderType
}

func NewStore() (*Store, error) {
	db, err := sql.Open("postgres", "postgresql://postgres@localhost:5432/tq?sslmode=disable")
	if err != nil {
		return nil, nil
	}

	initDbDdl := `CREATE TABLE IF NOT EXISTS repos (
		id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
		url VARCHAR(255) NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS apps (
		id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		gitUrl VARCHAR(255) NOT NULL,
		gitBranch VARCHAR(255) NOT NULL,
		port VARCHAR(8) NOT NULL,
		buildCommand VARCHAR(1023) NOT NULL,
		runCommand VARCHAR(1023) NOT NULL,
		envs TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS spaces (
		id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		region VARCHAR(15) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS resources (
		id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
		spaceId VARCHAR(255) NOT NULL,
		key VARCHAR(255) NOT NULL,
		kind VARCHAR(55) NOT NULL,
		payload TEXT NOT NULL
	);
	`

	_, err = db.Exec(initDbDdl)
	if err != nil {
		return nil, fmt.Errorf("failed to init db tables: %w", err)
	}

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
