package domain

import (
	"context"
	"errors"

	"github.com/treenq/treenq/pkg/vel"
)

type ConnectBranchRequest struct {
	RepoID string
	Branch string
}

type ConnectBranchResponse struct {
	Repo InstalledRepository `json:"repo"`
}

func (h *Handler) ConnectBranch(ctx context.Context, req ConnectBranchRequest) (ConnectBranchResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return ConnectBranchResponse{}, rpcErr
	}

	repo, err := h.db.ConnectRepoBranch(ctx, profile.UserInfo.ID, req.RepoID, req.Branch)
	if err != nil {
		if errors.Is(err, ErrRepoNotFound) {
			return ConnectBranchResponse{}, &vel.Error{
				Code: "REPO_NOT_FOUND",
			}
		}
	}
	return ConnectBranchResponse{Repo: repo}, nil
}
