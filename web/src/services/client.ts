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

export type ConnectBranchRequest = {
  repoID: string
  branch: string
}

export type GetReposResponse = {
  installation: string
  repos: Repository[]
}

export type Repository = {
  treenqID: string
  full_name: string
  branch: string
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

  private async get<T>(path: string, opts?: RequestOptions): Promise<Result<T>> {
    return await this.request('GET', path, opts)
  }

  private async post<T>(path: string, body?: unknown, opts?: RequestOptions): Promise<Result<T>> {
    return await this.request('POST', path, { ...opts, body: JSON.stringify(body) })
  }

  async getProfile(): Promise<Result<GetProfileResponse>> {
    return await this.post('/getProfile')
  }

  async logout(): Promise<Result<undefined>> {
    return await this.post('/logout')
  }

  async connectBranch(repo: ConnectBranchRequest): Promise<Result<undefined>> {
    return await this.post('/connectRepoBranch', repo)
  }
  async getRepos(): Promise<Result<GetReposResponse>> {
    return await this.post('/getRepos')
  }
}
