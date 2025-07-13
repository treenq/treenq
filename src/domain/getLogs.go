package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dennypenta/vel"
)

type GetLogsRequest struct {
	RepoID string `schema:"repoID"`
}

type GetLogsResponse struct {
	Message ProgressMessage `json:"message"`
}

var ErrNoPodsRunning = errors.New("not pods running")

func (h *Handler) GetLogs(ctx context.Context, req GetLogsRequest) (GetLogsResponse, *vel.Error) {
	w := vel.WriterFromContext(ctx)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return GetLogsResponse{}, &vel.Error{
			Err: ErrStreamingUnsupported,
		}
	}

	profile, rpcErr := h.GetProfile(ctx, struct{}{})
	if rpcErr != nil {
		return GetLogsResponse{}, rpcErr
	}

	logChan := make(chan ProgressMessage, 100)

	go func() {
		defer close(logChan)
		err := h.kube.StreamLogs(ctx, h.kubeConfig, req.RepoID, profile.UserInfo.DisplayName, logChan)
		if errors.Is(err, ErrNoPodsRunning) {
			logChan <- ProgressMessage{
				ErrorCode: "NO_PODS_RUNNING",
			}
			return
		}

		if err != nil {
			logChan <- ProgressMessage{
				Payload: fmt.Sprintf("Error streaming logs: %v", err),
				Level:   slog.LevelError,
				Final:   true,
			}
		}
	}()

	for {
		select {
		case msg, ok := <-logChan:
			if !ok {
				return GetLogsResponse{}, nil
			}

			b, err := json.Marshal(GetLogsResponse{Message: msg})
			if err != nil {
				return GetLogsResponse{}, &vel.Error{
					Err:     err,
					Message: "failed to write a log message",
				}
			}

			fmt.Fprintf(w, "data: %s\n\n", b)
			flusher.Flush()

			if msg.Final {
				return GetLogsResponse{}, nil
			}
		case <-ctx.Done():
			return GetLogsResponse{}, nil
		}
	}
}
