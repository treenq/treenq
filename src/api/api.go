package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"unsafe"

	"github.com/digitalocean/godo"
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

		body, _ := io.ReadAll(r.Body)
		fmt.Println(string(body))
		// useless comment

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

type Router struct {
	mux *http.ServeMux

	handlersMeta []HandlerMeta
}

func (r *Router) Mux() *http.ServeMux {
	return r.mux
}

func (r *Router) Meta() []HandlerMeta {
	meta := make([]HandlerMeta, len(r.handlersMeta))
	copy(meta, r.handlersMeta)
	return meta
}

type HandlerMeta struct {
	Input       any
	Output      any
	OperationID string
}

func NewRouter() *Router {
	mux := http.NewServeMux()
	mux.Handle("GET /healthz", NewHandler(handlers.Health))

	return &Router{mux: mux}
}

func Register[I, O comparable](r *Router, operationID string, handler Handler[I, O]) {
	var i I
	var o O
	r.handlersMeta = append(r.handlersMeta, HandlerMeta{
		Input:       i,
		Output:      o,
		OperationID: operationID,
	})

	r.mux.Handle("POST /"+operationID, NewHandler(handler))
}

func New() (http.Handler, error) {
	store, err := repo.NewStore()
	if err != nil {
		return nil, err
	}

	doToken := "fake-token"

	doClient := godo.NewFromToken(doToken)
	provider := repo.NewProvider(doClient)
	_ = provider

	router := NewRouter()
	Register(router, "deploy", handlers.Deploy)
	Register(router, "connect", handlers.NewConnect(store))
	Register(router, "githubWebhook", handlers.GithubWebhook)
	Register(router, "info", handlers.Info)

	meta := router.Meta()
	_ = meta

	return router.Mux(), nil
}
