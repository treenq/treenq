package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type DeployRequest struct {
	RepoID  string
}

type DeployResponse struct {
}

func (h *Handler) Deploy(ctx context.Context, req DeployRequest) (DeployResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return DeployResponse{}, rpcErr
	}

	repo, err := h.db.GetRepoByID(ctx, profile.UserInfo.ID, req.RepoID)
	if err != nil {
		return DeployResponse{}, &vel.Error{
			Message: "failed to get installation id for a repo",
			Err:     err,
		}
	}

	return DeployResponse{}, h.deployRepo(
		ctx,
		profile.UserInfo.DisplayName,
		repo,
	)
}
