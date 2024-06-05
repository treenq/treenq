package handlers

import (
	"context"
)

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Meta    map[string]string `json:"meta"`
}

func Health(ctx context.Context, _ struct{}) (struct{}, *Error) {
	return struct{}{}, nil
}
