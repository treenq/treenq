package domain

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/treenq/treenq/pkg/vel"
)

type GetBuildProgressRequest struct {
	DeploymentID string `schema:"deploymentID"`
}

type GetBuildProgressResponse struct {
	Message ProgressMessage `json:"message"`
}

var ErrStreamingUnsupported = errors.New("streaming unsupported")

func (h *Handler) GetBuildProgress(ctx context.Context, req GetBuildProgressRequest) (GetBuildProgressResponse, *vel.Error) {
	w := vel.WriterFromContext(ctx)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return GetBuildProgressResponse{}, &vel.Error{
			Err: ErrStreamingUnsupported,
		}
	}

	messages := progress.Get(ctx, req.DeploymentID)

	for {
		select {
		case m, ok := <-messages:
			if !ok {
				return GetBuildProgressResponse{}, nil
			}
			b, err := json.Marshal(GetBuildProgressResponse{Message: m})
			if err != nil {
				return GetBuildProgressResponse{}, &vel.Error{
					Err:     err,
					Message: "failed to write a progress message",
				}
			}
			b = append(b, '\n')
			w.Write(b)
			flusher.Flush()
		case <-ctx.Done():
			return GetBuildProgressResponse{}, nil
		}
	}
}
