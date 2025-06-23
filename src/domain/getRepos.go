package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type GetReposResponse struct {
	Installation bool               `json:"installation"`
	Repos        []GithubRepository `json:"repos"`
}

func (h *Handler) GetRepos(ctx context.Context, _ struct{}) (GetReposResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return GetReposResponse{}, rpcErr
	}

	repos, err := h.db.GetGithubRepos(ctx, profile.UserInfo.ID)
	if err != nil {
		return GetReposResponse{}, &vel.Error{
			Code: "FAILED_GET_GITHUB_REPOS",
			Err:  err,
		}
	}
	return GetReposResponse{Repos: repos, Installation: len(repos) != 0}, nil
}
