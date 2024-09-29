package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/treenq/treenq/pkg/crypto"
	"github.com/treenq/treenq/pkg/jwt"
	"github.com/treenq/treenq/pkg/vel"
	"github.com/treenq/treenq/pkg/vel/auth"
	"github.com/treenq/treenq/pkg/vel/log"
	"github.com/treenq/treenq/src/domain"
	"github.com/treenq/treenq/src/repo"
	"github.com/treenq/treenq/src/repo/artifacts"
	"github.com/treenq/treenq/src/repo/extract"
	"github.com/treenq/treenq/src/services/cdk"
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

	jwtIssuer := jwt.NewIssuer(conf.JwtKey, []byte(conf.JwtSecret), conf.JwtTtl)

	githubClient := repo.NewGithubClient(jwtIssuer, http.DefaultClient)
	gitDir := filepath.Join(wd, "gits")
	gitClient := repo.NewGit(gitDir)
	docker := artifacts.NewDockerArtifactory(conf.DockerRegistry)
	extractor := extract.NewExtractor(filepath.Join(wd, "builder"), conf.BuilderPackage)

	ctx := context.Background()

	authConf := auth.Config{
		ID:       conf.AuthID,
		Secret:   string(conf.AuthSecret),
		KeyID:    conf.AuthKeyID,
		Endpoint: conf.AuthEndpoint,
	}
	var authProfiler *auth.Context
	authMiddleware, authProfiler, err := auth.NewAuthMiddleware(ctx, authConf, l)
	if err != nil {
		return nil, err
	}

	githubAuthMiddleware := vel.NoopMiddleware
	if conf.GithubWebhookSecretEnable {
		sha256Verifier := crypto.NewSha256SignatureVerifier(conf.GithubWebhookSecret, "sha256=")
		githubAuthMiddleware = crypto.NewSha256SignatureVerifierMiddleware(sha256Verifier, l)
	}

	kube := cdk.NewKube()

	handlers := domain.NewHandler(store, githubClient, gitClient, extractor, docker, authProfiler, kube)
	return NewRouter(handlers, authMiddleware, githubAuthMiddleware, log.NewLoggingMiddleware(l)).Mux(), nil
}

func NewRouter(handlers *domain.Handler, auth, githubAuth vel.Middleware, middlewares ...vel.Middleware) *vel.Router {
	router := vel.NewRouter()
	for i := range middlewares {
		router.Use(middlewares[i])
	}

	// insecure handlers, mus be covered with specific github secret middleware
	vel.Register(router, "githubWebhook", handlers.GithubWebhook, githubAuth)

	// regular authentication handlers
	vel.Register(router, "info", handlers.Info, auth)
	vel.Register(router, "getProfile", handlers.GetProfile, auth)

	return router
}
