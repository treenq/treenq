package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type RollbackRequest struct {
	AppID string `json:"appId"`
	Tag   string `json:"tag"`
	Sha   string `json:"sha"`
}

type RollbackResponse struct {
	History []AppDefinition
}

func (h *Handler) Rollback(ctx context.Context, req RollbackRequest) (RollbackResponse, *vel.Error) {
	history, err := h.db.GetDeploymentHistory(ctx, req.AppID)
	if err != nil {
		return RollbackResponse{}, &vel.Error{
			Code:    "UNKNOWN",
			Message: err.Error(),
		}
	}

	for i := range history {
		if history[i].Tag == req.Tag {
			if err := h.deployDefinition(history[i]); err != nil {
				return RollbackResponse{}, &vel.Error{
					Code:    "UNKNOWN",
					Message: err.Error(),
				}
			}
		}

		if history[i].Sha == req.Sha {
			if err := h.deployDefinition(history[i]); err != nil {
				return RollbackResponse{}, &vel.Error{
					Code:    "UNKNOWN",
					Message: err.Error(),
				}
			}
		}
	}

	return RollbackResponse{
		History: history,
	}, nil
}

func (h *Handler) deployDefinition(def AppDefinition) error {
	panic("NOT IMPLEMENTED")
	return nil
	// apply
	// update
}
