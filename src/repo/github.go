package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
)

type TokenIssuer interface {
	GenerateJwtToken(claims map[string]interface{}) (string, error)
}

type GithubClient struct {
	tokenIssuer   TokenIssuer
	appTokenCache *cache.Cache[int, string]
	client        *http.Client
}

func NewGithubClient(tokenIssuer TokenIssuer, client *http.Client) *GithubClient {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &GithubClient{
		tokenIssuer:   tokenIssuer,
		client:        client,
		appTokenCache: cache.New[int, string](),
	}
}

var responseBody struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *GithubClient) IssueAccessToken(installationID int) (string, error) {
	if token, ok := c.appTokenCache.Get(installationID); ok {
		return token, nil
	}
	jwtToken, err := c.tokenIssuer.GenerateJwtToken(nil)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create new request %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.client.Do(req)
	if err != nil || resp == nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to process request: %d, body=%s", resp.StatusCode, string(respBody))
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	ttl := time.Until(responseBody.ExpiresAt)
	c.appTokenCache.Set(installationID, responseBody.Token, cache.WithExpiration(ttl-time.Second*20))

	return responseBody.Token, nil
}

type githubAppResponse struct {
	ClientID    string               `json:"client_id"`
	CreatedAt   string               `json:"created_at"`
	Description string               `json:"description"`
	Events      []string             `json:"events"`
	ExternalURL string               `json:"external_url"`
	HTMLURL     string               `json:"html_url"`
	ID          int64                `json:"id"`
	Name        string               `json:"name"`
	NodeID      string               `json:"node_id"`
	Owner       githubAppOwner       `json:"owner"`
	Permissions githubAppPermissions `json:"permissions"`
	Slug        string               `json:"slug"`
	UpdatedAt   string               `json:"updated_at"`
}
type githubAppOwner struct {
	AvatarURL         string `json:"avatar_url"`
	EventsURL         string `json:"events_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	GravatarID        string `json:"gravatar_id"`
	HTMLURL           string `json:"html_url"`
	ID                int64  `json:"id"`
	Login             string `json:"login"`
	NodeID            string `json:"node_id"`
	OrganizationsURL  string `json:"organizations_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	ReposURL          string `json:"repos_url"`
	SiteAdmin         bool   `json:"site_admin"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	Type              string `json:"type"`
	URL               string `json:"url"`
	UserViewType      string `json:"user_view_type"`
}
type githubAppPermissions struct {
	Contents          string `json:"contents"`
	Emails            string `json:"emails"`
	Metadata          string `json:"metadata"`
	OrganizationHooks string `json:"organization_hooks"`
	PullRequests      string `json:"pull_requests"`
}

func (c *GithubClient) GetApp(ctx context.Context, token, slug string) (githubAppResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/apps/"+slug, nil)
	if err != nil {
		return githubAppResponse{}, err
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	response, err := c.client.Do(req)
	if err != nil {
		return githubAppResponse{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return githubAppResponse{}, fmt.Errorf("GitHub API responded with a %d trying to fetch app information", response.StatusCode)
	}
	var githubApp githubAppResponse
	if err := json.NewDecoder(response.Body).Decode(&githubApp); err != nil {
		return githubApp, fmt.Errorf("failed to marshal github app response: %w", err)
	}

	return githubApp, nil
}
