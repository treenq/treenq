package gen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type InfoRequest struct {
}

type InfoResponse struct {
	Version string `json:"version"`
}

type InfoClient struct {
	client  *http.Client
	baseUrl string
}

func (c *InfoClient) Info(ctx context.Context, _ InfoRequest) (InfoResponse, error) {
	var res InfoResponse

	r, err := http.NewRequest("POST", c.baseUrl+"/info", nil)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}

	r = r.WithContext(ctx)
	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to get info: %w", err)
	}

	defer resp.Body.Close()

	err = CheckResp(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode info: %w", err)
	}
	return res, nil
}

func NewInfoClient(baseUrl string, client *http.Client) *InfoClient {
	return &InfoClient{
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

func CheckResp(resp *http.Response) error {
	if resp.StatusCode >= 500 {
		return &Error{
			Code:    "UNKNOWN",
			Message: fmt.Sprintf("failed to get info (code: %d)", resp.StatusCode),
		}
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
