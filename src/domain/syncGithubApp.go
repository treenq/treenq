package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

func (h *Handler) SyncGithubApp(ctx context.Context, _ struct{}) (GetReposResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return GetReposResponse{}, rpcErr
	}
	githubInstallation, err := h.githubClient.GetUserInstallation(ctx, profile.UserInfo.DisplayName)
	if err != nil {
		return GetReposResponse{}, &vel.Error{
			Message: "failed to get a github installation",
			Err:     err,
		}
	}
	installedRepos, err := h.githubClient.ListRepositories(ctx, githubInstallation)
	if err != nil {
		return GetReposResponse{}, &vel.Error{
			Message: "failed to get github repos",
			Err:     err,
		}
	}

	savedInstallation, err := h.db.LinkGithub(ctx, githubInstallation, profile.UserInfo.DisplayName, installedRepos)
	if err != nil {
		return GetReposResponse{}, &vel.Error{
			Message: "failed to sync a github repos link",
			Err:     err,
		}
	}

	repos, err := h.db.GetGithubRepos(ctx, profile.UserInfo.ID)
	if err != nil {
		return GetReposResponse{}, &vel.Error{
			Code: "FAILED_GET_GITHUB_REPOS",
			Err:  err,
		}
	}

	return GetReposResponse{
		Installation: savedInstallation,
		Repos:        repos,
	}, nil
}
