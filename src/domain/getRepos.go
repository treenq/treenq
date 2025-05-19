package domain

import (
	"context"
	"errors"

	"github.com/treenq/treenq/pkg/vel"
)

type GetReposResponse struct {
	Installation string       `json:"installationID"`
	Repos        []Repository `json:"repos"`
}

func (h *Handler) GetRepos(ctx context.Context, _ struct{}) (GetReposResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return GetReposResponse{}, rpcErr
	}

	installation, _, err := h.db.GetInstallationID(ctx, profile.UserInfo.ID)
	if err != nil {
		if errors.Is(err, ErrInstallationNotFound) {
			return GetReposResponse{}, &vel.Error{
				Code: "INSTALLATION_NOT_FOUND",
			}
		}
		return GetReposResponse{}, &vel.Error{
			Message: "failed to get installation",
			Err:     err,
		}
	}

	repos, err := h.db.GetGithubRepos(ctx, profile.UserInfo.ID)
	if err != nil {
		return GetReposResponse{}, &vel.Error{
			Code: "FAILED_GET_GITHUB_REPOS",
			Err:  err,
		}
	}
	return GetReposResponse{Repos: repos, Installation: installation}, nil
}
