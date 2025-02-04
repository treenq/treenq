package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type ConnectBranchRequest struct {
	RepoID int
	Branch string
}

type ConnectBranchResponse struct{}

func (h *Handler) ConnectBranch(ctx context.Context, req ConnectBranchRequest) (ConnectBranchResponse, *vel.Error) {
	return ConnectBranchResponse{}, nil
}
