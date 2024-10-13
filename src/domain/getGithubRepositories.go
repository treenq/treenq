package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"time"

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
}

// GetGithubRepositories returns a list of repositories for the authenticated user using the saved access token
func (h *Handler) GetGithubRepositories(ctx context.Context, req GetGithubRepositoriesRequest) (GetGithubRepositoriesResponse, *vel.Error) {
	email := h.authProfiler.GetProfile(ctx).Email
	tokenPair, ok := tokens[email]
	if !ok {
		return GetGithubRepositoriesResponse{}, &vel.Error{
			Code:    "NO_GITHUB_TOKEN",
			Message: "No access token found for user",
		}
	}

	if time.Now().Add(time.Second * 3).After(tokenPair.ExpiresIn) {
		// refresh token
		body := map[string]string{
			"client_id":     h.githubClientID,
			"client_secret": h.githubSecret,
			"grant_type":    "refresh_token",
			"refresh_token": tokenPair.RefreshToken,
		}
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return GetGithubRepositoriesResponse{}, &vel.Error{
				Code:    "INVALID_GITHUB_TOKEN_BODY",
				Message: "Invalid refresh token body",
			}
		}
		r, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewReader(jsonBody))
		if err != nil {
			return GetGithubRepositoriesResponse{}, &vel.Error{
				Code:    "INVALID_GITHUB_TOKEN_REQUEST",
				Message: "Invalid refresh token request",
			}
		}

		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			return GetGithubRepositoriesResponse{}, &vel.Error{
				Code:    "FAILED_TO_REFRESH_GITHUB_TOKEN",
				Message: "Failed to refresh github token",
			}
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return GetGithubRepositoriesResponse{}, &vel.Error{
				Code:    "INVALID_GITHUB_TOKEN_RESPONSE",
				Message: "Invalid github token response",
			}
		}

		tokenPair.AccessToken = result["access_token"].(string)
		tokenPair.RefreshToken = result["refresh_token"].(string)
		tokenPair.ExpiresIn = time.Now().Add(time.Duration(result["expires_in"].(int)) * time.Second)
		tokens[email] = tokenPair
	}

	r, err := http.NewRequest("GET", "https://api.github.com/repositories", nil)
	if err != nil {
		return GetGithubRepositoriesResponse{}, &vel.Error{
			Code:    "INVALID_GITHUB_REPOSITORIES_REQUEST",
			Message: "Invalid github repositories request",
		}
	}
	r.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
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

	repos[email] = filtered
	return GetGithubRepositoriesResponse{Repositories: filtered}, nil
}
