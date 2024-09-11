package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/digitalocean/godo"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/treenq/treenq/pkg/jwt"
	"github.com/treenq/treenq/pkg/vel"
	"github.com/treenq/treenq/pkg/vel/auth"
	"github.com/treenq/treenq/pkg/vel/log"
	"github.com/treenq/treenq/src/domain"
	"github.com/treenq/treenq/src/repo"
	"github.com/treenq/treenq/src/repo/artifacts"
	"github.com/treenq/treenq/src/repo/extract"
)

func New(conf Config) (http.Handler, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	l := log.NewLogger(os.Stdout, slog.LevelDebug)
	db, err := sqlx.Connect("pgx", conf.DbDsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	migrationsDir := filepath.Join(filepath.Join("file:///", wd), conf.MigrationsDir)
	m, err := migrate.New(
		migrationsDir,
		conf.DbDsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	store, err := repo.NewStore(db)
	if err != nil {
		return nil, err
	}

	doClient := godo.NewFromToken(conf.DoToken)
	provider := repo.NewProvider(doClient)
	jwtIssuer := jwt.NewJwtIssuer(conf.JwtKey, conf.JwtSecret, conf.JwtTtl)

	githubClient := repo.NewGithubClient(jwtIssuer, http.DefaultClient)
	gitClient := repo.NewGit(wd)
	docker := artifacts.NewDockerArtifactory("tq-staging")
	extractor := extract.NewExtractor(filepath.Join(wd, "builder"), "cmd/server")

	ctx := context.Background()

	authConf := auth.Config{
		ID:       conf.AuthID,
		Secret:   conf.AuthSecret,
		KeyID:    conf.AuthKeyID,
		Endpoint: conf.AuthEndpoint,
	}
	authMiddleware, authProfiler, err := auth.NewAuthMiddleware(ctx, authConf)
	if err != nil {
		return nil, err
	}

	handlers := domain.NewHandler(store, githubClient, gitClient, extractor, docker, provider, authProfiler)

	return NewRouter(handlers, log.NewLoggingMiddleware(l), authMiddleware).Mux(), nil
}

func NewRouter(handlers *domain.Handler, middlewares ...vel.Middleware) *vel.Router {
	router := vel.NewRouter()
	for i := range middlewares {
		router.Use(middlewares[i])
	}

	vel.Register(router, "deploy", handlers.Deploy)
	vel.Register(router, "githubWebhook", handlers.GithubWebhook)
	vel.Register(router, "info", handlers.Info)
	vel.Register(router, "getProfile", handlers.GetProfile)

	return router
}
