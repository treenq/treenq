package handlers

import (
	"context"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
)

type DeployRequest struct {
}

type DeployResponse struct {
}

func Deploy(ctx context.Context, req DeployRequest) (DeployResponse, *Error) {
	return DeployResponse{}, nil
}

type App struct {
	ID string
	tqsdk.App
}
