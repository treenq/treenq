package handlers

import (
	"context"
	"os"
	"path/filepath"

	"github.com/treenq/treenq/pkg/artifacts"
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

type GithubWebhookResponse struct {
}

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
	imageRepo, err := h.docker.Build(ctx, artifacts.Args{
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

	return GithubWebhookResponse{}, nil
}
