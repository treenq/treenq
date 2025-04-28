package domain

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/treenq/treenq/pkg/vel"
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
	state := writeStateCookie(w)
	authUrl := h.oauthProvider.AuthUrl(state)
	http.Redirect(w, r, authUrl, http.StatusTemporaryRedirect)
}

func writeStateCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(time.Second * 90)

	state := uuid.NewString()
	cookie := http.Cookie{Name: "authstate", Value: state, Expires: expiration, HttpOnly: true, SameSite: http.SameSiteStrictMode}
	http.SetCookie(w, &cookie)

	return state
}

func (h *Handler) writeTokenCookie(w http.ResponseWriter, token string) {
	expiration := time.Now().Add(h.authTtl)

	state := uuid.NewString()
	cookie := http.Cookie{Name: "accessToken", Value: state, Expires: expiration, HttpOnly: true, SameSite: http.SameSiteStrictMode}
	http.SetCookie(w, &cookie)
}

type CodeExchangeRequest struct {
	State string `schema:"state"`
	Code  string `schema:"code"`
}

// GithubCallbackHandler is the handler for the callback from Github
// It exchanges the code for an access token and returns the given access and refresh tokens
func (h *Handler) GithubCallbackHandler(ctx context.Context, req CodeExchangeRequest) *vel.Error {
	r := vel.RequestFromContext(ctx)
	w := vel.WriterFromContext(ctx)
	oauthState, _ := r.Cookie("authstate")
	if oauthState == nil {
		return &vel.Error{
			Code:    "COOKIE_IS_EMPTY",
			Message: "cookie auth status is expected",
		}
	}

	if req.State != oauthState.Value {
		return &vel.Error{
			Code: "STATE_DOESNT_MATCH",
			Err:  errors.New("state doesn't match"),
		}
	}

	oauthToken, err := h.oauthProvider.ExchangeCode(ctx, req.Code)
	if err != nil {
		return &vel.Error{
			Message: "failed to exchange code to token",
			Err:     err,
		}
	}

	user, err := h.oauthProvider.FetchUser(r.Context(), oauthToken)
	if err != nil {
		return &vel.Error{
			Code:    "UNKNOWN",
			Message: "failed to fetch user form oauth provider",
			Err:     err,
		}
	}

	// save user if doesn't exist
	savedUser, err := h.db.GetOrCreateUser(r.Context(), user)
	if err != nil {
		return &vel.Error{
			Message: "failed get to create user",
			Err:     err,
		}
	}

	token, err := h.jwtIssuer.GenerateJwtToken(map[string]interface{}{
		"id":          savedUser.ID,
		"email":       savedUser.Email,
		"displayName": savedUser.DisplayName,
	})
	if err != nil {
		return &vel.Error{
			Message: "failed ogenerate jwt token",
			Err:     err,
		}
	}
	h.writeTokenCookie(w, token)
	http.Redirect(w, r, h.authRedirectUrl, http.StatusTemporaryRedirect)
	return nil
}
