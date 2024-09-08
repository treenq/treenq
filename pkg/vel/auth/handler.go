package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/sessions"
	logto "github.com/logto-io/go/client"
	"github.com/treenq/treenq/pkg/vel"
)

type AuthConfig struct {
	ID                 string
	Secret             string
	Endpoint           string
	Resource           string
	SignInRedirectUri  string
	SignOutRedirectUri string
}

func (c AuthConfig) newLogtoConfig() *logto.LogtoConfig {
	return &logto.LogtoConfig{
		Endpoint:  c.Endpoint,
		AppId:     c.ID,
		AppSecret: c.Secret,
		Resources: []string{c.Resource},
		Scopes:    []string{"openid", "profile", "offline_access", "email"},
	}
}

type AuthHandler struct {
	conf  AuthConfig
	store sessions.Store
	l     *slog.Logger

	SignIn          http.Handler
	SignInCallback  http.Handler
	SignOut         http.Handler
	SignOutCallback http.Handler
}

func NewAuthHandler(conf AuthConfig, store sessions.Store, l *slog.Logger) *AuthHandler {
	h := &AuthHandler{
		conf:  conf,
		store: store,
		l:     l,
	}
	h.SignIn = NewSignInHandler(h)
	h.SignInCallback = NewSignInCallbackHandler(h)
	h.SignOut = NewSignOutHandler(h)
	h.SignOutCallback = NewSignOutCallbackHandler(h)

	return h
}

func (h *AuthHandler) Use(m func(http.Handler) http.Handler) {
	h.SignIn = m(h.SignIn)
	h.SignInCallback = m(h.SignInCallback)
	h.SignOut = m(h.SignOut)
	h.SignOutCallback = m(h.SignOutCallback)
}

func NewSignInHandler(h *AuthHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := SessionFromContext(r.Context())
		client := logto.NewLogtoClient(h.conf.newLogtoConfig(), &SessionStore{session})
		uri, err := client.SignIn(h.conf.SignInRedirectUri)
		if err != nil {
			writeErr(r, w, ErrSignIn, h.l)
			return
		}

		http.Redirect(w, r, uri, http.StatusTemporaryRedirect)
	})
}

func NewSignInCallbackHandler(h *AuthHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := SessionFromContext(r.Context())
		client := logto.NewLogtoClient(h.conf.newLogtoConfig(), &SessionStore{session})
		if err := client.HandleSignInCallback(r); err != nil {
			writeErr(r, w, ErrSignInCallback, h.l)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

func NewSignOutHandler(h *AuthHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := SessionFromContext(r.Context())
		client := logto.NewLogtoClient(h.conf.newLogtoConfig(), &SessionStore{session})
		uri, err := client.SignOut(h.conf.SignOutRedirectUri)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			h.l.ErrorContext(r.Context(), "failed to sign out", "err", err)
		}

		http.Redirect(w, r, uri, http.StatusTemporaryRedirect)
	})
}

func NewSignOutCallbackHandler(h *AuthHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func writeErr(r *http.Request, w http.ResponseWriter, e vel.Error, l *slog.Logger) {
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(e); err != nil {
		l.ErrorContext(r.Context(), "failed to write auth error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
