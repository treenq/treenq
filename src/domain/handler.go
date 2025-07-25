package domain

import (
	"context"
	"io"
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
	GetUserWorkspaces(ctx context.Context, userID string) ([]Workspace, error)
	GetDefaultWorkspace(ctx context.Context, userID string) (Workspace, error)
	GetWorkspaceByID(ctx context.Context, workspaceID string) (Workspace, error)
	GetWorkspaceByUserDisplayName(ctx context.Context, userDisplayName string) (Workspace, error)
	DeploymentBelongsToWorkspace(ctx context.Context, workspaceID, deploymentID string) (bool, error)

	// Deployment domain
	// ////////////////
	SaveDeployment(ctx context.Context, def AppDeployment) (AppDeployment, error)
	UpdateDeployment(ctx context.Context, def AppDeployment) error
	GetDeployment(ctx context.Context, workspaceID, deploymentID string) (AppDeployment, error)
	GetDeployments(ctx context.Context, workspaceID, repoID string) ([]AppDeployment, error)

	// Github repos domain
	// //////////////////////
	LinkGithub(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) (string, error)
	LinkAllGithubInstallations(ctx context.Context, profile UserInfo, installationsRepos map[int][]GithubRepository) error
	SaveGithubRepos(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) error
	RemoveGithubRepos(ctx context.Context, installationID int, repos []InstalledRepository) error
	GetGithubRepos(ctx context.Context, workspaceID string) ([]GithubRepository, bool, error)
	GetInstallationID(ctx context.Context, workspaceID, fullName string) (int, error)
	ConnectRepo(ctx context.Context, workspaceID, repoID, branchName string, space tqsdk.Space) (GithubRepository, error)
	GetRepoByGithub(ctx context.Context, githubRepoID int) (GithubRepository, error)
	GetRepoByID(ctx context.Context, workspaceID, repoID string) (GithubRepository, error)
	RepoIsConnected(ctx context.Context, repoID string) (bool, error)
	GetSpace(ctx context.Context, repoID string) (tqsdk.Space, error)
	SaveSpace(ctx context.Context, repoID string, space tqsdk.Space) error

	// Secrets
	// ////////////////////////
	SaveSecret(ctx context.Context, repoID, key, workspaceID string) error
	GetRepositorySecretKeys(ctx context.Context, repoID, workspaceID string) ([]string, error)
	RepositorySecretKeyExists(ctx context.Context, repoID, key, workspaceID string) (bool, error)
	RemoveSecret(ctx context.Context, repoID, key, workspaceID string) error

	// Installation cleanup
	// ////////////////////////
	RemoveInstallation(ctx context.Context, installationID int) error
}

type GithubClient interface {
	IssueAccessToken(installationID int) (string, error)
	GetUserAccessibleInstallations(ctx context.Context, userGithubToken string) ([]int, error)
	ListRepositories(ctx context.Context, installationID int) ([]GithubRepository, error)
	ListAllRepositoriesForInstallations(ctx context.Context, installationIDs []int) (map[int][]GithubRepository, error)
	ListAllRepositoriesForUser(ctx context.Context, userGithubToken string) (map[int][]GithubRepository, error)
	GetBranches(ctx context.Context, installationID int, repoName string, fresh bool) ([]string, error)
	GetRepoSpace(ctx context.Context, installationID int, repoFullName, ref string) (tqsdk.Space, error)
}

type Git interface {
	Clone(repo Repository, accesstoken, branch, sha, tag string, progress io.Writer) (GitRepo, error)
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
	DefineApp(ctx context.Context, id, nsName string, app tqsdk.Space, image Image, secretKeys []string) (string, error)
	Apply(ctx context.Context, rawConig, data string) error
	StoreSecret(ctx context.Context, rawConfig, nsName, repoID, key, value string) error
	GetSecret(ctx context.Context, rawConfig, nsName, repoID, key string) (string, error)
	RemoveSecret(ctx context.Context, rawConfig string, space, repoID, key string) error
	StreamLogs(ctx context.Context, rawConfig, repoID, spaceName string, logChan chan<- ProgressMessage) error
	RemoveNamespace(ctx context.Context, rawConfig, id, nsName string) error
	GetWorkloadStats(ctx context.Context, rawConfig, repoID, spaceName string) (WorkloadStats, error)
}

type OauthProvider interface {
	AuthUrl(string) string
	ExchangeUser(ctx context.Context, code string) (UserInfo, error)
	GetUserGithubToken(userDisplayName string) (string, error)
}

type JwtIssuer interface {
	GenerateJwtToken(claims map[string]any) (string, error)
}
