package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	client *http.Client

	baseUrl string
	headers http.Header
}

func NewClient(baseUrl string, client *http.Client, headers map[string]string) *Client {
	h := make(http.Header)
	for k, v := range headers {
		h.Set(k, v)
	}
	return &Client{
		client:  client,
		baseUrl: baseUrl,
		headers: h,
	}
}

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Meta    map[string]string `json:"meta"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s, %s", e.Code, e.Message)
}

func HandleErr(resp *http.Response) error {
	if resp.StatusCode >= 500 {
		var errResp Error
		err := json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			return &Error{
				Code:    "UNKNOWN",
				Message: "failed to decode error response: " + err.Error(),
			}
		}
		return &errResp
	}

	if resp.StatusCode >= 400 {
		var errResp Error
		err := json.NewDecoder(resp.Body).Decode(&errResp)
		if err != nil {
			return &Error{
				Code:    "UNKNOWN",
				Message: "failed to decode error response: " + err.Error(),
			}
		}
		return &errResp
	}

	return nil
}

func (c *Client) Auth(ctx context.Context) error {
	q := make(url.Values)

	r, err := http.NewRequest("GET", c.baseUrl+"/auth?"+q.Encode(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return fmt.Errorf("failed to call auth: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return err
	}

	return nil
}

type CodeExchangeRequest struct {
	State string
	Code  string
}

func (c *Client) AuthCallback(ctx context.Context, req CodeExchangeRequest) error {
	q := make(url.Values)
	q.Set("state", req.State)
	q.Set("code", req.Code)

	r, err := http.NewRequest("GET", c.baseUrl+"/authCallback?"+q.Encode(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return fmt.Errorf("failed to call authCallback: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return err
	}

	return nil
}

type GithubWebhookRequest struct {
	After               string       `json:"after"`
	Installation        Installation `json:"installation"`
	Sender              Sender       `json:"sender"`
	Action              string       `json:"action"`
	Repositories        []Repository `json:"repositories"`
	RepositoriesAdded   []Repository `json:"repositories_added"`
	RepositoriesRemoved []Repository `json:"repositories_removed"`
	Ref                 string       `json:"ref"`
	Repository          Repository   `json:"repository"`
}

type Installation struct {
	ID      int                 `json:"id"`
	Account InstallationAccount `json:"account"`
}

type InstallationAccount struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Login string `json:"login"`
}

type Sender struct {
	Login string `json:"login"`
}

type Repository struct {
	ID             int    `json:"id"`
	FullName       string `json:"full_name"`
	Private        bool   `json:"private"`
	Branch         string `json:"branch"`
	InstallationID int    `json:"installationID"`
	TreenqID       string `json:"treenqID"`
	Status         string `json:"status"`
}

func (c *Client) GithubWebhook(ctx context.Context, req GithubWebhookRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/githubWebhook", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return fmt.Errorf("failed to call githubWebhook: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Logout(ctx context.Context) error {
	body := bytes.NewBuffer(nil)

	r, err := http.NewRequest("POST", c.baseUrl+"/logout", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return fmt.Errorf("failed to call logout: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return err
	}

	return nil
}

type InfoResponse struct {
	Version string `json:"version"`
}

func (c *Client) Info(ctx context.Context) (InfoResponse, error) {
	var res InfoResponse

	body := bytes.NewBuffer(nil)

	r, err := http.NewRequest("POST", c.baseUrl+"/info", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call info: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode info response: %w", err)
	}

	return res, nil
}

type GetProfileResponse struct {
	UserInfo UserInfo `json:"userInfo"`
}

type UserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

func (c *Client) GetProfile(ctx context.Context) (GetProfileResponse, error) {
	var res GetProfileResponse

	body := bytes.NewBuffer(nil)

	r, err := http.NewRequest("POST", c.baseUrl+"/getProfile", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call getProfile: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode getProfile response: %w", err)
	}

	return res, nil
}

type GetReposResponse struct {
	Installation string       `json:"installationID"`
	Repos        []Repository `json:"repos"`
}

func (c *Client) GetRepos(ctx context.Context) (GetReposResponse, error) {
	var res GetReposResponse

	body := bytes.NewBuffer(nil)

	r, err := http.NewRequest("POST", c.baseUrl+"/getRepos", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call getRepos: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode getRepos response: %w", err)
	}

	return res, nil
}

type GetBranchesRequest struct {
	RepoName string `json:"repoName"`
}

type GetBranchesResopnse struct {
	Branches []string `json:"branches"`
}

func (c *Client) GetBranches(ctx context.Context, req GetBranchesRequest) (GetBranchesResopnse, error) {
	var res GetBranchesResopnse

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/getBranches", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call getBranches: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode getBranches response: %w", err)
	}

	return res, nil
}

func (c *Client) SyncGithubApp(ctx context.Context) (GetReposResponse, error) {
	var res GetReposResponse

	body := bytes.NewBuffer(nil)

	r, err := http.NewRequest("POST", c.baseUrl+"/syncGithubApp", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call syncGithubApp: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode syncGithubApp response: %w", err)
	}

	return res, nil
}

type ConnectBranchRequest struct {
	RepoID string `json:"repoID"`
	Branch string `json:"branch"`
}

type ConnectBranchResponse struct {
	Repo Repository `json:"repo"`
}

func (c *Client) ConnectRepoBranch(ctx context.Context, req ConnectBranchRequest) (ConnectBranchResponse, error) {
	var res ConnectBranchResponse

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/connectRepoBranch", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call connectRepoBranch: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode connectRepoBranch response: %w", err)
	}

	return res, nil
}

type DeployRequest struct {
	RepoID string `json:"repoID"`
}

type DeployResponse struct {
	DeploymentID string    `json:"deploymentID"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (c *Client) Deploy(ctx context.Context, req DeployRequest) (DeployResponse, error) {
	var res DeployResponse

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/deploy", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call deploy: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode deploy response: %w", err)
	}

	return res, nil
}

type GetDeploymentRequest struct {
	DeploymentID string `json:"deploymentID"`
}

type GetDeploymentResponse struct {
	Deployment AppDeployment `json:"deployment"`
}

type AppDeployment struct {
	ID              string    `json:"id"`
	RepoID          string    `json:"repoID"`
	Space           Space     `json:"space"`
	Sha             string    `json:"sha"`
	BuildTag        string    `json:"buildTag"`
	UserDisplayName string    `json:"userDisplayName"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Status          string    `json:"status"`
}

type Space struct {
	Key     string
	Region  string
	Service Service
}

type Service struct {
	Key                 string
	DockerfilePath      string
	BuildEnvs           map[string]string
	RuntimeEnvs         map[string]string
	BuildSecrets        []string
	RuntimeSecrets      []string
	HttpPort            int
	Replicas            int
	Name                string
	SizeSlug            string
	ComputationResource ComputationResource
}

type ComputationResource struct {
	CpuUnits   int
	MemoryMibs int
	DiskGibs   int
}

func (c *Client) GetDeployment(ctx context.Context, req GetDeploymentRequest) (GetDeploymentResponse, error) {
	var res GetDeploymentResponse

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/getDeployment", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call getDeployment: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode getDeployment response: %w", err)
	}

	return res, nil
}

type GetBuildProgressRequest struct {
	DeploymentID string
}

type GetBuildProgressResponse struct {
	Message ProgressMessage `json:"message"`
}

type ProgressMessage struct {
	Payload string     `json:"payload"`
	Level   slog.Level `json:"level"`
	Final   bool       `json:"-"`
}

func (c *Client) GetBuildProgress(ctx context.Context, req GetBuildProgressRequest) (GetBuildProgressResponse, error) {
	var res GetBuildProgressResponse

	q := make(url.Values)
	q.Set("deploymentID", req.DeploymentID)

	r, err := http.NewRequest("GET", c.baseUrl+"/getBuildProgress?"+q.Encode(), nil)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call getBuildProgress: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode getBuildProgress response: %w", err)
	}

	return res, nil
}
