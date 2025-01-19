package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/treenq/treenq/pkg/vel"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
)

type Interceptor[T authorization.Ctx] struct {
	authorizer *authorization.Authorizer[T]
	l          *slog.Logger
}

func NewInterceptor[T authorization.Ctx](authorizer *authorization.Authorizer[T], l *slog.Logger) *Interceptor[T] {
	return &Interceptor[T]{
		authorizer: authorizer,
		l:          l,
	}
}

func (i *Interceptor[T]) RequireAuthorization(options ...authorization.CheckOption) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, err := i.authorizer.CheckAuthorization(req.Context(), req.Header.Get(authorization.HeaderName), options...)
			if err != nil {
				var e *authorization.UnauthorizedErr
				if errors.As(err, &e) {
					w.WriteHeader(http.StatusUnauthorized)
					if encodeErr := json.NewEncoder(w).Encode(vel.Error{
						Code:    "UNAUTHORIZED",
						Message: err.Error(),
					}); encodeErr != nil {
						i.l.ErrorContext(req.Context(), "failed to encode unauthorized error", "err", encodeErr)
					}
					return
				}
				w.WriteHeader(http.StatusForbidden)
				if encodeErr := json.NewEncoder(w).Encode(vel.Error{
					Code:    "FORBIDDEN",
					Message: err.Error(),
				}); encodeErr != nil {
					i.l.ErrorContext(req.Context(), "failed to encode unauthorized error", "err", encodeErr)
				}
				return
			}
			req = req.WithContext(authorization.WithAuthContext(req.Context(), ctx))
			next.ServeHTTP(w, req)
		})
	}
}

func (i *Interceptor[T]) Context(ctx context.Context) T {
	return authorization.Context[T](ctx)
}
