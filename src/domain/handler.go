package domain

import (
	"context"
	"log/slog"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

type Handler struct {
	db           Database
	githubClient GithubCleint
	git          Git
	extractor    Extractor
	docker       DockerArtifactory
	kube         Kube

	kubeConfig string

	oauthProvider    OauthProvider
	jwtIssuer        JwtIssuer
	githubWebhookURL string

	l *slog.Logger
}

func NewHandler(
	db Database,
	githubClient GithubCleint,
	git Git,
	extractor Extractor,
	docker DockerArtifactory,
	kube Kube,
	kubeConfig string,

	oauthProvider OauthProvider,
	jwtIssuer JwtIssuer,
	githubWebhookURL string,
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

		oauthProvider:    oauthProvider,
		jwtIssuer:        jwtIssuer,
		githubWebhookURL: githubWebhookURL,
		l:                l,
	}
}

type Database interface {
	// User domain
	////////////////////////
	GetOrCreateUser(ctx context.Context, user UserInfo) (UserInfo, error)

	// Deployment domain
	// ////////////////
	SaveDeployment(ctx context.Context, def AppDefinition) (AppDefinition, error)
	GetDeploymentHistory(ctx context.Context, appID string) ([]AppDefinition, error)

	// Github repos domain
	// //////////////////////
	LinkGithub(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) error
	SaveGithubRepos(ctx context.Context, installationID int, senderLogin string, repos []InstalledRepository) error
	RemoveGithubRepos(ctx context.Context, installationID int, repos []InstalledRepository) error
	GetGithubRepos(ctx context.Context, email string) ([]InstalledRepository, error)
	ConnectRepoBranch(ctx context.Context, repoID int, branch string) error
}

type GithubCleint interface {
	IssueAccessToken(installationID int) (string, error)
}

type Git interface {
	Clone(url string, installationID, repoID int, accesstoken string) (string, error)
}

type Extractor interface {
	Open() (string, error)
	ExtractConfig(id, repoDir string) (tqsdk.Space, error)
	Close(string) error
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
	GenerateJwtToken(claims map[string]interface{}) (string, error)
}
