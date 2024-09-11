package domain

import (
	"context"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/pkg/vel/auth"
)

type Handler struct {
	db           Database
	githubClient GithubCleint
	git          Git
	extractor    Extractor
	docker       DockerArtifactory
	provider     Provider
	authProfiler *auth.Context
}

type Database interface {
	SaveSpace(ctx context.Context, name string, region string) error
	SaveResource(ctx context.Context, resource Resource) error
}

type GithubCleint interface {
	IssueAccessToken(installationID int) (string, error)
}

type Git interface {
	Clone(url string, accesstoken string) (string, error)
}

type Extractor interface {
	Open() (string, error)
	ExtractConfig(id, repoDir string) (tqsdk.Space, error)
	Close(string) error
}

type DockerArtifactory interface {
	Build(ctx context.Context, args BuildArtifactRequest) (Image, error)
}

type Provider interface {
	CreateAppResource(ctx context.Context, image Image, app tqsdk.Space) error
}

func NewHandler(
	db Database,
	githubClient GithubCleint,
	git Git,
	extractor Extractor,
	docker DockerArtifactory,
	provider Provider,
	authProfiler *auth.Context,
) *Handler {
	return &Handler{
		db:           db,
		githubClient: githubClient,
		git:          git,
		extractor:    extractor,
		docker:       docker,
		provider:     provider,
		authProfiler: authProfiler,
	}
}
