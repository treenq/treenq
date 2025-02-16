package repo

import (
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
	tokenIssuer TokenIssuer
	cachee      *cache.Cache[int, string]
	client      *http.Client
}

func NewGithubClient(tokenIssuer TokenIssuer, client *http.Client) *GithubClient {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	newCache := cache.New[int, string]()
	return &GithubClient{
		tokenIssuer: tokenIssuer,
		client:      client,
		cachee:      newCache,
	}
}

var responseBody struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (c *GithubClient) IssueAccessToken(installationID int) (string, error) {
	if token, ok := c.cachee.Get(installationID); ok {
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
	c.cachee.Set(installationID, responseBody.Token, cache.WithExpiration(ttl-time.Second*20))

	return responseBody.Token, nil
}
