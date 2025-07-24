package domain

import (
	"context"
	"errors"

	"github.com/dennypenta/vel"
)

type RemoveSecretRequest struct {
	RepoID string `json:"repoID"`
	Key    string `json:"key"`
}

func (h *Handler) RemoveSecret(ctx context.Context, req RemoveSecretRequest) (struct{}, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return struct{}{}, rpcErr
	}

	if err := h.db.RemoveSecret(ctx, req.RepoID, req.Key, profile.UserInfo.CurrentWorkspace); err != nil {
		return struct{}{}, &vel.Error{
			Message: "failed to remove secret from database",
			Err:     err,
		}
	}

	workspace, err := h.db.GetWorkspaceByID(ctx, profile.UserInfo.CurrentWorkspace)
	if err != nil {
		if errors.Is(err, ErrWorkspaceNotFound) {
			return struct{}{}, &vel.Error{
				Code: "WORKSPACE_NOT_FOUND",
			}
		}

		return struct{}{}, &vel.Error{
			Message: "failed to get workspace info",
			Err:     err,
		}
	}

	err = h.kube.RemoveSecret(ctx, h.kubeConfig, workspace.Name, req.RepoID, req.Key)
	if err != nil {
		return struct{}{}, &vel.Error{
			Message: "failed to remove secret from Kubernetes",
			Err:     err,
		}
	}

	return struct{}{}, nil
}
