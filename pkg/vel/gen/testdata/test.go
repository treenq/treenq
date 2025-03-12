package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	r.Header = c.headers

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
	Data             TestStruct          `json:"data"`
	Chunk            []uint8             `json:"chunk"`
	NextLevelSlice   []HighElem          `json:"slice"`
	Map              map[int]HighMapElem `json:"map"`
	NextLevelNestedP *HighPointer        `json:"nextP"`
}

type TestStruct struct {
	Row              int                   `json:"row"`
	Line             string                `json:"line"`
	NextLevelNested  TestNextLevelStruct   `json:"next"`
	NextLevelSlice   []TestNextLevelElem   `json:"slice"`
	Map              map[int]MapValue      `json:"map"`
	NextLevelNestedP *TestNextLevelStructP `json:"nextP"`
}

type TestNextLevelStruct struct {
	Extra string `json:"extra"`
}

type TestNextLevelElem struct {
	Int int `json:"int"`
}

type MapValue struct {
	Value string
}

type TestNextLevelStructP struct {
	Extra string `json:"extra"`
}

type HighElem struct {
	Int int `json:"int"`
}

type HighMapElem struct {
	Value string
}

type HighPointer struct {
	Extra string `json:"extra"`
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
	r.Header = c.headers

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

func (c *Client) TestEmpty(ctx context.Context) error {
	body := bytes.NewBuffer(nil)

	r, err := http.NewRequest("POST", c.baseUrl+"/testEmpty", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return fmt.Errorf("failed to call testEmpty: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return err
	}

	return nil
}

type GetQuery struct {
	Value string
	Field int
}

type GetResp struct {
	Getting int
}

func (c *Client) TestGet(ctx context.Context, req GetQuery) (GetResp, error) {
	var res GetResp

	q := make(url.Values)
	q.Set("value", req.Value)
	q.Set("field", req.Field)

	r, err := http.NewRequest("GET", c.baseUrl+"/testGet?"+q.Encode(), nil)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call testGet: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode testGet response: %w", err)
	}

	return res, nil
}
