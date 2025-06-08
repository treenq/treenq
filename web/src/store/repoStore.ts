import { httpClient } from '@/services/client'
import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type ReposState = {
  installation: string
  repos: Repo[]
}

export type Repo = {
  treenqID: string
  fullName: string
  branch: string
}

const newDefaultRepoState = (): ReposState => ({ installation: '', repos: [] })

function createReposStore() {
  const [store, setStore] = createStore(newDefaultRepoState())

  return mergeProps(store, {
    connectRepo: async (id: string, branch: string) => {
      await httpClient.connectBranch({ repoID: id, branch: branch })

      setStore('repos', (it) => it.treenqID === id, 'branch', branch)
    },
    getRepos: async () => {
      const res = await httpClient.getRepos()
      if ('error' in res) {
        return []
      }

      setStore(
        'repos',
        (res.data.repos || []).map((it) => ({
          treenqID: it.treenqID,
          fullName: it.full_name,
          branch: it.branch,
        })),
      )
      setStore('installation', res.data.installationID)
      return res.data.repos
    },
    syncGithubApp: async () => {
      const res = await httpClient.syncGithubApp()
      if ('error' in res) return

      setStore(
        'repos',
        (res.data.repos || []).map((it) => ({
          treenqID: it.treenqID,
          fullName: it.full_name,
          branch: it.branch,
        })),
      )
      setStore('installation', res.data.installationID)
    },
    getBranches: async (repoName: string) => {
      const res = await httpClient.getBranches({ repoName: repoName })
      if ('error' in res) return []
      return (res.data.branches as string[]) || []
    },
  })
}

export const reposStore = createReposStore()
