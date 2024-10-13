package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (h *Handler) GithubAuthHandler(w http.ResponseWriter, r *http.Request) {
	state := uuid.New().String()
	states[state] = h.authProfiler.GetProfile(r.Context()).Email
	url := fmt.Sprintf(`https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s&scope=repo`, h.githubClientID, h.githubRedirectURI, state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

var states = make(map[string]string)
var tokens = make(map[string]TokenPair)
var repos = make(map[string][]GithubRepository)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Time
}

// GithubCallbackHandler is the handler for the callback from Github
// It exchanges the code for an access token and returns the given access and refresh tokens
func (h *Handler) GithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "State not found", http.StatusBadRequest)
		return
	}
	email, ok := states[state]
	if !ok {
		http.Error(w, "State not found", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	tokenPair, err := h.exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	tokens[email] = tokenPair
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) exchangeCodeForToken(code string) (TokenPair, error) {
	url := "https://github.com/login/oauth/access_token"
	data := map[string]string{
		"client_id":     h.githubClientID,
		"client_secret": h.githubSecret,
		"code":          code,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return TokenPair{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return TokenPair{}, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TokenPair{}, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return TokenPair{}, err
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return TokenPair{}, fmt.Errorf("access token not found in response")
	}
	refreshToken, ok := result["refresh_token"].(string)
	if !ok {
		return TokenPair{}, fmt.Errorf("refresh token not found in response")
	}
	expiresIn, ok := result["expires_in"].(int)
	if !ok {
		return TokenPair{}, fmt.Errorf("expires in not found in response")
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    time.Now().Add(time.Duration(expiresIn) * time.Second),
	}, nil
}
