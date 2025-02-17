package api

import (
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
	"github.com/treenq/treenq/pkg/vel"
	"github.com/treenq/treenq/pkg/vel/auth"
	"github.com/treenq/treenq/pkg/vel/log"
	"github.com/treenq/treenq/src/domain"
	"github.com/treenq/treenq/src/repo"
	"github.com/treenq/treenq/src/repo/artifacts"
	"github.com/treenq/treenq/src/repo/extract"

	authService "github.com/treenq/treenq/src/services/auth"
	"github.com/treenq/treenq/src/services/cdk"
)

func OpenDB(dbDsn, migrationsDirName string) (*sqlx.DB, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	db, err := sqlx.Connect("pgx", dbDsn)
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

func New(conf Config) (http.Handler, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	l := log.NewLogger(os.Stdout, slog.LevelDebug)

	db, err := OpenDB(conf.DbDsn, conf.MigrationsDir)
	if err != nil {
		return nil, err
	}
	store, err := repo.NewStore(db)
	if err != nil {
		return nil, err
	}

	githubJwtIssuer := auth.NewJwtIssuer(conf.GithubClientID, []byte(conf.GithubPrivateKey), nil, conf.JwtTtl)
	authJwtIssuer := auth.NewJwtIssuer("treenq-api", []byte(conf.AuthPrivateKey), []byte(conf.AuthPublicKey), conf.AuthTtl)
	githubClient := repo.NewGithubClient(githubJwtIssuer, http.DefaultClient)
	gitDir := filepath.Join(wd, "gits")
	gitClient := repo.NewGit(gitDir)
	docker := artifacts.NewDockerArtifactory(conf.DockerRegistry)
	extractor := extract.NewExtractor(filepath.Join(wd, "builder"), conf.BuilderPackage)

	authMiddleware := auth.NewJwtMiddleware(authJwtIssuer, l)
	githubAuthMiddleware := vel.NoopMiddleware
	if conf.GithubWebhookSecretEnable {
		sha256Verifier := crypto.NewSha256SignatureVerifier(conf.GithubWebhookSecret, "sha256=")
		githubAuthMiddleware = crypto.NewSha256SignatureVerifierMiddleware(sha256Verifier, l)
	}

	oauthProvider := authService.New(conf.GithubClientID, conf.GithubSecret, conf.GithubRedirectURL)
	kube := cdk.NewKube()
	handlers := domain.NewHandler(
		store,
		githubClient,
		gitClient,
		extractor,
		docker,
		kube,
		conf.KubeConfig,
		oauthProvider,
		authJwtIssuer,
		conf.GithubWebhookURL,
		l,
	)
	return NewRouter(handlers, authMiddleware, githubAuthMiddleware, log.NewLoggingMiddleware(l)).Mux(), nil
}

func NewRouter(handlers *domain.Handler, auth, githubAuth vel.Middleware, middlewares ...vel.Middleware) *vel.Router {
	router := vel.NewRouter()
	for i := range middlewares {
		router.Use(middlewares[i])
	}

	vel.RegisterHandlerFunc(router, "/auth", handlers.GithubAuthHandler)
	vel.RegisterHandlerFunc(router, "/authCallback", handlers.GithubCallbackHandler)

	vel.Register(router, "githubWebhook", handlers.GithubWebhook, githubAuth)

	// regular authentication handlers
	vel.Register(router, "info", handlers.Info, auth)
	vel.Register(router, "getProfile", handlers.GetProfile, auth)
	vel.Register(router, "getRepos", handlers.GetRepos, auth)
	vel.Register(router, "connectRepoBranch", handlers.ConnectBranch, auth)

	return router
}
