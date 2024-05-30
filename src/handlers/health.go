package handlers

import (
	"context"
)

type Error struct {
	Status int `json:"-"`

	Code    string `json:"code"`
	Message string `json:"message"`
}

func Health(ctx context.Context, _ struct{}) (struct{}, *Error) {
	return struct{}{}, nil
}
