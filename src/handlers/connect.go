package handlers

import (
	"context"
)

type ConnectRequest struct {
	Url string `json:"url"`
}

type ConnectResponse struct {
	ID  string `json:"id"`
	Url string `json:"url"`
}

type RepoCreator interface {
	CreateRepo(ctx context.Context, req ConnectRequest) (ConnectResponse, error)
}

func NewConnect(repo RepoCreator) func(ctx context.Context, req ConnectRequest) (ConnectResponse, *Error) {
	return func(ctx context.Context, req ConnectRequest) (ConnectResponse, *Error) {
		res, err := repo.CreateRepo(ctx, req)
		if err != nil {
			return res, &Error{
				Code:    "UNKNOWN",
				Message: err.Error(),
			}
		}

		return res, nil
	}
}
