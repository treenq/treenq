package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/digitalocean/godo"
	"github.com/treenq/treenq/pkg/artifacts"
	"github.com/treenq/treenq/pkg/extract"
	"github.com/treenq/treenq/pkg/jwt"
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
	mux.Handle("GET /healthz", NewHandler(func(ctx context.Context, _ struct{}) (struct{}, *handlers.Error) {
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
	extractor := extract.NewExtractor(filepath.Join(wd, "builder"))

	handlers := handlers.NewHandler(store, githubClient, gitClient, extractor, docker, provider)

	router := NewRouter()
	Register(router, "deploy", handlers.Deploy)
	Register(router, "connect", handlers.Connect)
	Register(router, "githubWebhook", handlers.GithubWebhook)
	Register(router, "info", handlers.Info)

	return router.Mux(), nil
}
