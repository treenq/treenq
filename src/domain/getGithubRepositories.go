package domain

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/treenq/treenq/pkg/vel"
)

type GetGithubRepositoriesRequest struct {
}

type GetGithubRepositoriesResponse struct {
	Repositories []GithubRepository
}

type GithubRepository struct {
	ID int `json:"id"`
	// TODO: perhaps remove if only FullName is used
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Permissions struct {
		Admin bool `json:"admin"`
		Push  bool `json:"push"`
		Pull  bool `json:"pull"`
	} `json:"permissions"`
	IsTemplate    bool   `json:"is_template"`
	DefaultBranch string `json:"default_branch"`

	Connected bool `json:"connected"`
	WebhookID int  `json:"webhook_id"`
}

// GetGithubRepositories returns a list of repositories for the authenticated user using the saved access token
func (h *Handler) GetGithubRepositories(ctx context.Context, req GetGithubRepositoriesRequest) (GetGithubRepositoriesResponse, *vel.Error) {
	profile, profileErr := h.GetProfile(ctx, struct{}{})
	if profileErr != nil {
		return GetGithubRepositoriesResponse{}, profileErr
	}
	email := profile.UserInfo.Email
	tokenPair, err := h.db.GetTokenPair(ctx, email)
	if err != nil {
		return GetGithubRepositoriesResponse{}, &vel.Error{
			Code:    "NO_GITHUB_TOKEN",
			Message: "No access token found for user",
		}
	}

	r, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repositories", nil)
	if err != nil {
		return GetGithubRepositoriesResponse{}, &vel.Error{
			Code:    "INVALID_GITHUB_REPOSITORIES_REQUEST",
			Message: "Invalid github repositories request",
		}
	}
	r.Header.Set("Authorization", "Bearer "+tokenPair)
	r.Header.Set("Accept", "application/vnd.github+json")
	r.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return GetGithubRepositoriesResponse{}, &vel.Error{
			Code:    "FAILED_TO_GET_GITHUB_REPOSITORIES",
			Message: "Failed to get github repositories",
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusForbidden {
		return GetGithubRepositoriesResponse{}, &vel.Error{
			Code:    "GITHUB_TOKEN_EXPIRED",
			Message: "GitHub token has expired. Please sign in again.",
		}
	}

	var result []GithubRepository
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return GetGithubRepositoriesResponse{}, &vel.Error{
			Code:    "INVALID_GITHUB_REPOSITORIES_RESPONSE",
			Message: "Invalid github repositories response",
		}
	}

	connectedRepos, err := h.db.GetConnectedRepositories(ctx, email)
	if err != nil {
		return GetGithubRepositoriesResponse{}, &vel.Error{
			Code:    "FAILED_TO_GET_CONNECTED_REPOSITORIES",
			Message: "Failed to get connected repositories",
		}
	}

	filtered := make([]GithubRepository, 0, len(result))
	for _, repo := range result {
		if repo.IsTemplate {
			continue
		}
		if !repo.Permissions.Pull {
			continue
		}
		filtered = append(filtered, repo)
	}

	for i := range filtered {
		filtered[i].Connected = slices.ContainsFunc(connectedRepos, func(repo GithubRepository) bool {
			return repo.ID == filtered[i].ID
		})
	}

	if err := h.db.SaveGithubRepos(r.Context(), email, filtered); err != nil {
		return GetGithubRepositoriesResponse{}, &vel.Error{
			Code:    "FAILED_TO_SAVE_GITHUB_REPOSITORIES",
			Message: "Failed to save github repositories",
		}
	}
	return GetGithubRepositoriesResponse{Repositories: filtered}, nil
}
