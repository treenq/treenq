type FetchFn = typeof fetch

type RequestOptions = Omit<RequestInit, 'method'> & {
  query?: Record<string, string | number | boolean>
  headers?: Record<string, string>
}

export type Failure = {
  error: ApiErrorPayload
}

export type Success<T = void> = {
  data: T
}

export type Result<T = void> = Failure | Success<T>

export type ApiErrorPayload = {
  code: string
  message: string
  meta: Record<string, string>
}

{{- range .Apis }}
{{- range .DataTypes }}
export type {{ .Name }} = {
  {{- range .Fields }}
  {{- if ne .JsonTag "" }}
  {{ .JsonTag }}: {{ .TSTypeName }}
  {{- else }}
  {{ .Name }}: {{ .TSTypeName }}
  {{- end }}
  {{- end }}
}

{{ end }}

{{ end }}

class {{ .Client.TypeName }} {
  constructor(
    private baseUrl: string,
    private fetchFn: FetchFn = window.fetch.bind(window),
  ) {
    if (!baseUrl.endsWith('/')) {
      baseUrl = baseUrl + '/'
    }
    this.baseUrl = baseUrl
    this.fetchFn = fetchFn
  }

  private buildUrl(path: string, query?: Record<string, string | number | boolean>): string {
    if (path.startsWith('/')) {
      path = path.slice(1)
    }
    const url = new URL(path, this.baseUrl)
    if (query) {
      for (const [key, val] of Object.entries(query)) {
        url.searchParams.set(key, String(val))
      }
    }
    return url.toString()
  }

  private async request<T>(
    method: string,
    path: string,
    opts: RequestOptions = {},
  ): Promise<Result<T>> {
    const url = this.buildUrl(path, opts.query)
    const res = await this.fetchFn(url, {
      method,
      credentials: 'include',
      ...opts,
      headers: {
        'Content-Type': 'application/json',
        ...opts.headers,
      },
    })

    if (!res.ok) {
      if (res.status >= 500) {
        const errText = await res.text()
        throw Error('http error: ' + errText)
      }
      const jsonErr = await res.json()
      return { error: jsonErr as ApiErrorPayload }
    }

    const response = await res.text()
    if (response) {
      const resp = JSON.parse(response)
      return { data: resp as T }
    }
    return { data: {} as T }
  }

  private async post<T>(path: string, body?: unknown, opts?: RequestOptions): Promise<Result<T>> {
    return await this.request('POST', path, { ...opts, body: JSON.stringify(body) })
  }

  private async get<T>(path: string, opts?: RequestOptions): Promise<Result<T>> {
    return await this.request('GET', path, opts)
  }

{{- range .Apis }}
  async {{ .FuncName }}({{ if ne .Input.Name "" }}req: {{ .Input.Name }}{{ end }}): Promise<Result<{{ if ne .Output.Name "" }}{{ .Output.Name }}{{ else }}void{{ end }}>> {
    {{- if eq .Method "GET" }}
    const query: Record<string, string | number | boolean> = {}
    {{- range .Input.Fields }}
    query['{{ .SchemaTag }}'] = req.{{ .Name }}
    {{- end }}
    return await this.get('{{ .OperationID }}', { query })
    {{- else }}
    return await this.post('{{ .OperationID }}'{{ if ne .Input.Name "" }}, req{{ end }})
    {{- end }}
  }
{{ end }}
}

