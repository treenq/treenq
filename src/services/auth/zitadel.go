package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrProviderNotFound = errors.New("provider not found")

type Zitadel struct {
	client *http.Client

	domain     string
	token      string
	idps       map[string]string
	successUrl string
	failureUrl string
}

func NewZitadel(
	domain string,
	token string,
	idps map[string]string,
	successUrl string,
	failureUrl string,
) *Zitadel {
	return &Zitadel{
		client: http.DefaultClient,

		domain:     domain,
		token:      token,
		idps:       idps,
		successUrl: successUrl,
		failureUrl: failureUrl,
	}
}

type intentRequest struct {
	IDPId string     `json:"idpId"`
	Urls  intentUrls `json:"urls"`
}

type intentUrls struct {
	SuccessURL string `json:"successUrl"`
	FailureURL string `json:"failureUrl"`
}

type intentResponse struct {
	AuthURL string `json:"authUrl"`
}

func (z *Zitadel) Start(ctx context.Context, provider string) (string, error) {
	idpID, ok := z.idps[provider]
	if !ok {
		return "", ErrProviderNotFound
	}

	// Prepare the request payload
	reqPayload := intentRequest{
		IDPId: idpID,
		Urls: intentUrls{
			SuccessURL: z.successUrl,
			FailureURL: z.failureUrl,
		},
	}

	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request payload: %w", err)
	}

	url := fmt.Sprintf("%s/v2/idp_intents", z.domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+z.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := z.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected response: %s", string(body))
	}

	var respPayload intentResponse
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return respPayload.AuthURL, nil
}
