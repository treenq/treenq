package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/treenq/treenq/src/domain"
)

type TokenIssuer interface {
	GenerateJwtToken(claims map[string]interface{}) (string, error)
}

type cachedBranches struct {
	Branches []string
	SavedAt  time.Time
}

type GithubClient struct {
	tokenIssuer   TokenIssuer
	appTokenCache *cache.Cache[int, string]
	branchesList  *cache.Cache[string, cachedBranches]
	client        *http.Client

	cachedToken   string
	cachedTokenAt time.Time
}

func NewGithubClient(tokenIssuer TokenIssuer, client *http.Client) *GithubClient {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &GithubClient{
		tokenIssuer:   tokenIssuer,
		client:        client,
		appTokenCache: cache.New[int, string](),
		branchesList:  cache.New[string, cachedBranches](),
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

type githubInstallation struct {
	ID int `json:"id"`
}

// GetUserInstallation gets a user's installation for the authenticated app
// username is the handle for the GitHub user account
func (c *GithubClient) GetUserInstallation(ctx context.Context, displayName string) (int, error) {
	token, err := c.issueJwt()
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.github.com/users/%s/installation", displayName), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	response, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return 0, domain.ErrInstallationNotFound
	}

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("GitHub API responded with %d trying to get user installation for %s", response.StatusCode, displayName)
	}

	var installation githubInstallation
	if err := json.NewDecoder(response.Body).Decode(&installation); err != nil {
		return 0, fmt.Errorf("failed to decode installation response: %w", err)
	}

	return installation.ID, nil
}

func (c *GithubClient) issueJwt() (string, error) {
	if c.cachedToken != "" && time.Until(c.cachedTokenAt) > time.Minute {
		return c.cachedToken, nil
	}

	token, err := c.tokenIssuer.GenerateJwtToken(nil)
	if err != nil {
		return "", fmt.Errorf("failed to issue jwt: %w", err)
	}

	c.cachedToken = token
	c.cachedTokenAt = time.Now()
	return token, nil
}

type githubRepositoriesResponse struct {
	TotalCount   int                          `json:"total_count"`
	Repositories []domain.InstalledRepository `json:"repositories"`
}

func (c *GithubClient) ListRepositories(ctx context.Context, installationID int) ([]domain.Repository, error) {
	token, err := c.IssueAccessToken(installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to issue an installation token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/installation/repositories", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	response, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("installation %d not found", installationID)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API responded with %d trying to list repositories for installation %d", response.StatusCode, installationID)
	}

	var repoResponse githubRepositoriesResponse
	if err := json.NewDecoder(response.Body).Decode(&repoResponse); err != nil {
		return nil, fmt.Errorf("failed to decode repositories response: %w", err)
	}

	return repoResponse.Repositories, nil
}

// GetBranches fetches the list of branch names for a given repo, using cache.
func (c *GithubClient) GetBranches(ctx context.Context, installationID int, owner string, repoName string, fresh bool) ([]string, error) {
	token, err := c.IssueAccessToken(installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to issue an installation token: %w", err)
	}

	cacheKey := owner + repoName
	branches, ok := c.branchesList.Get(cacheKey)
	isFresh := time.Since(branches.SavedAt) < time.Minute
	if ok && ((fresh && isFresh) || !fresh) {
		return branches.Branches, nil
	}
	if token == "" {
		var err error
		token, err = c.IssueAccessToken(installationID)
		if err != nil {
			return nil, err
		}
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches", owner, repoName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API responded with %d: %s", resp.StatusCode, string(respBody))
	}
	var branchesResp []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&branchesResp); err != nil {
		return nil, fmt.Errorf("failed to decode branches response: %w", err)
	}
	branchNames := make([]string, 0, len(branchesResp))
	for _, b := range branchesResp {
		branchNames = append(branchNames, b.Name)
	}
	c.branchesList.Set(cacheKey, cachedBranches{Branches: branchNames, SavedAt: time.Now()}, cache.WithExpiration(10*time.Minute))
	return branchNames, nil
}
