package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
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

	err := h.kube.StoreSecret(ctx, h.kubeConfig, profile.UserInfo.DisplayName, req.RepoID, req.Key, req.Value)
	if err != nil {
		return struct{}{}, &vel.Error{
			Message: "failed to store secret",
			Err:     err,
		}
	}

	if err := h.db.SaveSecret(ctx, req.RepoID, req.Key, profile.UserInfo.DisplayName); err != nil {
		return struct{}{}, &vel.Error{
			Message: "failed to save secret",
			Err:     err,
		}
	}

	return struct{}{}, nil
}
