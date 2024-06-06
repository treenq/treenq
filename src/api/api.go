package api

import (
	"context"
	"encoding/json"
	"net/http"
	"unsafe"

	"github.com/treenq/treenq/src/handlers"
	"github.com/treenq/treenq/src/repo"
)

type Handler[I, O comparable] func(ctx context.Context, i I) (O, *handlers.Error)

func NewHandler[I, O comparable](call Handler[I, O]) http.HandlerFunc {
	var iType I
	var oType O
	hasReqBody := unsafe.Sizeof(iType) != 0
	hasResBody := unsafe.Sizeof(oType) != 0

	return func(w http.ResponseWriter, r *http.Request) {
		var i I

		if hasReqBody {
			if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(handlers.Error{
					Code:    "FAILED_DECODING",
					Message: err.Error(),
				})
				return
			}
		}

		res, callErr := call(r.Context(), i)
		if callErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(callErr)
			return
		}

		if hasResBody {
			if err := json.NewEncoder(w).Encode(res); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(handlers.Error{
					Code:    "FAILED_ENCODING",
					Message: err.Error(),
				})
			}
		}
	}
}

func New() (http.Handler, error) {
	store, err := repo.NewStore()
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", NewHandler(handlers.Health))

	mux.Handle("POST /deploy", NewHandler(handlers.Deploy))
	mux.Handle("POST /connect", NewHandler(handlers.NewConnect(store)))

	mux.Handle("POST /githubWebhook", NewHandler(handlers.GithubWebhook))

	return mux, nil
}
