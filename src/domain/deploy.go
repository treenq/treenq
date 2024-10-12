package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type DeployRequest struct {
	Repo   string
	Branch string
	AppID  string
}

type DeployResponse struct {
}

func (h *Handler) Deploy(ctx context.Context, req DeployRequest) (DeployResponse, *vel.Error) {
	return DeployResponse{}, nil
}
