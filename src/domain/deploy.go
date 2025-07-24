package domain

import (
	"context"
	"errors"

	"github.com/dennypenta/vel"
)

type DeployRequest struct {
	RepoID           string `json:"repoID"`
	FromDeploymentID string `json:"fromDeploymentID"`
	Branch           string `json:"branch"`
	Sha              string `json:"sha"`
	Tag              string `json:"tag"`
}

func (h *Handler) Deploy(ctx context.Context, req DeployRequest) (GetDeploymentResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return GetDeploymentResponse{}, rpcErr
	}

	repo, err := h.db.GetRepoByID(ctx, profile.UserInfo.CurrentWorkspace, req.RepoID)
	if err != nil {
		return GetDeploymentResponse{}, &vel.Error{
			Message: "failed to get installation id for a repo",
			Err:     err,
		}
	}
	workspace, err := h.db.GetWorkspaceByID(ctx, profile.UserInfo.CurrentWorkspace)
	if err != nil {
		if errors.Is(err, ErrWorkspaceNotFound) {
			return GetDeploymentResponse{}, &vel.Error{
				Code: "WORKSPACE_NOT_FOUND",
			}
		}
		return GetDeploymentResponse{}, &vel.Error{
			Message: "failed to get workspace info",
			Err:     err,
		}
	}

	appDeployment, apiErr := h.deployRepo(
		ctx,
		profile.UserInfo.DisplayName,
		workspace,
		repo,
		req.FromDeploymentID,
		req.Branch,
		req.Sha,
		req.Tag,
	)
	if apiErr != nil {
		return GetDeploymentResponse{}, apiErr
	}

	return GetDeploymentResponse{
		Deployment: appDeployment,
	}, nil
}
