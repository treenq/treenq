package domain

import (
	"context"

	"github.com/dennypenta/vel"
)

type CreateWorkspaceRequest struct {
	UserID        string `json:"userID"`
	WorkspaceName string `json:"workspaceName"`
}

type CreateWorkspaceResponse struct {
	CreatedWorkspace Workspace `json:"createdWorkspace"`
}

func (h *Handler) CreateWorkspace(ctx context.Context, req CreateWorkspaceRequest) (CreateWorkspaceResponse, *vel.Error) {
	workspace, err := h.db.CreateWorkspace(ctx, req.UserID, req.WorkspaceName)
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
