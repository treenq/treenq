package domain

import (
	"context"

	"github.com/treenq/treenq/pkg/vel"
)

var version = "develop" // unstable

type InfoResponse struct {
	Version string `json:"version"`
}

func (h *Handler) Info(ctx context.Context, _ struct{}) (InfoResponse, *vel.Error) {
	resp := InfoResponse{
		Version: version,
	}
	return resp, nil
}
