package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	client *http.Client

	baseUrl string
}

func NewClient(baseUrl string, client *http.Client) *Client {
	return &Client{
		client:  client,
		baseUrl: baseUrl,
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

func (c *Client) Deploy(ctx context.Context, req struct{}) (struct{}, error) {
	var res struct{}

	body := bytes.NewBuffer(nil)

	r, err := http.NewRequest("POST", c.baseUrl+"/deploy", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call deploy: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	return res, nil
}

type GithubWebhookRequest struct {
	Ref          string       `json:"ref"`
	Installation Installation `json:"installation"`
	Repository   Repository   `json:"repository"`
}
type Installation struct {
	ID int `json:"id"`
}
type Repository struct {
	CloneUrl string `json:"clone_url"`
}

func (c *Client) GithubWebhook(ctx context.Context, req GithubWebhookRequest) (struct{}, error) {
	var res struct{}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/githubWebhook", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call githubWebhook: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	return res, nil
}

type InfoResponse struct {
	Version string `json:"version"`
}

func (c *Client) Info(ctx context.Context, req struct{}) (InfoResponse, error) {
	var res InfoResponse

	body := bytes.NewBuffer(nil)

	r, err := http.NewRequest("POST", c.baseUrl+"/info", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)

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
	Email    string `json:"email"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

func (c *Client) GetProfile(ctx context.Context, req struct{}) (GetProfileResponse, error) {
	var res GetProfileResponse

	body := bytes.NewBuffer(nil)

	r, err := http.NewRequest("POST", c.baseUrl+"/getProfile", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)

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
