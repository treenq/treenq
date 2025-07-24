package treenq

import (
	"net/http"
	"strings"
)

const WorkspaceHeader = "t-workspace"

var allowedHeaders = []string{
	"Content-Type",
	"Accept",
	"Authorization",
	"X-CSRF-Token",
	WorkspaceHeader,
}

func NewCorsMiddleware(origin string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ","))
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
