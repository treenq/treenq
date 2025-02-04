package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type GetReposRequest struct {
}

type GetReposResponse struct {
}

func (h *Handler) GetRepos(ctx context.Context, req GetReposRequest) (GetReposResponse, *vel.Error) {
	return GetReposResponse{}, nil
}

