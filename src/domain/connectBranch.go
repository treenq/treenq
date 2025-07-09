package domain

import (
	"context"
	"errors"

	"github.com/treenq/treenq/pkg/vel"
)

type ConnectBranchRequest struct {
	RepoID string `json:"repoID"`
	Branch string `json:"branch"`
}

type ConnectBranchResponse struct {
	Repo GithubRepository `json:"repo"`
}

var (
	ErrNoTqJsonFound    = errors.New("tq.json not found")
	ErrTqIsNotValidJson = errors.New("tq.json contains invalid json")
	ErrNoSpaceFound     = errors.New("no space found")
)

func (h *Handler) ConnectBranch(ctx context.Context, req ConnectBranchRequest) (ConnectBranchResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return ConnectBranchResponse{}, rpcErr
	}

	repo, err := h.db.GetRepoByID(ctx, profile.UserInfo.ID, req.RepoID)
	if err != nil {
		if errors.Is(err, ErrRepoNotFound) {
			return ConnectBranchResponse{}, &vel.Error{
				Code: "REPO_NOT_FOUND",
			}
		}

		return ConnectBranchResponse{}, &vel.Error{
			Err:     err,
			Message: "failed to get repo",
		}
	}

	installationID, err := h.db.GetInstallationID(ctx, profile.UserInfo.ID, repo.FullName)
	if err != nil {
		if errors.Is(err, ErrInstallationNotFound) {
			return ConnectBranchResponse{}, &vel.Error{
				Code: "INSTALLATION_NOT_FOUND",
			}
		}
		return ConnectBranchResponse{}, &vel.Error{
			Message: "failed to get installation",
			Err:     err,
		}
	}

	space, err := h.githubClient.GetRepoSpace(ctx, installationID, repo.FullName, req.Branch)
	if err != nil {
		if errors.Is(err, ErrNoTqJsonFound) {
			return ConnectBranchResponse{}, &vel.Error{
				Code: "TQ_JSON_NOT_FOUND",
			}
		}
		if errors.Is(err, ErrTqIsNotValidJson) {
			return ConnectBranchResponse{}, &vel.Error{
				Code: "TQ_JSON_INVALID",
			}
		}
		return ConnectBranchResponse{}, &vel.Error{
			Message: "failed to get space from github",
		}
	}

	repo, err = h.db.ConnectRepo(ctx, profile.UserInfo.ID, req.RepoID, req.Branch, space)
	if err != nil {
		if errors.Is(err, ErrRepoNotFound) {
			return ConnectBranchResponse{}, &vel.Error{
				Code: "REPO_NOT_FOUND",
			}
		}

		return ConnectBranchResponse{}, &vel.Error{
			Err:     err,
			Message: "failed to connect repo",
		}
	}

	// _, deployErr := h.deployRepo(ctx, profile.UserInfo.DisplayName, repo, "", "", "", "")
	// if deployErr != nil {
	// 	return ConnectBranchResponse{}, deployErr
	// }
	return ConnectBranchResponse{Repo: repo}, nil
}
