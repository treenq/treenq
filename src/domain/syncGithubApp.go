package domain

import (
	"context"
	"errors"

	"github.com/treenq/treenq/pkg/vel"
)

func (h *Handler) SyncGithubApp(ctx context.Context, _ struct{}) (GetReposResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return GetReposResponse{}, rpcErr
	}

	// Get the user's GitHub access token from OAuth provider cache
	githubToken, err := h.oauthProvider.GetUserGithubToken(profile.UserInfo.DisplayName)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			return GetReposResponse{}, &vel.Error{
				Code:    "UNAUTHORIZED",
				Message: "GitHub token expired or invalid, please re-authenticate with GitHub",
			}
		}
		return GetReposResponse{}, &vel.Error{
			Message: "failed to get user GitHub token",
			Err:     err,
		}
	}

	allRepos, err := h.githubClient.ListAllRepositoriesForUser(ctx, githubToken)
	if err != nil {
		return GetReposResponse{}, &vel.Error{
			Message: "failed to get github repos for user installations",
			Err:     err,
		}
	}

	if len(allRepos) == 0 {
		return GetReposResponse{}, &vel.Error{
			Code:    "NO_INSTALLATIONS_FOUND",
			Message: "no accessible github app installations found for user",
		}
	}

	err = h.db.LinkAllGithubInstallations(ctx, profile.UserInfo, allRepos)
	if err != nil {
		if errors.Is(err, ErrInstallationNotFound) {
			return GetReposResponse{}, &vel.Error{
				Code: "INSTALLATION_NOT_FOUND",
			}
		}

		return GetReposResponse{}, &vel.Error{
			Message: "failed to sync github installations and repos",
			Err:     err,
		}
	}

	repos, hasInstallation, err := h.db.GetGithubRepos(ctx, profile.UserInfo.ID)
	if err != nil {
		return GetReposResponse{}, &vel.Error{
			Code: "FAILED_GET_GITHUB_REPOS",
			Err:  err,
		}
	}

	return GetReposResponse{
		Installation: hasInstallation,
		Repos:        repos,
	}, nil
}
