package resources

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func OpenDB(ctx context.Context, dbDsn, migrationsDirName string) (*sqlx.DB, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db, err := sqlx.ConnectContext(ctx, "pgx", dbDsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	migrationsPath := filepath.Join(filepath.Join("file:///", wd), migrationsDirName)
	fmt.Println("[DEBUG] create migration instance on path=", migrationsPath)
	m, err := migrate.New(
		migrationsPath,
		dbDsn,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}
