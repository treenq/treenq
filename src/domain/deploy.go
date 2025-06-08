package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type DeployRequest struct {
	RepoID           string `json:"repoID"`
	FromDeploymentID string `json:"fromDeploymentID"`
}

func (h *Handler) Deploy(ctx context.Context, req DeployRequest) (GetDeploymentResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return GetDeploymentResponse{}, rpcErr
	}

	repo, err := h.db.GetRepoByID(ctx, profile.UserInfo.ID, req.RepoID)
	if err != nil {
		return GetDeploymentResponse{}, &vel.Error{
			Message: "failed to get installation id for a repo",
			Err:     err,
		}
	}

	appDeployment, apiErr := h.deployRepo(
		ctx,
		profile.UserInfo.DisplayName,
		repo,
		req.FromDeploymentID,
	)
	if apiErr != nil {
		return GetDeploymentResponse{}, apiErr
	}

	return GetDeploymentResponse{
		Deployment: appDeployment,
	}, nil
}
