package domain

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rs/xid"
	"github.com/treenq/treenq/pkg/auth"
	"github.com/treenq/treenq/pkg/vel"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrRepoNotFound         = errors.New("repo not found")
	ErrInstallationNotFound = errors.New("installation not found")
	ErrUnauthorized         = errors.New("unauthorized: github token expired or invalid")
)

type UserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

func (h *Handler) GithubAuthHandler(w http.ResponseWriter, r *http.Request) {
	state := h.writeState(w)
	authUrl := h.oauthProvider.AuthUrl(state)
	http.Redirect(w, r, authUrl, http.StatusTemporaryRedirect)
}

func (h *Handler) writeState(w http.ResponseWriter) string {
	state := xid.New().String()
	exp := time.Second * 300
	expiration := time.Now().Add(exp)

	cookie := http.Cookie{Name: "authstate", Value: state, Expires: expiration, MaxAge: int(exp.Seconds()), HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: h.isProd}
	http.SetCookie(w, &cookie)
	return state
}

func (h *Handler) writeToken(w http.ResponseWriter, token string) {
	expiration := time.Now().Add(h.authTtl)

	cookie := http.Cookie{Name: auth.AuthKey, Value: token, Expires: expiration, MaxAge: int(h.authTtl.Seconds()), HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: h.isProd}
	http.SetCookie(w, &cookie)
}

type CodeExchangeRequest struct {
	State string `schema:"state"`
	Code  string `schema:"code"`
}

type GithubCallbackResponse struct{}

// GithubCallbackHandler is the handler for the callback from Github
// It exchanges the code for an access token and returns the given access and refresh tokens
func (h *Handler) GithubCallbackHandler(ctx context.Context, req CodeExchangeRequest) (GithubCallbackResponse, *vel.Error) {
	r := vel.RequestFromContext(ctx)
	w := vel.WriterFromContext(ctx)
	oauthState, _ := r.Cookie("authstate")
	if oauthState == nil {
		return GithubCallbackResponse{}, &vel.Error{
			Code:    "COOKIE_IS_EMPTY",
			Message: "cookie auth status is expected",
		}
	}

	if req.State != oauthState.Value {
		return GithubCallbackResponse{}, &vel.Error{
			Code: "STATE_DOESNT_MATCH",
			Err:  errors.New("state doesn't match"),
		}
	}

	user, err := h.oauthProvider.ExchangeUser(r.Context(), req.Code)
	if err != nil {
		return GithubCallbackResponse{}, &vel.Error{
			Code:    "UNKNOWN",
			Message: "failed to fetch user form oauth provider",
			Err:     err,
		}
	}

	// save user if doesn't exist
	savedUser, err := h.db.GetOrCreateUser(r.Context(), user)
	if err != nil {
		return GithubCallbackResponse{}, &vel.Error{
			Message: "failed get to create user",
			Err:     err,
		}
	}

	token, err := h.jwtIssuer.GenerateJwtToken(map[string]any{
		"id":          savedUser.ID,
		"email":       savedUser.Email,
		"displayName": savedUser.DisplayName,
	})
	if err != nil {
		return GithubCallbackResponse{}, &vel.Error{
			Message: "failed ogenerate jwt token",
			Err:     err,
		}
	}
	h.writeToken(w, token)
	http.Redirect(w, r, h.authRedirectUrl, http.StatusTemporaryRedirect)
	return GithubCallbackResponse{}, nil
}

func (h *Handler) Logout(ctx context.Context, req struct{}) (struct{}, *vel.Error) {
	w := vel.WriterFromContext(ctx)

	cookie := http.Cookie{Name: auth.AuthKey, Value: "", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, &cookie)

	return struct{}{}, nil
}
