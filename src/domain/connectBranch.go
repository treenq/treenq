package domain

import (
	"context"
	"errors"

	"github.com/treenq/treenq/pkg/vel"
)

type ConnectBranchRequest struct {
	RepoID string `json:"repoID"`
	Branch string `json:"branch"`
}

type ConnectBranchResponse struct {
	Repo Repository `json:"repo"`
}

func (h *Handler) ConnectBranch(ctx context.Context, req ConnectBranchRequest) (ConnectBranchResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return ConnectBranchResponse{}, rpcErr
	}

	repo, err := h.db.ConnectRepo(ctx, profile.UserInfo.ID, req.RepoID, req.Branch)
	if err != nil {
		if errors.Is(err, ErrRepoNotFound) {
			return ConnectBranchResponse{}, &vel.Error{
				Code: "REPO_NOT_FOUND",
			}
		}

		return ConnectBranchResponse{}, &vel.Error{
			Err:     err,
			Message: "failed to connect repo",
		}
	}
	return ConnectBranchResponse{Repo: repo}, nil
}
