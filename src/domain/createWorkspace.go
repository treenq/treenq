package domain

import (
	"context"

	"github.com/dennypenta/vel"
)

type CreateWorkspaceRequest struct {
	WorkspaceName string `json:"workspaceName"`
}

type CreateWorkspaceResponse struct {
	CreatedWorkspace Workspace `json:"createdWorkspace"`
}

func (h *Handler) CreateWorkspace(ctx context.Context, req CreateWorkspaceRequest) (CreateWorkspaceResponse, *vel.Error) {
	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return CreateWorkspaceResponse{}, rpcErr
	}

	workspace, err := h.db.CreateWorkspace(ctx, profile.UserInfo.ID, req.WorkspaceName)
	if err != nil {
		return CreateWorkspaceResponse{}, &vel.Error{
			Message: "failed to create workspace",
			Err:     err,
		}
	}

	return CreateWorkspaceResponse{
		CreatedWorkspace: workspace,
	}, nil
}
