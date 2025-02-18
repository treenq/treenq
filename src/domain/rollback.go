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
			Message: "failed to get deployment history",
			Err:     err,
		}
	}

	for i := range history {
		if history[i].Tag == req.Tag {
			if velErr := h.deployDefinition(history[i]); velErr != nil {
				return RollbackResponse{}, velErr
			}
		}

		if history[i].Sha == req.Sha {
			if velErr := h.deployDefinition(history[i]); velErr != nil {
				return RollbackResponse{}, velErr
			}
		}
	}

	return RollbackResponse{
		History: history,
	}, nil
}

func (h *Handler) deployDefinition(def AppDefinition) *vel.Error {
	panic("NOT IMPLEMENTED")
	return nil
	// apply
	// update
}
