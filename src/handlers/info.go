package handlers

import (
	"context"
)

var version = "develop" // unstable

type InfoResponse struct {
	Version string `json:"version"`
}

func (h *Handler) Info(ctx context.Context, _ struct{}) (InfoResponse, *Error) {
	resp := InfoResponse{
		Version: version,
	}
	return resp, nil
}
