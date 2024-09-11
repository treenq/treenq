{{- range .DataTypes }}
type {{ .Name }} struct {
	{{- range .Fields }}
	{{ .Name }} {{ .TypeName }}{{ if ne .JsonTag "" }} `json:"{{ .JsonTag }}"`{{ end }}
	{{- end }}
}
{{- end }}
func (c *Client) {{ .FuncName }}(ctx context.Context, req {{ .Input.Name }}) ({{ .Output.Name }}, error) {
	var res {{ .Output.Name }}
	{{ if gt (len .Input.Fields) 0 }}
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return res, fmt.Errorf("failed to marshal request: %w", err)
	}
    body := bytes.NewBuffer(bodyBytes)
    {{ else }}
    body := bytes.NewBuffer(nil)
    {{ end }}
	r, err := http.NewRequest("POST", c.baseUrl+"/{{ .OperationID }}", body)
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	r = r.WithContext(ctx)

	resp, err := c.client.Do(r)
	if err != nil {
		return res, fmt.Errorf("failed to call {{ .OperationID }}: %w", err)
	}
	defer resp.Body.Close()

	err = HandleErr(resp)
	if err != nil {
		return res, err
	}
	{{ if gt (len .Output.Fields) 0 }}
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode {{ .OperationID }} response: %w", err)
	}
	{{ end }}
	return res, nil
}
