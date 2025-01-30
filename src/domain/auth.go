package domain

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type UserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

func (h *Handler) GithubAuthHandler(w http.ResponseWriter, r *http.Request) {
	state := generateStateOauthCookie(w)
	authUrl := h.oauthProvider.AuthUrl(state)
	http.Redirect(w, r, authUrl, http.StatusTemporaryRedirect)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(time.Second * 90)

	state := uuid.NewString()
	cookie := http.Cookie{Name: "authstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    time.Time `json:"expires_in"`
}

// GithubCallbackHandler is the handler for the callback from Github
// It exchanges the code for an access token and returns the given access and refresh tokens
func (h *Handler) GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	oauthState, _ := r.Cookie("authstate")

	if r.URL.Query().Get("state") != oauthState.Value {
		log.Println("invalid auth state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	token, err := h.oauthProvider.ExchangeCode(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	user, err := h.oauthProvider.FetchUser(r.Context(), token)
	if err != nil {
		http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
		return
	}

	// save user if doesn't exist
	savedUser, err := h.db.GetOrCreateUser(r.Context(), user)
	if err != nil {
		http.Error(w, "Failed to get or create user", http.StatusInternalServerError)
		return
	}

	tokens, err := h.jwtIssuer.GenerateJwtToken(map[string]interface{}{
		"id":          savedUser.ID,
		"email":       savedUser.Email,
		"displayName": savedUser.DisplayName,
	})
	if err != nil {
		http.Error(w, "failed to issue jwt token", http.StatusInternalServerError)
	}
	// issue a token
	// TODO: issue a token pair instead: access, refresh, expiry
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tokens); err != nil {
		http.Error(w, "Failed to write tokens to response", http.StatusInternalServerError)
		return
	}
}
