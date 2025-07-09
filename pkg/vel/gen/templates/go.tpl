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
{{- range .Apis }}
{{- range .DataTypes }}
type {{ .Name }} struct {
	{{- range .Fields }}
	{{ .Name }} {{ .TypeName }}{{ if ne .JsonTag "" }} `json:"{{ .JsonTag }}"`{{ end }}
	{{- end }}
}

{{ end }}

func (c *{{ $.Client.TypeName }}) {{ .FuncName }}(ctx context.Context{{ if ne .Input.Name "" }}, req {{ .Input.Name }}{{ end }}) ({{if ne .Output.Name "" }}{{ .Output.Name }}, {{ end }}error) {
    {{- if gt (len .Output.Fields) 0 }}
    var res {{ .Output.Name }}

    {{ end }}
    {{- if eq .Method "GET" }}
	q := make(url.Values)

	{{- range .Input.Fields }}
    q.Set("{{ .SchemaTag }}", req.{{ .Name }})
	{{- end }}

    r, err := http.NewRequest("GET", c.baseUrl+"/{{ .OperationID }}?" + q.Encode(), nil)
    {{- else }}
    {{- if ne .Input.Name "" }}
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return {{if ne .Output.Name "" }}res, {{ end }}fmt.Errorf("failed to marshal request: %w", err)
	}
    body := bytes.NewBuffer(bodyBytes)
    {{- else }}
    body := bytes.NewBuffer(nil)
    {{- end }}

	r, err := http.NewRequest("POST", c.baseUrl+"/{{ .OperationID }}", body)
    {{- end }}
	if err != nil {
		return {{if ne .Output.Name "" }}res, {{ end }}fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)
	r.Header = c.headers

	resp, err := c.client.Do(r)
	if err != nil {
		return {{if ne .Output.Name "" }}res, {{ end }}fmt.Errorf("failed to call {{ .OperationID }}: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return {{if ne .Output.Name "" }}res, {{ end }}err
	}
	{{- if gt (len .Output.Fields) 0 }}

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return {{if ne .Output.Name "" }}res, {{ end }}fmt.Errorf("failed to decode {{ .OperationID }} response: %w", err)
	}
	{{- end }}

	return {{if ne .Output.Name "" }}res, {{ end }}nil
}

{{- end }}

