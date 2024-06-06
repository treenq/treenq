package handlers

import (
	"context"
	"fmt"
)

type GithubWebhookRequest struct {
}

type GithubWebhookResponse struct {
}

func GithubWebhook(ctx context.Context, req GithubWebhookRequest) (GithubWebhookResponse, *Error) {
	fmt.Println(req)
	return GithubWebhookResponse{}, nil
}
