package handlers

import (
	"context"
	"fmt"
)

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Meta    map[string]string `json:"meta"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s, %s", e.Code, e.Message)
}

func Health(ctx context.Context, _ struct{}) (struct{}, *Error) {
	return struct{}{}, nil
}
