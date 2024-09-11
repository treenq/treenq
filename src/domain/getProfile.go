package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

type GetProfileResponse struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

func (h *Handler) GetProfile(ctx context.Context, _ struct{}) (GetProfileResponse, *vel.Error) {
	profile := h.authProfiler.GetProfile(ctx)
	return GetProfileResponse{
		Email:    profile.Email,
		Username: profile.Username,
		Name:     profile.Name,
	}, nil
}
