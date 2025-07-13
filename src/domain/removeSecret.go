package domain

import (
	"context"

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

	if err := h.db.RemoveSecret(ctx, req.RepoID, req.Key, profile.UserInfo.DisplayName); err != nil {
		return struct{}{}, &vel.Error{
			Message: "failed to remove secret from database",
			Err:     err,
		}
	}

	err := h.kube.RemoveSecret(ctx, h.kubeConfig, profile.UserInfo.DisplayName, req.RepoID, req.Key)
	if err != nil {
		return struct{}{}, &vel.Error{
			Message: "failed to remove secret from Kubernetes",
			Err:     err,
		}
	}

	return struct{}{}, nil
}
