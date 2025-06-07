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

	l      *slog.Logger
	isProd bool
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
	isProd bool,
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
		isProd:          isProd,
	}
}

type Database interface {
	// User domain
	////////////////////////
	GetOrCreateUser(ctx context.Context, user UserInfo) (UserInfo, error)

	// Deployment domain
	// ////////////////
	SaveDeployment(ctx context.Context, def AppDeployment) (AppDeployment, error)
	UpdateDeployment(ctx context.Context, def AppDeployment) error
	GetDeployment(ctx context.Context, deploymentID string) (AppDeployment, error)
	GetDeployments(ctx context.Context, repoID string) ([]AppDeployment, error)

	// Github repos domain
	// //////////////////////
	LinkGithub(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) (string, error)
	SaveGithubRepos(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) error
	RemoveGithubRepos(ctx context.Context, installationID int, repos []InstalledRepository) error
	GetGithubRepos(ctx context.Context, email string) ([]GithubRepository, error)
	GetInstallationID(ctx context.Context, userID string) (string, int, error)
	SaveInstallation(ctx context.Context, userID string, githubID int) (string, error)
	ConnectRepo(ctx context.Context, userID, repoID, branchName string) (GithubRepository, error)
	GetRepoByGithub(ctx context.Context, githubRepoID int) (GithubRepository, error)
	GetRepoByID(ctx context.Context, userID, repoID string) (GithubRepository, error)
	RepoIsConnected(ctx context.Context, repoID string) (bool, error)

	// Secrets
	// ////////////////////////
	SaveSecret(ctx context.Context, repoID, key, userDisplayName string) error
	GetRepositorySecretKeys(ctx context.Context, repoID, userDisplayName string) ([]string, error)
	RepositorySecretKeyExists(ctx context.Context, repoID, key, userDisplayName string) (bool, error)
	RemoveSecret(ctx context.Context, repoID, key, userDisplayName string) error
}

type GithubClient interface {
	IssueAccessToken(installationID int) (string, error)
	GetUserInstallation(ctx context.Context, displayName string) (int, error)
	ListRepositories(ctx context.Context, installationID int) ([]GithubRepository, error)
	GetBranches(ctx context.Context, installationID int, owner string, repoName string, fresh bool) ([]string, error)
}

type Git interface {
	Clone(repo Repository, accesstoken, branch, sha string) (GitRepo, error)
}

type Extractor interface {
	ExtractConfig(repoDir string) (tqsdk.Space, error)
}

type DockerArtifactory interface {
	Image(name, tag string) Image
	Build(ctx context.Context, args BuildArtifactRequest, progress *ProgressBuf) (Image, error)
	Inspect(ctx context.Context, deploy AppDeployment) (Image, error)
}

type Kube interface {
	DefineApp(ctx context.Context, id, nsName string, app tqsdk.Space, image Image, secretKeys []string) string
	Apply(ctx context.Context, rawConig, data string) error
	StoreSecret(ctx context.Context, rawConfig, nsName, repoID, key, value string) error
	GetSecret(ctx context.Context, rawConfig, nsName, repoID, key string) (string, error)
	RemoveSecret(ctx context.Context, rawConfig string, space, repoID, key string) error
}

type OauthProvider interface {
	AuthUrl(string) string
	ExchangeUser(ctx context.Context, code string) (UserInfo, error)
}

type JwtIssuer interface {
	GenerateJwtToken(claims map[string]any) (string, error)
}
