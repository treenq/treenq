package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type RevealSecretRequest struct {
	RepoID string `json:"repoID"`
	Key    string `json:"key"`
}

type RevealSecretResponse struct {
	Value string `json:"value"`
}

func (h *Handler) RevealSecret(ctx context.Context, req RevealSecretRequest) (RevealSecretResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return RevealSecretResponse{}, rpcErr
	}
	exists, err := h.db.RepositorySecretKeyExists(ctx, req.RepoID, req.Key, profile.UserInfo.DisplayName)
	if err != nil {
		return RevealSecretResponse{}, &vel.Error{
			Message: "failed to lookup a secret key",
			Err:     err,
		}
	}
	if !exists {
		return RevealSecretResponse{}, &vel.Error{
			Code: "SECRET_DOESNT_EXIST",
		}
	}

	value, err := h.kube.GetSecret(ctx, h.kubeConfig, profile.UserInfo.DisplayName, req.RepoID, req.Key)
	if err != nil {
		return RevealSecretResponse{}, &vel.Error{
			Message: "failed to reveal secret",
			Err:     err,
		}
	}

	return RevealSecretResponse{
		Value: value,
	}, nil
}
