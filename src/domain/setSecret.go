package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type SecretKVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SetSecretsRequest struct {
	RepoID  string         `json:"repoID"`
	Secrets []SecretKVPair `json:"secrets"`
}

func (h *Handler) SetSecrets(ctx context.Context, req SetSecretsRequest) (struct{}, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return struct{}{}, rpcErr
	}

	for _, secret := range req.Secrets {
		err := h.kube.StoreSecret(ctx, h.kubeConfig, profile.UserInfo.DisplayName, req.RepoID, secret.Key, secret.Value)
		if err != nil {
			return struct{}{}, &vel.Error{
				Message: "failed to store secret: " + secret.Key,
				Err:     err,
			}
		}

		if err := h.db.SaveSecret(ctx, req.RepoID, secret.Key, profile.UserInfo.DisplayName); err != nil {
			return struct{}{}, &vel.Error{
				Message: "failed to save secret: " + secret.Key,
				Err:     err,
			}
		}
	}

	return struct{}{}, nil
}
