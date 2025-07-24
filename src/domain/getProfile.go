package domain

import (
	"context"
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

	// Extract workspaces from claims
	var workspaces []string
	if workspacesRaw, exists := claims["workspaces"]; exists && workspacesRaw != nil {
		if workspacesList, ok := workspacesRaw.([]interface{}); ok {
			workspaces = make([]string, len(workspacesList))
			for i, workspace := range workspacesList {
				if workspaceStr, ok := workspace.(string); ok {
					workspaces[i] = workspaceStr
				}
			}
		}
	}

	var currentWorkspace string
	if len(workspaces) == 1 {
		currentWorkspace = workspaces[0]
	} else {
		r := vel.RequestFromContext(ctx)
		currentWorkspace = r.Header.Get(treenq.WorkspaceHeader)
		if currentWorkspace == "" {
			return GetProfileResponse{}, &vel.Error{
				Code: "CURRENT_WORKSPACE_HEADER_REQUIRED",
			}
		}
		if !slices.Contains(workspaces, currentWorkspace) {
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
