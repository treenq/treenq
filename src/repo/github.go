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

type GithubClient struct {
	tokenIssuer   TokenIssuer
	appTokenCache *cache.Cache[int, string]
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
	// AppID   int64  `json:"app_id"`
	// AppSlug string `json:"app_slug"`
	// TargetID            int64                `json:"target_id"`
	// TargetType          string               `json:"target_type"`
	// SingleFileName      string               `json:"single_file_name"`
	// RepositorySelection string               `json:"repository_selection"`
	// AccessTokensURL     string               `json:"access_tokens_url"`
	// HTMLURL             string               `json:"html_url"`
	// RepositoriesURL     string               `json:"repositories_url"`
	// Events              []string             `json:"events"`
	// Account             *githubAccount       `json:"account"`
	// Permissions         githubAppPermissions `json:"permissions"`
	// CreatedAt          string               `json:"created_at"`
	// UpdatedAt          string               `json:"updated_at"`
	// SuspendedBy        *githubAccount       `json:"suspended_by"`
	// SuspendedAt        *string              `json:"suspended_at"`
}

// type githubAccount struct {
// 	Name              *string `json:"name"`
// 	Email             *string `json:"email"`
// 	Login             string  `json:"login"`
// 	ID                int64   `json:"id"`
// 	NodeID            string  `json:"node_id"`
// 	AvatarURL         string  `json:"avatar_url"`
// 	GravatarID        string  `json:"gravatar_id"`
// 	URL               string  `json:"url"`
// 	HTMLURL           string  `json:"html_url"`
// 	FollowersURL      string  `json:"followers_url"`
// 	FollowingURL      string  `json:"following_url"`
// 	GistsURL          string  `json:"gists_url"`
// 	StarredURL        string  `json:"starred_url"`
// 	SubscriptionsURL  string  `json:"subscriptions_url"`
// 	OrganizationsURL  string  `json:"organizations_url"`
// 	ReposURL          string  `json:"repos_url"`
// 	EventsURL         string  `json:"events_url"`
// 	ReceivedEventsURL string  `json:"received_events_url"`
// 	Type              string  `json:"type"`
// 	SiteAdmin         bool    `json:"site_admin"`
// }

// type githubAppPermissions struct {
// 	Actions                               string `json:"actions"`
// 	Administration                        string `json:"administration"`
// 	Checks                               string `json:"checks"`
// 	Codespaces                           string `json:"codespaces"`
// 	Contents                             string `json:"contents"`
// 	DependabotSecrets                    string `json:"dependabot_secrets"`
// 	Deployments                          string `json:"deployments"`
// 	Environments                         string `json:"environments"`
// 	Issues                               string `json:"issues"`
// 	Metadata                             string `json:"metadata"`
// 	Packages                             string `json:"packages"`
// 	Pages                                string `json:"pages"`
// 	PullRequests                         string `json:"pull_requests"`
// 	RepositoryCustomProperties           string `json:"repository_custom_properties"`
// 	RepositoryHooks                      string `json:"repository_hooks"`
// 	RepositoryProjects                   string `json:"repository_projects"`
// 	SecretScanningAlerts                 string `json:"secret_scanning_alerts"`
// 	Secrets                              string `json:"secrets"`
// 	SecurityEvents                       string `json:"security_events"`
// 	SingleFile                           string `json:"single_file"`
// 	Statuses                            string `json:"statuses"`
// 	VulnerabilityAlerts                 string `json:"vulnerability_alerts"`
// 	Workflows                           string `json:"workflows"`
// 	Members                             string `json:"members"`
// 	OrganizationAdministration          string `json:"organization_administration"`
// 	OrganizationCustomRoles             string `json:"organization_custom_roles"`
// 	OrganizationCustomOrgRoles          string `json:"organization_custom_org_roles"`
// 	OrganizationCustomProperties        string `json:"organization_custom_properties"`
// 	OrganizationCopilotSeatManagement   string `json:"organization_copilot_seat_management"`
// 	OrganizationAnnouncementBanners     string `json:"organization_announcement_banners"`
// 	OrganizationEvents                  string `json:"organization_events"`
// 	OrganizationHooks                   string `json:"organization_hooks"`
// 	OrganizationPersonalAccessTokens    string `json:"organization_personal_access_tokens"`
// 	OrganizationPersonalAccessTokenRequests string `json:"organization_personal_access_token_requests"`
// 	OrganizationPlan                    string `json:"organization_plan"`
// 	OrganizationProjects                string `json:"organization_projects"`
// 	OrganizationPackages                string `json:"organization_packages"`
// 	OrganizationSecrets                 string `json:"organization_secrets"`
// 	OrganizationSelfHostedRunners       string `json:"organization_self_hosted_runners"`
// 	OrganizationUserBlocking            string `json:"organization_user_blocking"`
// 	TeamDiscussions                     string `json:"team_discussions"`
// 	EmailAddresses                      string `json:"email_addresses"`
// 	Followers                           string `json:"followers"`
// 	GitSSHKeys                          string `json:"git_ssh_keys"`
// 	GPGKeys                             string `json:"gpg_keys"`
// 	InteractionLimits                   string `json:"interaction_limits"`
// 	Profile                             string `json:"profile"`
// 	Starring                            string `json:"starring"`
// }

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
	if c.cachedToken != "" && c.cachedTokenAt.Sub(time.Now()) > time.Minute {
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
