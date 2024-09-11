package vel

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"unsafe"
)

type Handler[I, O comparable] func(ctx context.Context, i I) (O, *Error)

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
				err = json.NewEncoder(w).Encode(Error{
					Code:    "FAILED_DECODING",
					Message: err.Error(),
				})
				if err != nil {
					slog.Default().ErrorContext(r.Context(), "failed to write request marshal error", "err", err)
				}
				return
			}
		}

		res, callErr := call(r.Context(), i)
		if callErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(callErr)
			if err != nil {
				slog.Default().ErrorContext(r.Context(), "failed to write api call error", "err", err)
			}
			return
		}

		if hasResBody {
			if err := json.NewEncoder(w).Encode(res); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				err = json.NewEncoder(w).Encode(Error{
					Code:    "FAILED_ENCODING",
					Message: err.Error(),
				})
				if err != nil {
					slog.Default().ErrorContext(r.Context(), "failed to write request marshal error", "err", err)
				}
			}
		}
	}
}

type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware

	handlersMeta []HandlerMeta
}

func (r *Router) Mux() *http.ServeMux {
	return r.mux
}

func (r *Router) Use(m func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, m)
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

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Meta    map[string]string `json:"meta"`
}

func (e *Error) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewRouter() *Router {
	mux := http.NewServeMux()
	mux.Handle("GET /healthz", NewHandler(func(ctx context.Context, _ struct{}) (struct{}, *Error) {
		return struct{}{}, nil
	}))

	return &Router{mux: mux}
}

type Middleware func(http.Handler) http.Handler

func Register[I, O comparable](r *Router, operationID string, handler Handler[I, O], middlewares ...Middleware) {
	var i I
	var o O
	r.handlersMeta = append(r.handlersMeta, HandlerMeta{
		Input:       i,
		Output:      o,
		OperationID: operationID,
	})

	var h http.Handler = NewHandler(handler)
	RegisterHandler(r, "POST /"+operationID, h, middlewares...)
}

func RegisterHandler(r *Router, pattern string, handler http.Handler, middlewares ...Middleware) {
	for i := range r.middlewares {
		handler = r.middlewares[i](handler)
	}
	for i := range middlewares {
		handler = middlewares[i](handler)
	}
	r.mux.Handle(pattern, handler)
}

func RegisterHandlerFunc(r *Router, pattern string, h http.HandlerFunc, middlewares ...Middleware) {
	var handler http.Handler = h
	for i := range r.middlewares {
		handler = r.middlewares[i](handler)
	}
	for i := range middlewares {
		handler = middlewares[i](handler)
	}
	r.mux.Handle(pattern, handler)
}
