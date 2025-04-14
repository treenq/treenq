package vel

import "net/http"

func NewCorsMiddleware(origin string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			h.ServeHTTP(w, r)
		})
	}
}
