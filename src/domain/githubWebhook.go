package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type GithubWebhookRequest struct {
	Ref          string `json:"ref"`
	Installation struct {
		ID int `json:"id"`
	} `json:"installation"`
	Repository struct {
		CloneUrl string `json:"clone_url"`
	} `json:"repository"`
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

func (h *Handler) GithubWebhook(ctx context.Context, req GithubWebhookRequest) (GithubWebhookResponse, *Error) {
	if req.Ref != "refs/heads/master" && req.Ref != "refs/heads/main" {
		return GithubWebhookResponse{}, nil
	}

	token, err := h.githubClient.IssueAccessToken(req.Installation.ID)
	if err != nil {
		return GithubWebhookResponse{}, &Error{
			Code:    "UNKNOWN",
			Message: err.Error(),
		}
	}

	repoDir, err := h.git.Clone(req.Repository.CloneUrl, token)
	if err != nil {
		return GithubWebhookResponse{}, &Error{
			Code:    "UNKNOWN",
			Message: err.Error(),
		}
	}
	defer os.RemoveAll(repoDir)

	extractorID, err := h.extractor.Open()
	if err != nil {
		return GithubWebhookResponse{}, &Error{
			Code:    "UNKNOWN",
			Message: err.Error(),
		}
	}
	defer h.extractor.Close(extractorID)

	appDef, err := h.extractor.ExtractConfig(extractorID, repoDir)
	if err != nil {
		return GithubWebhookResponse{}, &Error{
			Code:    "UNKNOWN",
			Message: err.Error(),
		}
	}

	dockerFilePath := filepath.Join(repoDir, appDef.Service.DockerfilePath)
	imageRepo, err := h.docker.Build(ctx, BuildArtifactRequest{
		Name:       appDef.Name,
		Path:       repoDir,
		Dockerfile: dockerFilePath,
	})
	if err != nil {
		return GithubWebhookResponse{}, &Error{
			Code:    "UNKNOWN",
			Message: err.Error(),
		}
	}

	err = h.provider.CreateAppResource(ctx, imageRepo, appDef)
	if err != nil {
		return GithubWebhookResponse{}, &Error{
			Code:    "UNKNOWN",
			Message: err.Error(),
		}
	}

	payload, err := json.Marshal(appDef)
	if err != nil {
		return GithubWebhookResponse{}, &Error{
			Code:    "UNKNOWN",
			Message: err.Error(),
		}
	}
	if err := h.db.SaveResource(ctx, Resource{
		Key:     appDef.Service.Key,
		Kind:    ResourceKindService,
		Payload: payload,
	}); err != nil {
		return GithubWebhookResponse{}, &Error{
			Code:    "UNKNOWN",
			Message: err.Error(),
		}
	}

	return GithubWebhookResponse{}, nil
}
