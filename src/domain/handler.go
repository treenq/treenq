package domain

import (
	"context"
	"log/slog"
	"time"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

type Handler struct {
	db           Database
	githubClient GithubClient
	git          Git
	extractor    Extractor
	docker       DockerArtifactory
	kube         Kube

	kubeConfig string

	oauthProvider   OauthProvider
	jwtIssuer       JwtIssuer
	authRedirectUrl string
	authTtl         time.Duration

	l *slog.Logger
}

func NewHandler(
	db Database,
	githubClient GithubClient,
	git Git,
	extractor Extractor,
	docker DockerArtifactory,
	kube Kube,
	kubeConfig string,

	oauthProvider OauthProvider,
	jwtIssuer JwtIssuer,
	authRedirectUrl string,
	authTtl time.Duration,

	l *slog.Logger,
) *Handler {
	return &Handler{
		db:           db,
		githubClient: githubClient,
		git:          git,
		extractor:    extractor,
		docker:       docker,
		kube:         kube,

		kubeConfig: kubeConfig,

		oauthProvider:   oauthProvider,
		jwtIssuer:       jwtIssuer,
		authRedirectUrl: authRedirectUrl,
		authTtl:         authTtl,
		l:               l,
	}
}

type Database interface {
	// User domain
	////////////////////////
	GetOrCreateUser(ctx context.Context, user UserInfo) (UserInfo, error)

	// Deployment domain
	// ////////////////
	SaveDeployment(ctx context.Context, def AppDeployment) (AppDeployment, error)
	GetDeploymentHistory(ctx context.Context, appID string) ([]AppDeployment, error)

	// Github repos domain
	// //////////////////////
	LinkGithub(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) error
	SaveGithubRepos(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) error
	RemoveGithubRepos(ctx context.Context, installationID int, repos []InstalledRepository) error
	GetGithubRepos(ctx context.Context, email string) ([]Repository, error)
	GetInstallationID(ctx context.Context, userID string) (string, error)
	ConnectRepo(ctx context.Context, userID, repoID, branchName string) (Repository, error)
	GetRepoByGithub(ctx context.Context, githubRepoID int) (Repository, error)
	GetRepoByID(ctx context.Context, userID, repoID string) (Repository, error)
	RepoIsConnected(ctx context.Context, repoID string) (bool, error)
}

type GithubClient interface {
	IssueAccessToken(installationID int) (string, error)
}

type Git interface {
	Clone(url string, installationID int, repoID string, accesstoken string) (GitRepo, error)
}

type Extractor interface {
	ExtractConfig(repoDir string) (tqsdk.Space, error)
}

type DockerArtifactory interface {
	Image(args BuildArtifactRequest) Image
	Build(ctx context.Context, args BuildArtifactRequest) (Image, error)
}

type Kube interface {
	DefineApp(ctx context.Context, id string, app tqsdk.Space, image Image) string
	Apply(ctx context.Context, rawConig, data string) error
}

type OauthProvider interface {
	AuthUrl(string) string
	ExchangeCode(ctx context.Context, code string) (string, error)
	FetchUser(ctx context.Context, token string) (UserInfo, error)
}

type JwtIssuer interface {
	GenerateJwtToken(claims map[string]any) (string, error)
}
