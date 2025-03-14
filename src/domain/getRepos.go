package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type GetReposRequest struct{}

type GetReposResponse struct {
	Repos []Repository `json:"repos"`
}

func (h *Handler) GetRepos(ctx context.Context, req GetReposRequest) (GetReposResponse, *vel.Error) {
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
	return GetReposResponse{Repos: repos}, nil
}
