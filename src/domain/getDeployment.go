package domain

import (
	"context"
	"errors"

	"github.com/dennypenta/vel"
)

var ErrDeploymentNotFound = errors.New("deployment not found")

type GetDeploymentRequest struct {
	DeploymentID string `json:"deploymentID"`
}

type GetDeploymentResponse struct {
	Deployment AppDeployment `json:"deployment"`
}

func (h *Handler) GetDeployment(ctx context.Context, req GetDeploymentRequest) (GetDeploymentResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return GetDeploymentResponse{}, rpcErr
	}

	deployment, err := h.db.GetDeployment(ctx, profile.UserInfo.CurrentWorkspace, req.DeploymentID)
	if err != nil {
		if errors.Is(err, ErrDeploymentNotFound) {
			return GetDeploymentResponse{}, &vel.Error{
				Code: "DEPLOYMENT_NOT_FOUND",
			}
		}
		return GetDeploymentResponse{}, &vel.Error{
			Message: "failed to get deployment",
			Err:     err,
		}
	}

	return GetDeploymentResponse{Deployment: deployment}, nil
}
