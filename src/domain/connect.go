package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/treenq/treenq/pkg/vel"
)

type ConnectRequest struct {
	ID int `json:"id"`
}

type ConnectResponse struct {
}

type WebhookResponse struct {
	ID int `json:"id"`
}

func (h *Handler) ConnectRepository(ctx context.Context, req ConnectRequest) (ConnectResponse, *vel.Error) {
	email := h.authProfiler.GetProfile(ctx).Email
	userRepos, err := h.db.GetConnectedRepositories(ctx, email)
	if err != nil {
		return ConnectResponse{}, &vel.Error{
			Code:    "REPO_NOT_FOUND",
			Message: "Repository not found",
		}
	}
	idx := slices.IndexFunc(userRepos, func(repo GithubRepository) bool {
		return repo.ID == req.ID
	})
	if idx == -1 {
		return ConnectResponse{}, &vel.Error{
			Code:    "REPO_NOT_FOUND",
			Message: "Repository not found",
		}
	}

	// create github webhook
	webhookBody := map[string]interface{}{
		"name":   "tq-listener",
		"events": []string{"push", "pull_request"},
		"config": map[string]interface{}{
			"content_type": "json",
			"url":          h.githubWebhookURL,
			"secret":       h.githubWebhookSecret,
		},
	}
	body, err := json.Marshal(webhookBody)
	if err != nil {
		return ConnectResponse{}, &vel.Error{
			Code:    "INVALID_WEBHOOK_BODY",
			Message: "Invalid webhook body",
		}
	}
	r, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://api.github.com/repos/%s/hooks", userRepos[idx].FullName), bytes.NewBuffer(body))
	if err != nil {
		return ConnectResponse{}, &vel.Error{
			Code:    "INVALID_WEBHOOK_REQUEST",
			Message: "Invalid webhook request",
		}
	}
	tokenPair, err := h.db.GetTokenPair(ctx, email)
	if err != nil {
		return ConnectResponse{}, &vel.Error{
			Code:    "INVALID_TOKEN_PAIR",
			Message: "Invalid token pair",
		}
	}
	r.Header.Set("Authorization", "Bearer "+tokenPair)
	r.Header.Set("Accept", "application/vnd.github+json")
	r.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return ConnectResponse{}, &vel.Error{
			Code:    "FAILED_TO_CREATE_WEBHOOK",
			Message: "Failed to create webhook",
		}
	}
	defer resp.Body.Close()

	var result WebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ConnectResponse{}, &vel.Error{
			Code:    "INVALID_WEBHOOK_RESPONSE",
			Message: "Invalid webhook response",
		}
	}

	userRepos[idx].WebhookID = result.ID
	if err := h.db.SaveConnectedRepository(ctx, email, userRepos[idx]); err != nil {
		return ConnectResponse{}, &vel.Error{
			Code:    "FAILED_TO_SAVE_WEBHOOK",
			Message: "Failed to save webhook",
		}
	}
	return ConnectResponse{}, nil
}
