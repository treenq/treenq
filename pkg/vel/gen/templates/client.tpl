package {{ .Client.PackageName }}

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type {{ .Client.TypeName }} struct {
	client *http.Client

	baseUrl string
	headers http.Header
}

func New{{ .Client.TypeName }}(baseUrl string, client *http.Client, headers map[string]string) *{{ .Client.TypeName }} {
	h := make(http.Header)
	for k, v := range headers {
		h.Set(k, v)
	}
	return &{{ .Client.TypeName }}{
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
