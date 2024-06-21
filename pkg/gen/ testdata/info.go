package gen

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/treenq/treenq/pkg/gen"
	"net/http"
)

type InfoResponse struct {
	Version string `json:"version"`
}

type InfoClient struct {
	client  *http.Client
	baseUrl string
}

func (i *InfoClient) Get(ctx context.Context) (InfoResponse, error) {
	r, err := http.NewRequest("GET", i.baseUrl+"/info", nil)
	r = r.WithContext(ctx)
	resp, err := i.client.Do(r)
	var info InfoResponse
	if err != nil {
		return info, fmt.Errorf("failed to get info: %w", err)
	}
	defer resp.Body.Close()

	err = gen.CheckResp(resp)
	if err != nil {
		return info, err
	}

	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return info, fmt.Errorf("failed to decode info: %w", err)
	}
	if info.Version == "" {
		return info, fmt.Errorf("response does not contain version: %w", err)
	}
	return info, nil
}

func NewInfoClient(baseUrl string, client *http.Client) *InfoClient {
	return &InfoClient{
		client:  client,
		baseUrl: baseUrl,
	}
}
