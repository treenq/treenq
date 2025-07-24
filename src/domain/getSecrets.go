package domain

import (
	"context"

	"github.com/dennypenta/vel"
)

type GetSecretsRequest struct {
	RepoID string `json:"repoID"`
}

type GetSecretsResponse struct {
	Keys []string `json:"keys"`
}

func (h *Handler) GetSecrets(ctx context.Context, req GetSecretsRequest) (GetSecretsResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return GetSecretsResponse{}, rpcErr
	}
	keys, err := h.db.GetRepositorySecretKeys(ctx, req.RepoID, profile.UserInfo.CurrentWorkspace)
	if err != nil {
		return GetSecretsResponse{}, &vel.Error{
			Message: "failed to get secrets keys",
			Err:     err,
		}
	}

	return GetSecretsResponse{
		Keys: keys,
	}, nil
}
