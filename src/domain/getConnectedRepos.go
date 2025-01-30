package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type GetConnectedReposRequest struct{}

type GetConnectedReposResponse struct {
	Repos []GithubRepository `json:"repos"`
}

func (h *Handler) GetConnectedRepos(ctx context.Context, req GetConnectedReposRequest) (GetConnectedReposResponse, *vel.Error) {
	profile, profileErr := h.GetProfile(ctx, struct{}{})
	if profileErr != nil {
		return GetConnectedReposResponse{}, profileErr
	}
	email := profile.UserInfo.Email
	userRepos, err := h.db.GetConnectedRepositories(ctx, email)
	if err != nil {
		return GetConnectedReposResponse{}, &vel.Error{
			Code:    "CONNECTED_REPOS_NOT_FOUND",
			Message: "Repository not found",
		}
	}

	return GetConnectedReposResponse{Repos: userRepos}, nil
}
