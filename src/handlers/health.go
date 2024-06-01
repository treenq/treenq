package handlers

import (
	"context"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Health(ctx context.Context, _ struct{}) (struct{}, *Error) {
	return struct{}{}, nil
}
