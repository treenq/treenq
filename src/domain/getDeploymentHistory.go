package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type GetDeploymentHistoryRequest struct {
	RepoID string
}

type GetDeploymentHistoryResponse struct {
	History []AppDeployment
}

func (h *Handler) GetDeploymentHistory(ctx context.Context, req GetDeploymentHistoryRequest) (GetDeploymentHistoryResponse, *vel.Error) {
	history, err := h.db.GetDeploymentHistory(ctx, req.RepoID)
	if err != nil {
		return GetDeploymentHistoryResponse{}, &vel.Error{
			Message: "failed get deployment history",
			Err:     err,
		}
	}

	return GetDeploymentHistoryResponse{
		History: history,
	}, nil
}
