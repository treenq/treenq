package domain

import (
	"context"
	"errors"
	"slices"

	"github.com/dennypenta/vel"
	"github.com/treenq/treenq/pkg/auth"
	"github.com/treenq/treenq/pkg/treenq"
)

type GetProfileResponse struct {
	UserInfo UserInfo `json:"userInfo"`
}

func (h *Handler) GetProfile(ctx context.Context, _ struct{}) (GetProfileResponse, *vel.Error) {
	claims := auth.ClaimsFromCtx(ctx)

	userID := claims["id"].(string)
	workspaces, err := h.db.GetUserWorkspaces(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrWorkspaceNotFound) {
			return GetProfileResponse{}, &vel.Error{
				Code: "WORKSPACE_NOT_FOUND",
			}
		}
		return GetProfileResponse{}, &vel.Error{
			Message: "failed to get workspace info",
			Err:     err,
		}
	}

	var currentWorkspace string
	if len(workspaces) == 1 {
		currentWorkspace = workspaces[0].ID
	} else {
		r := vel.RequestFromContext(ctx)
		currentWorkspace = r.Header.Get(treenq.WorkspaceHeader)
		if currentWorkspace == "" {
			return GetProfileResponse{}, &vel.Error{
				Code: "CURRENT_WORKSPACE_HEADER_REQUIRED",
			}
		}
		workspaceIDs := make([]string, len(workspaces))
		for i := range workspaces {
			workspaceIDs[i] = workspaces[i].ID
		}
		if !slices.Contains(workspaceIDs, currentWorkspace) {
			return GetProfileResponse{}, &vel.Error{
				Code: "CURRENT_WORKSPACE_HEADER_REQUIRED",
			}
		}
	}

	return GetProfileResponse{
		UserInfo: UserInfo{
			ID:               claims["id"].(string),
			Email:            claims["email"].(string),
			DisplayName:      claims["displayName"].(string),
			CurrentWorkspace: currentWorkspace,
			Workspaces:       workspaces,
		},
	}, nil
}
