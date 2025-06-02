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
  installationID: string
  repos: Repository[]
}

export type Repository = {
  treenqID: string
  full_name: string
  branch: string
}

export type GetBranchesRequest = {
  repoName: string
}

export type GetBranchesResponse = {
  branches: string[]
}

export type DeployRequest = {
  repoID: string
}

export type DeployResponse = {
  deploymentID: string
  fromDeploymentID: string
  repoID: string
  sha: string
  commitMessage: string
  buildTag: string
  userDisplayName: string
  status: 'run' | 'failed' | 'done'
  createdAt: string
  updatedAt: string
}

export type SetSecretRequest = { repoID: string; key: string; value: string }

export type GetSecretsRequest = { repoID: string }

export type GetSecretsResponse = { keys: string[] | null }

export type RevealSecretRequest = { repoID: string; key: string }

export type RevealSecretResponse = { value: string }

export type GetBuildProgressMessage = {
  message: BuildProgressMessage
}

export type TLevelMessage = 'INFO' | 'DEBUG' | 'ERROR'

export type BuildProgressMessage = {
  payload: string
  level: TLevelMessage
  final: boolean
  timestamp: string
}

class HttpClient {
  constructor(
    private baseUrl: string,
    private fetchFn: FetchFn = window.fetch.bind(window),
  ) {
    if (!baseUrl.endsWith('/')) {
      // required in order to fix URL joining without overriding a path after host
      baseUrl = baseUrl + '/'
    }
    this.baseUrl = baseUrl
    this.fetchFn = fetchFn
  }

  private buildUrl(path: string, query?: Record<string, string | number | boolean>): string {
    if (path.startsWith('/')) {
      // required to fix URL joining
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

  private async get<T>(path: string, opts?: RequestOptions): Promise<Result<T>> {
    return await this.request('GET', path, opts)
  }

  private async post<T>(path: string, body?: unknown, opts?: RequestOptions): Promise<Result<T>> {
    return await this.request('POST', path, { ...opts, body: JSON.stringify(body) })
  }

  async getProfile(): Promise<Result<GetProfileResponse>> {
    return await this.post('getProfile')
  }

  async logout(): Promise<Result<undefined>> {
    return await this.post('logout')
  }

  async connectBranch(repo: ConnectBranchRequest): Promise<Result<undefined>> {
    return await this.post('connectRepoBranch', repo)
  }
  async getRepos(): Promise<Result<GetReposResponse>> {
    return await this.post('getRepos')
  }
  async syncGithubApp(): Promise<Result<GetReposResponse>> {
    return await this.post('syncGithubApp')
  }

  async getBranches(req: GetBranchesRequest): Promise<Result<GetBranchesResponse>> {
    return await this.post('getBranches', req)
  }

  async deploy(req: DeployRequest): Promise<Result<DeployResponse>> {
    return await this.post('deploy', req)
  }

  async setSecret(req: SetSecretRequest): Promise<Result<undefined>> {
    return await this.post('setSecret', req)
  }

  async getSecrets(req: GetSecretsRequest): Promise<Result<GetSecretsResponse>> {
    return await this.post('getSecrets', req)
  }

  async revealSecret(req: RevealSecretRequest): Promise<Result<RevealSecretResponse>> {
    return await this.post('revealSecret', req)
  }

  listenProgress(deploymentID: string, callback: (data: GetBuildProgressMessage) => void) {
    const url = this.buildUrl('getBuildProgress', { deploymentID })

    const eventSource = new EventSource(url)

    eventSource.addEventListener('message', (event) => {
      const data: GetBuildProgressMessage = JSON.parse(event.data)
      callback(data)

      if (data.message.final) {
        eventSource.close()
        console.log('FINISH Event Source, listenProgress')
      }
    })
  }
}

export const httpClient = new HttpClient(import.meta.env.APP_API_HOST)
