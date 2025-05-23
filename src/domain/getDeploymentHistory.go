package domain

import (
	"context"
	"time"

	"github.com/treenq/treenq/pkg/vel"
)

type GetDeploymentHistoryRequest struct {
	RepoID string
}

type DeploymentHistoryItem struct {
	ID              string    `json:"id"`
	RepoID          string    `json:"repoId"`
	Status          string    `json:"status"`
	CommitHash      string    `json:"commitHash"`
	Message         string    `json:"message"`
	Timestamp       time.Time `json:"timestamp"`
	UserDisplayName string    `json:"userDisplayName"`
}

type GetDeploymentHistoryResponse struct {
	History []DeploymentHistoryItem `json:"history"`
}

func (h *Handler) GetDeploymentHistory(ctx context.Context, req GetDeploymentHistoryRequest) (GetDeploymentHistoryResponse, *vel.Error) {
	deployments, err := h.db.GetDeploymentHistory(ctx, req.RepoID)
	if err != nil {
		return GetDeploymentHistoryResponse{}, &vel.Error{
			Message: "failed get deployment history",
			Err:     err,
		}
	}

	historyItems := make([]DeploymentHistoryItem, len(deployments))
	for i, dep := range deployments {
		historyItems[i] = DeploymentHistoryItem{
			ID:              dep.ID,
			RepoID:          dep.RepoID,
			Status:          dep.Status,
			CommitHash:      dep.Sha,
			Message:         dep.BuildTag,
			Timestamp:       dep.CreatedAt,
			UserDisplayName: dep.UserDisplayName,
		}
	}

	return GetDeploymentHistoryResponse{
		History: historyItems,
	}, nil
}
