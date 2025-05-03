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

export type GetProfileResponse = {
  userInfo: UserInfo
}

export type UserInfo = {
  id: string
  email: string
  displayName: string
}

export type ApiErrorPayload = {
  code: string
  message: string
  meta: Record<string, string>
}

export class HttpClient {
  constructor(
    private baseUrl: string,
    private fetchFn: FetchFn = window.fetch.bind(window),
  ) {
    this.baseUrl = baseUrl
    this.fetchFn = fetchFn
  }

  private buildUrl(path: string, query?: Record<string, string | number | boolean>): string {
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
      const jsonErr = await res.json()
      return { error: jsonErr as ApiErrorPayload }
    }

    const response = res.json()
    return { data: response as T }
  }

  private async get<T>(path: string, opts?: RequestOptions): Promise<Result<T>> {
    return await this.request('GET', path, opts)
  }

  private async post<T>(path: string, body?: unknown, opts?: RequestOptions): Promise<Result<T>> {
    return await this.request('POST', path, { ...opts, body: JSON.stringify(body) })
  }

  async getProfile(): Promise<Result<GetProfileResponse>> {
    return await this.post('/getProfile')
  }
}
