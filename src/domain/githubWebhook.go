package domain

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/treenq/treenq/pkg/vel"
)

type GithubWebhookRequest struct {
	Installation Installation `json:"installation"`
	Sender       Sender       `json:"sender"`

	// installation only fields
	Action              string                `json:"action"`
	Repositories        []InstalledRepository `json:"repositories"`
	RepositoriesAdded   []InstalledRepository `json:"repositories_added"`
	RepositoriesRemoved []InstalledRepository `json:"repositories_removed"`

	// commits only
	Ref        string     `json:"ref"`
	Repository Repository `json:"repository"`
}

func (g GithubWebhookRequest) ReposToProcess() []InstalledRepository {
	// app install
	if g.Action == "created" {
		return g.Repositories
	}
	// repo added
	if g.Action == "added" {
		return g.RepositoriesAdded
	}
	// branch
	if g.Action == "" {
		if g.Ref != "refs/heads/master" && g.Ref != "refs/heads/main" {
			return nil
		}
		return []InstalledRepository{
			{
				ID:       g.Repository.ID,
				FullName: g.Repository.FullName,
				Private:  g.Repository.Private,
			},
		}
	}

	return nil
}

type Sender struct {
	Login string `json:"login"`
}

type Installation struct {
	ID int `json:"id"`
}

type Repository struct {
	ID       int    `json:"id"`
	CloneUrl string `json:"clone_url"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}

type InstalledRepository struct {
	ID       int    `json:"id"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
}

func (r InstalledRepository) CloneUrl() string {
	return fmt.Sprintf("https://github.com/%s.git", r.FullName)
}

type BuildArtifactRequest struct {
	Name       string
	Path       string
	Dockerfile string
}

type Image struct {
	// Registry is a registry name in the cloud provider
	Registry string
	// Repository is a facto name of the image
	Repository string
	// Tag is a version of the image
	Tag string
}

func (i Image) Image() string {
	return fmt.Sprintf("%s:%s", i.Repository, i.Tag)
}

func (i Image) FullPath() string {
	return fmt.Sprintf("%s/%s:%s", i.Registry, i.Repository, i.Tag)
}

type GithubWebhookResponse struct {
}

type Resource struct {
	Key     string
	Kind    ResourceKind
	Payload []byte
}

type ResourceKind string

const (
	ResourceKindService ResourceKind = "service"
)

type RepoConnection struct {
	Username   string
	Connect    []InstalledRepository
	Disconnect []InstalledRepository
}

type ReposConnector interface {
	ConnectRepos(ctx context.Context, repo RepoConnection) error
}

func (h *Handler) GithubWebhook(ctx context.Context, req GithubWebhookRequest) (GithubWebhookResponse, *vel.Error) {
	// clone, extract config
	// build and push
	// define and apply
	for _, repo := range req.ReposToProcess() {
		token := ""
		if repo.Private {
			var err error
			token, err = h.githubClient.IssueAccessToken(req.Installation.ID)
			if err != nil {
				return GithubWebhookResponse{}, &vel.Error{
					Code:    "UNKNOWN",
					Message: err.Error(),
				}
			}
		}

		repoDir, err := h.git.Clone(repo.CloneUrl(), req.Installation.ID, repo.ID, token)
		if err != nil {
			return GithubWebhookResponse{}, &vel.Error{
				Code:    "UNKNOWN",
				Message: err.Error(),
			}
		}
		defer os.RemoveAll(repoDir)

		extractorID, err := h.extractor.Open()
		if err != nil {
			return GithubWebhookResponse{}, &vel.Error{
				Code:    "UNKNOWN",
				Message: err.Error(),
			}
		}
		defer h.extractor.Close(extractorID)

		appDef, err := h.extractor.ExtractConfig(extractorID, repoDir)
		if err != nil {
			return GithubWebhookResponse{}, &vel.Error{
				Code:    "UNKNOWN",
				Message: err.Error(),
			}
		}

		dockerFilePath := filepath.Join(repoDir, appDef.Service.DockerfilePath)
		image, err := h.docker.Build(ctx, BuildArtifactRequest{
			Name:       appDef.Service.Name,
			Path:       repoDir,
			Dockerfile: dockerFilePath,
		})
		if err != nil {
			return GithubWebhookResponse{}, &vel.Error{
				Code:    "UNKNOWN",
				Message: err.Error(),
			}
		}
		_ = image

		//if err := h.db.SaveSpace(ctx, appDef.Service.Name, appDef.Region); err != nil {
		//	return GithubWebhookResponse{}, &vel.Error{
		//		Code:    "UNKNOWN",
		//		Message: err.Error(),
		//	}
		//}
		//
		//err = h.provider.CreateAppResource(ctx, image, appDef)
		//if err != nil {
		//	return GithubWebhookResponse{}, &vel.Error{
		//		Code:    "UNKNOWN",
		//		Message: err.Error(),
		//	}
		//}
		//
		//payload, err := json.Marshal(appDef)
		//if err != nil {
		//	return GithubWebhookResponse{}, &vel.Error{
		//		Code:    "UNKNOWN",
		//		Message: err.Error(),
		//	}
		//}
		//if err := h.db.SaveResource(ctx, Resource{
		//	Key:     appDef.Service.Key,
		//	Kind:    ResourceKindService,
		//	Payload: payload,
		//}); err != nil {
		//	return GithubWebhookResponse{}, &vel.Error{
		//		Code:    "UNKNOWN",
		//		Message: err.Error(),
		//	}
		//}
	}

	return GithubWebhookResponse{}, nil
}
