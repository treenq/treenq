package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type GetDeploymentsRequest struct {
	RepoID string `json:"repoID"`
}

type GetDeploymentsResponse struct {
	Deployments []AppDeployment `json:"deployments"`
}

func (h *Handler) GetDeployments(ctx context.Context, req GetDeploymentsRequest) (GetDeploymentsResponse, *vel.Error) {
	history, err := h.db.GetDeployments(ctx, req.RepoID)
	if err != nil {
		return GetDeploymentsResponse{}, &vel.Error{
			Message: "failed get deployment history",
			Err:     err,
		}
	}

	return GetDeploymentsResponse{
		Deployments: history,
	}, nil
}
