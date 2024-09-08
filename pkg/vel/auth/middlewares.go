package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	logto "github.com/logto-io/go/client"
	"github.com/logto-io/go/core"
)

type sessionKey struct{}

func SessionToContext(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionKey{}, session)
}

func SessionFromContext(ctx context.Context) *Session {
	v := ctx.Value(sessionKey{})
	if v == nil {
		return nil
	}

	return v.(*Session)
}

type claimsKey struct{}

func ClaimsToContext(ctx context.Context, claims core.IdTokenClaims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

func ClaimsFromContext(ctx context.Context) core.IdTokenClaims {
	v := ctx.Value(claimsKey{})
	if v == nil {
		return core.IdTokenClaims{}
	}

	return v.(core.IdTokenClaims)
}

func NewSessionMiddleware(store sessions.Store, l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := NewSession(store, r, w, l)
			ctx := SessionToContext(r.Context(), session)
			*r = *r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func NewRequiresAuthMiddleware(conf AuthConfig, l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := SessionFromContext(r.Context())
			client := logto.NewLogtoClient(conf.newLogtoConfig(), &SessionStore{session})

			claims, err := client.GetIdTokenClaims()
			if err != nil {
				writeErr(r, w, ErrUnauthorized, l)
				return
			}

			ctx := ClaimsToContext(r.Context(), claims)
			*r = *r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
