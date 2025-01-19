package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

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
		http.Error(w, "Failed to exchange code", http.StatusBadRequest)
	}

	user, err :=	h.oauthProvider.FetchUser(r.Context(), token)
	err != nil {
		http.Error(w, "Failed to fetch user", http.StatusBadRequest)
		return
	}

	// save user if doesn\t exist
	// issue a token
	// TODO: issue a token pair instead: access, refresh, expiry 
	w.WriteHeader(http.StatusOK)
}

type GithubTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (h *Handler) exchangeCodeForToken(code string) (TokenPair, error) {
	urlStr := "https://github.com/login/oauth/access_token"
	q := make(url.Values)
	q.Set("client_id", h.githubClientID)
	q.Set("client_secret", h.githubSecret)
	q.Set("code", code)
	urlStr += "?" + q.Encode()

	req, err := http.NewRequest("POST", urlStr, nil)
	if err != nil {
		return TokenPair{}, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TokenPair{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return TokenPair{}, fmt.Errorf("failed to exchange code for token: %s", resp.Status)
	}

	var result GithubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    time.Now().UTC().Add(time.Duration(result.ExpiresIn) * time.Second).Add(time.Second * -10),
	}, nil
}

type UserInfo struct {
	ID          string
	Email       string
	DisplayName string
}

var ErrUserNotFound = errors.New("user not found")
