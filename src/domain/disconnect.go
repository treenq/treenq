package domain

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/treenq/treenq/pkg/vel"
)

type DisconnectRequest struct {
	ID int `json:"id"`
}

type DisconnectResponse struct {
}

func (h *Handler) DisconnectRepository(ctx context.Context, req DisconnectRequest) (DisconnectResponse, *vel.Error) {
	profile, profileErr := h.GetProfile(ctx, struct{}{})
	if profileErr != nil {
		return DisconnectResponse{}, profileErr
	}
	email := profile.UserInfo.Email
	userRepos, err := h.db.GetConnectedRepositories(ctx, email)
	if err != nil {
		return DisconnectResponse{}, &vel.Error{
			Code:    "REPO_NOT_FOUND",
			Message: "Repository not found",
		}
	}
	idx := slices.IndexFunc(userRepos, func(repo GithubRepository) bool {
		return repo.ID == req.ID
	})
	if idx == -1 {
		return DisconnectResponse{}, &vel.Error{
			Code:    "REPO_NOT_FOUND",
			Message: "Repository not found",
		}
	}

	// remove github webhook
	r, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("https://api.github.com/repos/%s/hooks/%d", userRepos[idx].FullName, userRepos[idx].WebhookID), nil)
	if err != nil {
		return DisconnectResponse{}, &vel.Error{
			Code:    "FAILED_TO_CREATE_REQUEST",
			Message: "Failed to create request",
		}
	}
	tokenPair, err := h.db.GetTokenPair(ctx, email)
	if err != nil {
		return DisconnectResponse{}, &vel.Error{
			Code:    "INVALID_TOKEN_PAIR",
			Message: "Invalid token pair",
		}
	}

	r.Header.Set("Authorization", "Bearer "+tokenPair)
	r.Header.Set("Accept", "application/vnd.github+json")
	r.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return DisconnectResponse{}, &vel.Error{
			Code:    "FAILED_TO_DELETE_WEBHOOK",
			Message: "Failed to delete webhook",
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return DisconnectResponse{}, &vel.Error{
			Code:    "FAILED_TO_DELETE_WEBHOOK",
			Message: "Failed to delete webhook",
		}
	}

	if err := h.db.RemoveConnectedRepository(ctx, email, userRepos[idx].ID); err != nil {
		return DisconnectResponse{}, &vel.Error{
			Code:    "FAILED_TO_REMOVE_CONNECTED_REPO",
			Message: "Failed to remove connected repository",
		}
	}
	return DisconnectResponse{}, nil
}
