package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/digitalocean/godo"
	"github.com/treenq/treenq/pkg/jwt"
	"github.com/treenq/treenq/src/domain"
	"github.com/treenq/treenq/src/repo"
	"github.com/treenq/treenq/src/repo/artifacts"
	"github.com/treenq/treenq/src/repo/extract"
)

type Handler[I, O comparable] func(ctx context.Context, i I) (O, *domain.Error)

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
				err = json.NewEncoder(w).Encode(domain.Error{
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
				err = json.NewEncoder(w).Encode(domain.Error{
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
	mux.Handle("GET /healthz", NewHandler(func(ctx context.Context, _ struct{}) (struct{}, *domain.Error) {
		return struct{}{}, nil
	}))

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

func New(conf Config) (http.Handler, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	store, err := repo.NewStore()
	if err != nil {
		return nil, err
	}

	doClient := godo.NewFromToken(conf.DoToken)
	provider := repo.NewProvider(doClient)
	jwtIssuer := jwt.NewJwtIssuer(conf.JwtKey, conf.JwtSecret, conf.JwtTtl)

	githubClient := repo.NewGithubClient(jwtIssuer, http.DefaultClient)
	gitClient := repo.NewGit()
	docker := artifacts.NewDockerArtifactory("tq-staging")
	extractor := extract.NewExtractor(filepath.Join(wd, "builder"), "cmd/server")

	handlers := domain.NewHandler(store, githubClient, gitClient, extractor, docker, provider)

	router := NewRouter()
	Register(router, "deploy", handlers.Deploy)
	Register(router, "connect", handlers.Connect)
	Register(router, "githubWebhook", handlers.GithubWebhook)
	Register(router, "info", handlers.Info)

	return router.Mux(), nil
}
