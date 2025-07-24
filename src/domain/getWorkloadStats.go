package domain

import (
	"context"
	"errors"

	"github.com/dennypenta/vel"
)

type GetWorkloadStatsRequest struct {
	RepoID string `json:"repoID"`
}

type GetWorkloadStatsResponse struct {
	WorkloadStats WorkloadStats `json:"workloadStats"`
}

type WorkloadStats struct {
	Name          string        `json:"name"`
	Replicas      Replicas      `json:"replicas"`
	Versions      []VersionInfo `json:"versions"`
	OverallStatus string        `json:"overallStatus"`
}

type Replicas struct {
	Desired int `json:"desired"`
	Running int `json:"running"`
	Pending int `json:"pending"`
	Failed  int `json:"failed"`
}

type VersionInfo struct {
	Version  string      `json:"version"`
	Replicas ReplicaInfo `json:"replicas"`
}

type ReplicaInfo struct {
	Running int `json:"running"`
	Pending int `json:"pending"`
	Failed  int `json:"failed"`
}

func (h *Handler) GetWorkloadStats(ctx context.Context, req GetWorkloadStatsRequest) (GetWorkloadStatsResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return GetWorkloadStatsResponse{}, rpcErr
	}

	workspace, err := h.db.GetWorkspaceByID(ctx, profile.UserInfo.CurrentWorkspace)
	if err != nil {
		if errors.Is(err, ErrWorkspaceNotFound) {
			return GetWorkloadStatsResponse{}, &vel.Error{
				Code: "WORKSPACE_NOT_FOUND",
			}
		}

		return GetWorkloadStatsResponse{}, &vel.Error{
			Message: "failed to get workspace info",
			Err:     err,
		}
	}

	stats, err := h.kube.GetWorkloadStats(ctx, h.kubeConfig, req.RepoID, workspace.Name)
	if errors.Is(err, ErrNoPodsRunning) {
		return GetWorkloadStatsResponse{}, &vel.Error{
			Code: "NO_PODS_RUNNING",
		}
	}
	if err != nil {
		return GetWorkloadStatsResponse{}, &vel.Error{
			Message: "failed to get workload stats",
			Err:     err,
		}
	}

	return GetWorkloadStatsResponse{
		WorkloadStats: stats,
	}, nil
}
