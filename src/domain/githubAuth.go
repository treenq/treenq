package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/treenq/treenq/pkg/vel"
)

func (h *Handler) GithubAuthHandler(w http.ResponseWriter, r *http.Request) {
	state := uuid.New().String()
	email := h.authProfiler.GetProfile(r.Context()).Email
	if err := h.db.SaveAuthState(r.Context(), email, state); err != nil {
		http.Error(w, "Failed to save auth state", http.StatusInternalServerError)
		return
	}
	url := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s&scope=openid+profile+email+repo", h.githubClientID, h.githubRedirectURI, state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    time.Time `json:"expires_in"`
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
	email, err := h.db.AuthStateExists(r.Context(), state)
	if err != nil {
		http.Error(w, "State not found", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	tokenPair, err := h.exchangeCodeForToken(code)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	if err := h.db.SaveTokenPair(r.Context(), email, tokenPair.AccessToken); err != nil {
		http.Error(w, "Failed to save token pair", http.StatusInternalServerError)
		return
	}
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

type LoginRequest struct {
	Provider string `json:"provider"`
}

func (h *Handler) Login(ctx context.Context, req LoginRequest) (struct{}, *vel.Error) {
	authUrl, err := h.authService.Start(ctx, req.Provider)
	if err != nil {
		return ConnectResponse{}, &vel.Error{
			Code:    "GET_AUTH_URL",
			Message: "Failed to get auth url",
		}
	}

	vel.Redirect(ctx, authUrl, http.StatusMovedPermanently)
	return struct{}{}, nil
}

func (h *Handler) HandleSuccessLogin(w http.ResponseWriter, r *http.Request) {
	// http://localhost:3000/authSuccess?id=297291436091508099&token=fD0aXha_zIo3wU7vEr9zREdzDaH2wPQOC0tuyewz_BpVpA&user=283666387590288148
	println()
}
