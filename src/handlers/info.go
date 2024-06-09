package handlers

import "context"


var Version = "develop" // unstable


type InfoResponse struct {
	Version string `json:"version"`
}

func Info(ctx context.Context, _ struct{}) (InfoResponse, *Error) {
	resp := InfoResponse{
		Version: Version,
	}
	return resp, nil
}
