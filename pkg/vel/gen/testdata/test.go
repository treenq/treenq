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

type TestTypeNoJsonTags struct {
	Value string
}

func (c *Client) Test1(ctx context.Context, req TestTypeNoJsonTags) (TestTypeNoJsonTags, error) {
	var res TestTypeNoJsonTags

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/test1", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call test1: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode test1 response: %w", err)
	}

	return res, nil
}

type TestTypeNestedTypes struct {
	Data  TestStruct `json:"data"`
	Chunk []uint8    `json:"chunk"`
}
type TestStruct struct {
	Row  int    `json:"row"`
	Line string `json:"line"`
}

func (c *Client) Test2(ctx context.Context, req TestTypeNestedTypes) (TestTypeNestedTypes, error) {
	var res TestTypeNestedTypes

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}
	body := bytes.NewBuffer(bodyBytes)

	r, err := http.NewRequest("POST", c.baseUrl+"/test2", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call test2: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode test2 response: %w", err)
	}

	return res, nil
}

func (c *Client) TestEmpty(ctx context.Context, req struct{}) (struct{}, error) {
	var res struct{}

	body := bytes.NewBuffer(nil)

	r, err := http.NewRequest("POST", c.baseUrl+"/testEmpty", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call testEmpty: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	return res, nil
}
