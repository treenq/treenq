package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"unsafe"

	"github.com/treenq/treenq/src/handlers"
)

func NewHandler[I, O comparable](call func(ctx context.Context, i I) (O, *handlers.Error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var i I

		if unsafe.Sizeof(i) != 0 {
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
			w.WriteHeader(callErr.Status)
			json.NewEncoder(w).Encode(callErr)
			return
		}

		if unsafe.Sizeof(res) != 0 {
			if err := json.NewEncoder(w).Encode(res); err != nil {
				json.NewEncoder(w).Encode(handlers.Error{
					Code:    "FAILED_ENCODING",
					Message: err.Error(),
				})
			}
		}
	}
}

type GetReq struct {
	ID int
}

type GetRes struct {
	Value string
	Time  time.Time
}

func New() http.Handler {

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", NewHandler(handlers.Health))

	mux.Handle("POST /deploy", NewHandler(handlers.Deploy))

	return mux
}
