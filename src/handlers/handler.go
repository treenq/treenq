package handlers

import (
	"context"

	"github.com/treenq/treenq/pkg/artifacts"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

type Handler struct {
	db           Database
	githubClient GithubCleint
	git          Git
	extractor    Extractor
	docker       DockerArtifactory
	provider     Provider
}

type Database interface {
	CreateRepo(ctx context.Context, req ConnectRequest) (ConnectResponse, error)
}

type GithubCleint interface {
	IssueAccessToken(installationID int) (string, error)
}

type Git interface {
	Clone(url string, accesstoken string) (string, error)
}

type Extractor interface {
	Open() (string, error)
	ExtractConfig(id, repoDir string) (tqsdk.App, error)
	Close(string) error
}

type DockerArtifactory interface {
	Build(ctx context.Context, args artifacts.Args) (string, error)
}

type Provider interface {
	CreateAppResource(ctx context.Context, image string, app tqsdk.App) error
}

func NewHandler(
	db Database,
	githubClient GithubCleint,
	git Git,
	extractor Extractor,
	docker DockerArtifactory,
	provider Provider,
) *Handler {
	return &Handler{
		db:           db,
		githubClient: githubClient,
		git:          git,
		extractor:    extractor,
		docker:       docker,
		provider:     provider,
	}
}
