package api

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/treenq/treenq/pkg/auth"
	"github.com/treenq/treenq/pkg/crypto"
	"github.com/treenq/treenq/pkg/treenq"
	"github.com/dennypenta/vel"
	"github.com/treenq/treenq/src/domain"
	"github.com/treenq/treenq/src/repo"
	"github.com/treenq/treenq/src/repo/artifacts"
	"github.com/treenq/treenq/src/repo/extract"
	"github.com/treenq/treenq/src/repo/git"
	"github.com/treenq/treenq/src/repo/github"
	"github.com/treenq/treenq/src/resources"

	authService "github.com/treenq/treenq/src/services/auth"
	"github.com/treenq/treenq/src/services/cdk"
)

func New(conf Config) (http.Handler, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	l := treenq.NewLogger(os.Stdout, slog.LevelDebug)

	db, err := resources.OpenDB(conf.DbDsn, conf.MigrationsDir)
	if err != nil {
		return nil, err
	}
	store, err := repo.NewStore(db)
	if err != nil {
		return nil, err
	}

	githubJwtIssuer := auth.NewJwtIssuer(conf.GithubClientID, []byte(conf.GithubPrivateKey), nil, conf.JwtTtl)
	authJwtIssuer := auth.NewJwtIssuer("treenq-api", []byte(conf.AuthPrivateKey), []byte(conf.AuthPublicKey), conf.AuthTtl)
	githubClient := github.NewGithubClient(githubJwtIssuer, http.DefaultClient)
	gitDir := filepath.Join(wd, "gits")
	gitClient := git.NewGit(gitDir)
	docker, err := artifacts.NewDockerArtifactory(
		conf.BuildkitHost,
		conf.BuildkitTLSCA,
		conf.DockerRegistry,
		conf.RegistryTLSVerify,
		conf.RegistryCert,
		conf.RegistryUsername,
		conf.RegistryPassword,
	)
	if err != nil {
		return nil, err
	}
	extractor := extract.NewExtractor()

	authMiddleware := auth.NewJwtMiddleware(authJwtIssuer, l)
	githubAuthMiddleware := vel.NoopMiddleware
	if conf.GithubWebhookSecretEnable {
		sha256Verifier := crypto.NewSha256SignatureVerifier(conf.GithubWebhookSecret, "sha256=")
		githubAuthMiddleware = crypto.NewSha256SignatureVerifierMiddleware(sha256Verifier, l)
	}

	oauthProvider := authService.New(conf.GithubClientID, conf.GithubSecret, conf.GithubRedirectURL)
	kube := cdk.NewKube(conf.Host, conf.DockerRegistry, conf.RegistryUsername, conf.RegistryPassword)
	handlers := domain.NewHandler(
		store,
		githubClient,
		gitClient,
		extractor,
		docker,
		kube,
		string(conf.KubeConfig),
		oauthProvider,
		authJwtIssuer,
		conf.AuthRedirectUrl,
		conf.AuthTtl,
		l,
		conf.IsProd,
	)
	return resources.NewRouter(handlers, authMiddleware, githubAuthMiddleware, treenq.NewLoggingMiddleware(l), treenq.NewCorsMiddleware(conf.CorsAllowOrigin)).Mux(), nil
}
