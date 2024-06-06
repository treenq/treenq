package handlers

import (
	"context"
	"fmt"
)

type GithubWebhookRequest struct {
	F string
}

type GithubWebhookResponse struct {
	Else string
}

func GithubWebhook(ctx context.Context, req GithubWebhookRequest) (GithubWebhookResponse, *Error) {
	fmt.Println(req)
	return GithubWebhookResponse{}, nil
}
