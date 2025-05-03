package domain

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/treenq/treenq/pkg/vel"
	"github.com/treenq/treenq/pkg/vel/auth"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrRepoNotFound = errors.New("repo not found")
)

type UserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

func (h *Handler) GithubAuthHandler(w http.ResponseWriter, r *http.Request) {
	state := uuid.NewString()
	writeCookie(w, "authstate", state, time.Second*90)
	authUrl := h.oauthProvider.AuthUrl(state)
	http.Redirect(w, r, authUrl, http.StatusTemporaryRedirect)
}

func writeCookie(w http.ResponseWriter, key, value string, exp time.Duration) {
	expiration := time.Now().Add(exp)

	cookie := http.Cookie{Name: key, Value: value, Expires: expiration, HttpOnly: true, SameSite: http.SameSiteStrictMode}
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

	oauthToken, err := h.oauthProvider.ExchangeCode(ctx, req.Code)
	if err != nil {
		return GithubCallbackResponse{}, &vel.Error{
			Message: "failed to exchange code to token",
			Err:     err,
		}
	}

	user, err := h.oauthProvider.FetchUser(r.Context(), oauthToken)
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
	writeCookie(w, auth.AuthKey, token, h.authTtl)
	http.Redirect(w, r, h.authRedirectUrl, http.StatusTemporaryRedirect)
	return GithubCallbackResponse{}, nil
}
