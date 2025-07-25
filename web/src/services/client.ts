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
  installation: boolean
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
  fromDeploymentID: string
  branch: string
  sha: string
  tag: string
}

export type Deployment = {
  id: string
  fromDeploymentID: string
  repoID: string
  space: unknown
  sha: string
  commitMessage: string
  buildTag: string
  userDisplayName: string
  createdAt: string
  updatedAt: string
  status: DeploymentStatus
  branch: string
}

export type DeploymentStatus = 'run' | 'failed' | 'done'
export type GetDeploymentsRequest = {
  repoID: string
}

export type GetDeploymentsResponse = {
  deployments: Deployment[]
}

export type GetDeploymentRequest = {
  deploymentID: string
}

export type GetDeploymentResponse = {
  deployment: Deployment
}

export type SetSecretRequest = { repoID: string; key: string; value: string }

export type GetSecretsRequest = { repoID: string }

export type GetSecretsResponse = { keys: string[] | undefined }

export type RevealSecretRequest = { repoID: string; key: string }

export type RevealSecretResponse = { value: string }

export type RemoveSecretRequest = { repoID: string; key: string }

export type GetBuildProgressMessage = {
  message: BuildProgressMessage
}

export type TLevelMessage = 'INFO' | 'DEBUG' | 'ERROR'

export type BuildProgressMessage = {
  payload: string
  level: TLevelMessage
  final: boolean
  timestamp: string
  deployment: Deployment
  errorCode: string
}

export type GetWorkloadStatsRequest = {
  repoID: string
}

export type ReplicaInfo = {
  running: number
  pending: number
  failed: number
}

export type VersionInfo = {
  version: string
  replicas: ReplicaInfo
}

export type WorkloadStats = {
  name: string
  replicas: {
    desired: number
    running: number
    pending: number
    failed: number
  }
  versions: VersionInfo[]
  overallStatus: string
}

export type GetWorkloadStatsResponse = {
  workloadStats: WorkloadStats
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

  // private async get<T>(path: string, opts?: RequestOptions): Promise<Result<T>> {
  //   return await this.request('GET', path, opts)
  // }

  private async post<T>(path: string, body?: unknown, opts?: RequestOptions): Promise<Result<T>> {
    return await this.request('POST', path, { ...opts, body: JSON.stringify(body) })
  }

  async getProfile(): Promise<Result<GetProfileResponse>> {
    return await this.post('getProfile')
  }

  async logout(): Promise<Result<void>> {
    return await this.post('logout')
  }

  async connectBranch(repo: ConnectBranchRequest): Promise<Result<void>> {
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

  async deploy(req: DeployRequest): Promise<Result<GetDeploymentResponse>> {
    return await this.post('deploy', req)
  }

  async getDeployments(req: GetDeploymentsRequest): Promise<Result<GetDeploymentsResponse>> {
    return await this.post('getDeployments', req)
  }

  async setSecret(req: SetSecretRequest): Promise<Result<void>> {
    return await this.post('setSecret', req)
  }

  async getSecrets(req: GetSecretsRequest): Promise<Result<GetSecretsResponse>> {
    return await this.post('getSecrets', req)
  }

  async revealSecret(req: RevealSecretRequest): Promise<Result<RevealSecretResponse>> {
    return await this.post('revealSecret', req)
  }

  async removeSecret(req: RemoveSecretRequest): Promise<Result<void>> {
    return await this.post('removeSecret', req)
  }

  async getDeployment(req: GetDeploymentRequest): Promise<Result<GetDeploymentResponse>> {
    return await this.post('getDeployment', req)
  }

  async getWorkloadStats(req: GetWorkloadStatsRequest): Promise<Result<GetWorkloadStatsResponse>> {
    return await this.post('getWorkloadStats', req)
  }

  listenProgress(
    deploymentID: string,
    callback: (data: GetBuildProgressMessage, isFinish: boolean) => void,
  ) {
    const url = this.buildUrl('getBuildProgress', { deploymentID })

    const eventSource = new EventSource(url, { withCredentials: true })

    eventSource.addEventListener('message', (event) => {
      const data: GetBuildProgressMessage = JSON.parse(event.data)
      if (data.message.final) {
        eventSource.close()
        console.log('FINISH Event Source, listenProgress')
        callback(data, true)
        return
      }
      callback(data, false)
    })
  }

  listenLogs(repoID: string, callback: (data: GetBuildProgressMessage) => void) {
    const url = this.buildUrl('getLogs', { repoID })

    const eventSource = new EventSource(url, { withCredentials: true })

    eventSource.addEventListener('message', (event) => {
      const data: GetBuildProgressMessage = JSON.parse(event.data)
      callback(data)

      if (data.message.final) {
        eventSource.close()
        console.log('FINISH Event Source, listenLogs')
      }
    })
  }
}

export const httpClient = new HttpClient(import.meta.env.APP_API_HOST)
