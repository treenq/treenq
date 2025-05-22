package domain

import (
	"context"
	"time" // Ensure time is imported

	"github.com/treenq/treenq/pkg/vel"
)

type DeployRequest struct {
	RepoID string `json:"repoID"`
}

type DeployResponse struct {
	DeploymentID string    `json:"deploymentID"`
	Status       string    `json:"status"`    // New field
	CreatedAt    time.Time `json:"createdAt"` // New field
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

	appDeployment, apiErr := h.deployRepo(
		ctx,
		profile.UserInfo.DisplayName,
		repo,
	)
	if apiErr != nil {
		return DeployResponse{}, apiErr
	}

	return DeployResponse{
		DeploymentID: appDeployment.ID,
		Status:       appDeployment.Status,
		CreatedAt:    appDeployment.CreatedAt,
	}, nil
}
