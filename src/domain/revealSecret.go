package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type RevealSecretsRequest struct {
	RepoID string   `json:"repoID"`
	Keys   []string `json:"keys"`
}

type RevealSecretsResponse struct {
	Values map[string]string `json:"values"`
}

func (h *Handler) RevealSecrets(ctx context.Context, req RevealSecretsRequest) (RevealSecretsResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return RevealSecretsResponse{}, rpcErr
	}

	response := RevealSecretsResponse{
		Values: make(map[string]string),
	}

	for _, key := range req.Keys {
		exists, err := h.db.RepositorySecretKeyExists(ctx, req.RepoID, key, profile.UserInfo.DisplayName)
		if err != nil {
			return RevealSecretsResponse{}, &vel.Error{
				Message: "failed to lookup a secret key: " + key,
				Err:     err,
			}
		}
		if !exists {
			return RevealSecretsResponse{}, &vel.Error{
				Code:    "SECRET_DOESNT_EXIST",
				Message: "secret key does not exist: " + key,
			}
		}

		value, err := h.kube.GetSecret(ctx, h.kubeConfig, profile.UserInfo.DisplayName, req.RepoID, key)
		if err != nil {
			return RevealSecretsResponse{}, &vel.Error{
				Message: "failed to reveal secret: " + key,
				Err:     err,
			}
		}
		response.Values[key] = value
	}

	return response, nil
}
