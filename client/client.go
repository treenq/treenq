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
	After               string             `json:"after"`
	Installation        Installation       `json:"installation"`
	Sender              Sender             `json:"sender"`
	Action              string             `json:"action"`
	Repositories        []GithubRepository `json:"repositories"`
	RepositoriesAdded   []GithubRepository `json:"repositories_added"`
	RepositoriesRemoved []GithubRepository `json:"repositories_removed"`
	Ref                 string             `json:"ref"`
	Repository          GithubRepository   `json:"repository"`
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

type GithubRepository struct {
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
	Installation bool               `json:"installation"`
	Repos        []GithubRepository `json:"repos"`
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
	Repo GithubRepository `json:"repo"`
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
	RepoID           string `json:"repoID"`
	FromDeploymentID string `json:"fromDeploymentID"`
}

type GetDeploymentResponse struct {
	Deployment AppDeployment `json:"deployment"`
}

type AppDeployment struct {
	ID               string    `json:"id"`
	FromDeploymentID string    `json:"fromDeploymentID"`
	RepoID           string    `json:"repoID"`
	Space            Space     `json:"space"`
	Sha              string    `json:"sha"`
	Branch           string    `json:"branch"`
	CommitMessage    string    `json:"commitMessage"`
	BuildTag         string    `json:"buildTag"`
	UserDisplayName  string    `json:"userDisplayName"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Status           string    `json:"status"`
}

type Space struct {
	Version string
	Service Service
}

type Service struct {
	Name                string              `json:"name"`
	DockerfilePath      string              `json:"dockerfilePath"`
	DockerContext       string              `json:"dockerContext"`
	RuntimeEnvs         map[string]string   `json:"runtimeEnvs"`
	HttpPort            int                 `json:"httpPort"`
	Replicas            int                 `json:"replicas"`
	ComputationResource ComputationResource `json:"computationResource"`
}

type ComputationResource struct {
	CpuUnits   int `json:"cpuUnits"`
	MemoryMibs int `json:"memoryMibs"`
	DiskGibs   int `json:"diskGibs"`
}

func (c *Client) Deploy(ctx context.Context, req DeployRequest) (GetDeploymentResponse, error) {
	var res GetDeploymentResponse

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
	Payload    string        `json:"payload"`
	Level      slog.Level    `json:"level"`
	Final      bool          `json:"final"`
	Timestamp  time.Time     `json:"timestamp"`
	Deployment AppDeployment `json:"deployment,omitzero"`
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

type GetDeploymentsRequest struct {
	RepoID string `json:"repoID"`
}

type GetDeploymentsResponse struct {
	Deployments []AppDeployment `json:"deployments"`
}

func (c *Client) GetDeployments(ctx context.Context, req GetDeploymentsRequest) (GetDeploymentsResponse, error) {
	var res GetDeploymentsResponse

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/getDeployments", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call getDeployments: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode getDeployments response: %w", err)
	}

	return res, nil
}

type SetSecretRequest struct {
	RepoID string `json:"repoID"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

func (c *Client) SetSecret(ctx context.Context, req SetSecretRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/setSecret", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return fmt.Errorf("failed to call setSecret: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return err
	}

	return nil
}

type GetSecretsRequest struct {
	RepoID string `json:"repoID"`
}

type GetSecretsResponse struct {
	Keys []string `json:"keys"`
}

func (c *Client) GetSecrets(ctx context.Context, req GetSecretsRequest) (GetSecretsResponse, error) {
	var res GetSecretsResponse

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/getSecrets", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call getSecrets: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode getSecrets response: %w", err)
	}

	return res, nil
}

type RevealSecretRequest struct {
	RepoID string `json:"repoID"`
	Key    string `json:"key"`
}

type RevealSecretResponse struct {
	Value string `json:"value"`
}

func (c *Client) RevealSecret(ctx context.Context, req RevealSecretRequest) (RevealSecretResponse, error) {
	var res RevealSecretResponse

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/revealSecret", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call revealSecret: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode revealSecret response: %w", err)
	}

	return res, nil
}

type RemoveSecretRequest struct {
	RepoID string `json:"repoID"`
	Key    string `json:"key"`
}

func (c *Client) RemoveSecret(ctx context.Context, req RemoveSecretRequest) error {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/removeSecret", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return fmt.Errorf("failed to call removeSecret: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return err
	}

	return nil
}
