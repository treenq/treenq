package repo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TokenIssuer interface {
	GeneratedJwtToken() (string, error)
}

type GithubClient struct {
	tokenIssuer TokenIssuer
	client      *http.Client
}

func NewGithubClient(tokenIssuer TokenIssuer, client *http.Client) *GithubClient {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &GithubClient{
		tokenIssuer: tokenIssuer,
		client:      client,
	}
}

var responseBody struct {
	Token string `json:"token"`
}

func (c *GithubClient) IssueAccessToken(installationID int) (string, error) {
	jwtToken, err := c.tokenIssuer.GeneratedJwtToken()
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
	if err != nil {
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

	return responseBody.Token, nil
}
