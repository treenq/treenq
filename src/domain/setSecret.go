package domain

import (
	"context"
	"errors"
	"regexp"

	"github.com/dennypenta/vel"
)

type SetSecretRequest struct {
	RepoID string `json:"repoID"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

var secretKeyRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.-]*$`)

func validateSecretKey(key string) bool {
	if key == "" {
		return false
	}
	return secretKeyRegex.MatchString(key)
}

func (h *Handler) SetSecret(ctx context.Context, req SetSecretRequest) (struct{}, *vel.Error) {
	if !validateSecretKey(req.Key) {
		return struct{}{}, &vel.Error{
			Code: "INVALID_SECRET_KEY",
		}
	}

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
