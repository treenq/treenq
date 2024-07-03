package domain

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

func (h *Handler) Connect(ctx context.Context, req ConnectRequest) (ConnectResponse, *Error) {
	res, err := h.db.CreateRepo(ctx, req)
	if err != nil {
		return res, &Error{
			Code:    "UNKNOWN",
			Message: err.Error(),
		}
	}

	return res, nil
}
