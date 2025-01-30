package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
	"github.com/treenq/treenq/pkg/vel/auth"
)

type GetProfileResponse struct {
	UserInfo UserInfo `json:"userInfo"`
}

func (h *Handler) GetProfile(ctx context.Context, _ struct{}) (GetProfileResponse, *vel.Error) {
	claims := auth.ClaimsFromCtx(ctx)
	return GetProfileResponse{
		UserInfo: UserInfo{
			ID:          claims["id"].(string),
			Email:       claims["email"].(string),
			DisplayName: claims["DisplayName"].(string),
		},
	}, nil
}
