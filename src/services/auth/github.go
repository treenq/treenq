// Package github implements the OAuth2 protocol for authenticating users through Github.
// This package can be used as a reference implementation of an OAuth2 provider for Goth.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

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
var ErrNoVerifiedGitHubPrimaryEmail = errors.New("The user does not have a verified, primary email address on GitHub")

// New creates a new Github provider, and sets up important connection details.
func New(clientKey, secret, callbackURL string, scopes ...string) *GithubOauthProvider {
	return &GithubOauthProvider{
		client: http.DefaultClient,
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
	client *http.Client
	config *oauth2.Config
}

func (p *GithubOauthProvider) AuthUrl(state string) string {
	url := p.config.AuthCodeURL(state)
	return url
}

func (p *GithubOauthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange github code to token", err)
	}

	return token.AccessToken, nil
}

type githubUser struct {
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

	err = json.NewDecoder(response.Body).Decode(&user)
	if err != nil {
		return user, err
	}

	if user.Email == "" {
		for _, scope := range p.config.Scopes {
			if strings.TrimSpace(scope) == "user" || strings.TrimSpace(scope) == "user:email" {
				user.Email, err = p.getPrivateMail(ctx, token)
				if err != nil {
					return user, err
				}
				break
			}
		}
	}
	return user, err
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
