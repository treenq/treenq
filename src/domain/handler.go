package domain

import (
	"context"

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
}

type Database interface {
	SaveSpace(ctx context.Context, name string, region string) error
	SaveResource(ctx context.Context, resource Resource) error
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
	Build(ctx context.Context, args BuildArtifactRequest) (Image, error)
}

type Kube interface {
	DefineApp(ctx context.Context, id string, app tqsdk.Space, image Image) string
}

func NewHandler(
	db Database,
	githubClient GithubCleint,
	git Git,
	extractor Extractor,
	docker DockerArtifactory,
	authProfiler *auth.Context,
	kube Kube,
) *Handler {
	return &Handler{
		db:           db,
		githubClient: githubClient,
		git:          git,
		extractor:    extractor,
		docker:       docker,
		authProfiler: authProfiler,
		kube:         kube,
	}
}
