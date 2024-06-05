package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/mattn/go-sqlite3"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/handlers"
)

type Store struct {
	db *sql.DB
	sq sq.StatementBuilderType
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
		id uuid DEFAULT gen_random_uuid(),
		url varchar(255) NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS apps (
		id uuid DEFAULT gen_random_uuid(),
		name varchar(255) NOT NULL,
		gitUrl varchar(255) NOT NULL,
		gitBranch varchar(255) NOT NULL,
		port varchar(8) NOT NULL,
		buildCommand varchar(1023) NOT NULL,
		runCommand varchar(1023) NOT NULL,
		envs TEXT NOT NULL
	)
	`

	_, err = db.Exec(initDbDdl)
	if err != nil {
		return nil, fmt.Errorf("failed to init db tables: %w", err)
	}

	sq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &Store{db: db, sq: sq}, nil
}

// TODO: add (url, user) unique constraint
func (s *Store) CreateRepo(ctx context.Context, req handlers.ConnectRequest) (handlers.ConnectResponse, error) {
	query, args, err := sq.Insert("repos").Columns("url").Values(req.Url).Suffix("RETURNING 'id'").ToSql()
	if err != nil {
		return handlers.ConnectResponse{}, fmt.Errorf("failed to build repo insert query: %w", err)
	}
	var id string
	err = s.db.QueryRow(query, args...).Scan(&id)
	if err != nil {
		return handlers.ConnectResponse{}, fmt.Errorf("failed to insert repo: %w", err)
	}

	return handlers.ConnectResponse{
		ID:  id,
		Url: req.Url,
	}, nil
}

var appsCols = []string{
	"name",
	"gitUrl",
	"gitBranch",
	"port",
	"buildCommand",
	"runCommand",
	"envs",
}

func (s *Store) SaveApp(app tqsdk.App) (handlers.App, error) {
	envs, err := json.Marshal(app.Envs)
	if err != nil {
		return handlers.App{}, fmt.Errorf("failed to marshal envs to json: %w", err)
	}
	query, args, err := sq.Insert("apps").Columns(appsCols...).Values(app.Name, app.Git.Url, app.Git.Branch, app.Port, app.BuildCommand, app.RunCommand, string(envs)).Suffix("RETURNING 'id'").ToSql()
	if err != nil {
		return handlers.App{}, fmt.Errorf("failed to build repo insert query: %w", err)
	}

	var id string
	if err := s.db.QueryRow(query, args...).Scan(&id); err != nil {
		return handlers.App{}, fmt.Errorf("failed to insert app: %w", err)
	}

	return handlers.App{
		ID:  id,
		App: app,
	}, nil
}
