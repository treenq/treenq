package handlers

import "context"

const version = "develop" // unstable

type InfoResponse struct {
	Version string `json:"version"`
}

func Info(ctx context.Context, _ struct{}) (InfoResponse, *Error) {
	resp := InfoResponse{
		Version: version,
	}
	return resp, nil
}
