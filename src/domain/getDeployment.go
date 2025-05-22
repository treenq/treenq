package domain

import (
	"context"
	"errors"

	"github.com/treenq/treenq/pkg/vel"
)

var ErrDeploymentNotFound = errors.New("deployment not found")

type GetDeploymentRequest struct {
	DeploymentID string `path:"deploymentID"` // Changed from json: to path:
}

type GetDeploymentResponse struct {
	Deployment AppDeployment `json:"deployment"`
}

func (h *Handler) GetDeployment(ctx context.Context, req GetDeploymentRequest) (GetDeploymentResponse, *vel.Error) {
	deployment, err := h.db.GetDeployment(ctx, req.DeploymentID)
	if err != nil {
		if errors.Is(err, ErrDeploymentNotFound) {
			return GetDeploymentResponse{}, &vel.Error{
				Code: "DEPLOYMENT_NOT_FOUND",
				Err:  err,
			}
		}
		return GetDeploymentResponse{}, &vel.Error{
			Code: "FAILED_GET_DEPLOYMENT",
			Err:  err,
		}
	}

	return GetDeploymentResponse{Deployment: deployment}, nil
}
