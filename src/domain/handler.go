package domain

import (
	"context"
	"log/slog"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/pkg/vel/auth"
)

type Handler struct {
	reposConnector ReposConnector

	db           Database
	githubClient GithubCleint
	git          Git
	extractor    Extractor
	docker       DockerArtifactory
	authProfiler *auth.Context
	kube         Kube

	kubeConfig string

	githubClientID      string
	githubSecret        string
	githubRedirectURI   string
	githubWebhookSecret string
	githubWebhookURL    string

	authService AuthService
	l           *slog.Logger
}

type ReposConnector interface {
	ConnectRepos(ctx context.Context, repo RepoConnection) error
}

type Database interface {
	SaveDeployment(ctx context.Context, def AppDefinition) error
	GetDeploymentHistory(ctx context.Context, appID string) ([]AppDefinition, error)

	GetConnectedRepositories(ctx context.Context, email string) ([]GithubRepository, error)
	SaveConnectedRepository(ctx context.Context, email string, repo GithubRepository) error
	RemoveConnectedRepository(ctx context.Context, email string, repoID int) error

	SaveAuthState(ctx context.Context, email, state string) error
	AuthStateExists(ctx context.Context, state string) (string, error)
	SaveTokenPair(ctx context.Context, email string, tokenPair string) error
	GetTokenPair(ctx context.Context, email string) (string, error)
	SaveGithubRepos(ctx context.Context, email string, repos []GithubRepository) error
	GetGithubRepos(ctx context.Context, email string) ([]GithubRepository, error)
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

type AuthService interface {
	Start(ctx context.Context, provider string) (authUrl string, err error)
	GetUser(ctx context.Context, intent, token string) (UserInfo, error)
	CreateUser(ctx context.Context, user UserInfo) error
	Login()
}

func NewHandler(
	db Database,
	githubClient GithubCleint,
	git Git,
	extractor Extractor,
	docker DockerArtifactory,
	authProfiler *auth.Context,
	kube Kube,
	kubeConfig string,

	githubClientID string,
	githubSecret string,
	githubRedirectURI string,
	githubWebhookSecret string,
	githubWebhookURL string,
	authService AuthService,
	l *slog.Logger,
) *Handler {
	return &Handler{
		db:           db,
		githubClient: githubClient,
		git:          git,
		extractor:    extractor,
		docker:       docker,
		authProfiler: authProfiler,
		kube:         kube,

		kubeConfig: kubeConfig,

		githubClientID:      githubClientID,
		githubSecret:        githubSecret,
		githubRedirectURI:   githubRedirectURI,
		githubWebhookSecret: githubWebhookSecret,
		githubWebhookURL:    githubWebhookURL,
		authService:         authService,
		l:                   l,
	}
}
