package domain

import (
	"context"
	"slices"

	"github.com/treenq/treenq/pkg/vel"
)

type ConnectRequest struct {
	ID int `json:"id"`
}

type ConnectResponse struct {
}

func (h *Handler) ConnectRepository(ctx context.Context, req ConnectRequest) (ConnectResponse, *vel.Error) {
	email := h.authProfiler.GetProfile(ctx).Email
	userRepos, ok := repos[email]
	if !ok {
		return ConnectResponse{}, &vel.Error{
			Code:    "REPO_NOT_FOUND",
			Message: "Repository not found",
		}
	}
	idx := slices.IndexFunc(userRepos, func(repo GithubRepository) bool {
		return repo.ID == req.ID
	})
	if idx == -1 {
		return ConnectResponse{}, &vel.Error{
			Code:    "REPO_NOT_FOUND",
			Message: "Repository not found",
		}
	}

	h.db.SaveConnectedRepository(ctx, email, userRepos[idx])
	return ConnectResponse{}, nil
}
