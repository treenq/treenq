// Package auth implements the OAuth2 protocol for authenticating users through Github.
// This package can be used as a reference implementation of an OAuth2 provider for Goth.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/treenq/treenq/src/domain"
	"golang.org/x/oauth2"
)

var (
	AuthURL    = "https://github.com/login/oauth/authorize"
	TokenURL   = "https://github.com/login/oauth/access_token"
	ProfileURL = "https://api.github.com/user"
	EmailURL   = "https://api.github.com/user/emails"
)

// ErrNoVerifiedGitHubPrimaryEmail user doesn't have verified primary email on GitHub
var ErrNoVerifiedGitHubPrimaryEmail = errors.New("the user does not have a verified, primary email address on GitHub")

// New creates a new Github provider, and sets up important connection details.
func New(clientKey, secret, callbackURL string) *GithubOauthProvider {
	return &GithubOauthProvider{
		client:     http.DefaultClient,
		tokenCache: cache.New[string, string](),
		config: &oauth2.Config{
			ClientID:     clientKey,
			ClientSecret: secret,
			RedirectURL:  callbackURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  AuthURL,
				TokenURL: TokenURL,
			},
			Scopes: []string{"profile", "email"},
		},
	}
}

type GithubOauthProvider struct {
	client     *http.Client
	config     *oauth2.Config
	tokenCache *cache.Cache[string, string] // userDisplayName -> token
}

func (p *GithubOauthProvider) AuthUrl(state string) string {
	url := p.config.AuthCodeURL(state)
	return url
}

func (p *GithubOauthProvider) ExchangeUser(ctx context.Context, code string) (domain.UserInfo, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return domain.UserInfo{}, fmt.Errorf("failed to exchange github code to token: %w", err)
	}

	userInfo, err := p.FetchUser(ctx, token.AccessToken)
	if err != nil {
		return domain.UserInfo{}, err
	}

	// Cache the GitHub access token for 15 minutes
	p.tokenCache.Set(userInfo.DisplayName, token.AccessToken, cache.WithExpiration(15*time.Minute))

	return userInfo, nil
}

type githubUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Login string `json:"login"`
}

// FetchUser will go to Github and access basic information about the user.
func (p *GithubOauthProvider) FetchUser(ctx context.Context, token string) (domain.UserInfo, error) {
	user := domain.UserInfo{}
	req, err := http.NewRequestWithContext(ctx, "GET", ProfileURL, nil)
	if err != nil {
		return user, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	response, err := p.client.Do(req)
	if err != nil {
		return user, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return user, fmt.Errorf("GitHub API responded with a %d trying to fetch user information", response.StatusCode)
	}

	var resp githubUser
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		return user, err
	}

	user = domain.UserInfo{
		Email:       resp.Email,
		DisplayName: resp.Login,
	}

	if user.Email == "" {
		for _, scope := range p.config.Scopes {
			if strings.TrimSpace(scope) == "email" || strings.TrimSpace(scope) == "user" || strings.TrimSpace(scope) == "user:email" {
				user.Email, err = p.getPrivateMail(ctx, token)
				if err != nil {
					return user, err
				}
				break
			}
		}
	}

	return user, nil
}

type mail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (p *GithubOauthProvider) getPrivateMail(ctx context.Context, token string) (email string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", EmailURL, nil)
	req.Header.Add("Authorization", "Bearer "+token)
	response, err := p.client.Do(req)
	if err != nil {
		if response != nil {
			response.Body.Close()
		}
		return email, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return email, fmt.Errorf("GitHub API responded with a %d trying to fetch user email", response.StatusCode)
	}

	var mailList []mail
	err = json.NewDecoder(response.Body).Decode(&mailList)
	if err != nil {
		return email, err
	}
	for _, v := range mailList {
		if v.Primary && v.Verified {
			return v.Email, nil
		}
	}
	return email, ErrNoVerifiedGitHubPrimaryEmail
}

// GetUserGithubToken retrieves the cached GitHub token for a user
// Returns domain.ErrUnauthorized if token is expired or not found
func (p *GithubOauthProvider) GetUserGithubToken(userDisplayName string) (string, error) {
	token, ok := p.tokenCache.Get(userDisplayName)
	if !ok {
		return "", domain.ErrUnauthorized
	}
	return token, nil
}
