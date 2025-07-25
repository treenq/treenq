package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"
	"sigs.k8s.io/yaml"
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
	ID      int                       `json:"id"`
	Account githubInstallationAccount `json:"account"`
}

type githubInstallationAccount struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Type  string `json:"type"`
}

type githubRepositoriesResponse struct {
	TotalCount   int                          `json:"total_count"`
	Repositories []domain.InstalledRepository `json:"repositories"`
}

// ListAllRepositoriesForUser fetches repositories from all installations accessible to a user
func (c *GithubClient) ListAllRepositoriesForUser(ctx context.Context, userGithubToken string) (map[int][]domain.GithubRepository, error) {
	accessibleInstallations, err := c.GetUserAccessibleInstallations(ctx, userGithubToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user accessible installations: %w", err)
	}

	if len(accessibleInstallations) == 0 {
		return nil, fmt.Errorf("no accessible installations found for user")
	}

	return c.ListAllRepositoriesForInstallations(ctx, accessibleInstallations)
}

// GetUserAccessibleInstallations gets installations accessible to a specific user using their GitHub access token
// This uses the GitHub API endpoint /user/installations which requires a user access token
func (c *GithubClient) GetUserAccessibleInstallations(ctx context.Context, userGithubToken string) ([]int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/installations", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+userGithubToken)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	response, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized: user token may be invalid or expired")
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API responded with %d trying to get user installations", response.StatusCode)
	}

	var userInstallationsResponse struct {
		TotalCount    int                  `json:"total_count"`
		Installations []githubInstallation `json:"installations"`
	}

	if err := json.NewDecoder(response.Body).Decode(&userInstallationsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode user installations response: %w", err)
	}

	installationIDs := make([]int, len(userInstallationsResponse.Installations))
	for i, installation := range userInstallationsResponse.Installations {
		installationIDs[i] = installation.ID
	}

	return installationIDs, nil
}

// ListAllRepositoriesForInstallations fetches repositories for all provided installation IDs
func (c *GithubClient) ListAllRepositoriesForInstallations(ctx context.Context, installationIDs []int) (map[int][]domain.GithubRepository, error) {
	result := make(map[int][]domain.GithubRepository)

	for _, installationID := range installationIDs {
		repos, err := c.ListRepositories(ctx, installationID)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for installation %d: %w", installationID, err)
		}
		result[installationID] = repos
	}

	return result, nil
}

func (c *GithubClient) ListRepositories(ctx context.Context, installationID int) ([]domain.GithubRepository, error) {
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

	// Set the installation ID for each repository
	for i := range repoResponse.Repositories {
		repoResponse.Repositories[i].InstallationID = installationID
	}

	return repoResponse.Repositories, nil
}

// GetBranches fetches the list of branch names for a given repo, using cache.
func (c *GithubClient) GetBranches(ctx context.Context, installationID int, repoFullName string, fresh bool) ([]string, error) {
	token, err := c.IssueAccessToken(installationID)
	if err != nil {
		return nil, fmt.Errorf("failed to issue an installation token: %w", err)
	}

	branches, ok := c.branchesList.Get(repoFullName)
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
	url := fmt.Sprintf("https://api.github.com/repos/%s/branches", repoFullName)
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
	c.branchesList.Set(repoFullName, cachedBranches{Branches: branchNames, SavedAt: time.Now()}, cache.WithExpiration(10*time.Minute))
	return branchNames, nil
}

type githubFileContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Content     []byte `json:"content"`
	Encoding    string `json:"encoding"`
}

// GetFileContent retrieves the content of a specific file from a GitHub repository
func (c *GithubClient) GetRepoSpace(ctx context.Context, installationID int, repoFullName, ref string) (tqsdk.Space, error) {
	token, err := c.IssueAccessToken(installationID)
	if err != nil {
		return tqsdk.Space{}, fmt.Errorf("failed to issue an installation token: %w", err)
	}

	// Try tq.json first
	space, err := c.getRepoSpaceFromFile(ctx, token, repoFullName, ref, "tq.json", true)
	if err == nil {
		return space, nil
	}

	// If tq.json not found, try tq.yaml
	if err == domain.ErrNoTqJsonFound || err == domain.ErrSpaceNotFound {
		space, yamlErr := c.getRepoSpaceFromFile(ctx, token, repoFullName, ref, "tq.yaml", false)
		if yamlErr == nil {
			return space, nil
		}
		// Return original JSON error if YAML also fails
		return tqsdk.Space{}, err
	}

	return tqsdk.Space{}, err
}

func (c *GithubClient) getRepoSpaceFromFile(ctx context.Context, token, repoFullName, ref, filename string, isJSON bool) (tqsdk.Space, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", repoFullName, filename)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return tqsdk.Space{}, fmt.Errorf("failed to create request: %w", err)
	}

	if ref != "" {
		q := req.URL.Query()
		q.Add("ref", ref)
		req.URL.RawQuery = q.Encode()
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := c.client.Do(req)
	if err != nil {
		return tqsdk.Space{}, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return tqsdk.Space{}, domain.ErrNoTqJsonFound
		}
		respBody, _ := io.ReadAll(resp.Body)
		return tqsdk.Space{}, fmt.Errorf("GitHub API responded with %d: %s", resp.StatusCode, string(respBody))
	}

	var fileContent githubFileContent
	if err := json.NewDecoder(resp.Body).Decode(&fileContent); err != nil {
		return tqsdk.Space{}, fmt.Errorf("failed to decode file content response: %w", err)
	}

	var space tqsdk.Space
	if isJSON {
		if err := json.Unmarshal(fileContent.Content, &space); err != nil {
			return tqsdk.Space{}, domain.ErrTqIsNotValidJson
		}
	} else {
		// Use yaml.Unmarshal from sigs.k8s.io/yaml which handles both JSON and YAML
		if err := yaml.Unmarshal(fileContent.Content, &space); err != nil {
			return tqsdk.Space{}, domain.ErrTqIsNotValidJson
		}
	}

	if err := space.Validate(); err != nil {
		return tqsdk.Space{}, err
	}

	return space, nil
}
