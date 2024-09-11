package domain

import (
	"context"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/pkg/vel"
)

type DeployRequest struct {
}

type DeployResponse struct {
}

func (h *Handler) Deploy(ctx context.Context, req DeployRequest) (DeployResponse, *vel.Error) {
	return DeployResponse{}, nil
}

type App struct {
	ID string
	tqsdk.Space
}
