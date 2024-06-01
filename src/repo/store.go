package repo

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/treenq/treenq/src/handlers"
)

type Store struct {
	db *sql.DB
}

func NewStore() (*Store, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get store path: %w", err)
	}
	path = filepath.Join(path, "sqlite")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sql connection: % w", err)
	}

	initDbDdl := `CREATE TABLE IF NOT EXISTS repos (
		id TEXT PRIMARY KEY,
		url TEXT NOT NULL
	);`

	_, err = db.Exec(initDbDdl)
	if err != nil {
		return nil, fmt.Errorf("failed to init db tables: %w", err)
	}

	return &Store{db: db}, nil
}

// TODO: add (url, user) unique constraint
func (s *Store) CreateRepo(ctx context.Context, req handlers.ConnectRequest) (handlers.ConnectResponse, error) {
	id := uuid.NewString()
	insertSQL := `INSERT INTO repos (id, url) VALUES (?, ?)`
	_, err := s.db.Exec(insertSQL, id, req.Url)
	if err != nil {
		return handlers.ConnectResponse{}, fmt.Errorf("failed to insert repo: %w", err)
	}

	return handlers.ConnectResponse{
		ID:  id,
		Url: req.Url,
	}, nil
}
