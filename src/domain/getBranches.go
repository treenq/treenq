package domain

import (
	"context"
	"errors"

	"github.com/dennypenta/vel"
)

type GetBranchesRequest struct {
	RepoName string `json:"repoName"`
}

type GetBranchesResopnse struct {
	Branches []string `json:"branches"`
}

func (h *Handler) GetBranches(ctx context.Context, req GetBranchesRequest) (GetBranchesResopnse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return GetBranchesResopnse{}, rpcErr
	}
	githubInstallationID, err := h.db.GetInstallationID(ctx, profile.UserInfo.CurrentWorkspace, req.RepoName)
	if err != nil {
		if errors.Is(err, ErrInstallationNotFound) {
			return GetBranchesResopnse{}, &vel.Error{
				Code: "INSTALLATION_NOT_FOUND",
			}
		}
		return GetBranchesResopnse{}, &vel.Error{
			Message: "failed to get installation",
			Err:     err,
		}
	}

	branches, err := h.githubClient.GetBranches(ctx, githubInstallationID, req.RepoName, true)
	if err != nil {
		return GetBranchesResopnse{}, &vel.Error{
			Message: "failed to get branches",
			Err:     err,
		}
	}

	return GetBranchesResopnse{Branches: branches}, nil
}
