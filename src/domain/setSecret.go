package domain

import (
	"context"
	"errors"

	"github.com/dennypenta/vel"
)

type SetSecretRequest struct {
	RepoID string `json:"repoID"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

func (h *Handler) SetSecret(ctx context.Context, req SetSecretRequest) (struct{}, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return struct{}{}, rpcErr
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

	err = h.kube.StoreSecret(ctx, h.kubeConfig, workspace.Name, req.RepoID, req.Key, req.Value)
	if err != nil {
		return struct{}{}, &vel.Error{
			Message: "failed to store secret",
			Err:     err,
		}
	}

	if err := h.db.SaveSecret(ctx, req.RepoID, req.Key, profile.UserInfo.CurrentWorkspace); err != nil {
		return struct{}{}, &vel.Error{
			Message: "failed to save secret",
			Err:     err,
		}
	}

	return struct{}{}, nil
}
