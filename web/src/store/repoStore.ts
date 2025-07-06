import { httpClient } from '@/services/client'
import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type ReposState = {
  installation: boolean
  repos: Repo[]
  isSyncing: boolean
}

export type Repo = {
  treenqID: string
  fullName: string
  branch: string
}

const newDefaultRepoState = (): ReposState => ({ installation: false, repos: [], isSyncing: false })

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
      setStore('installation', res.data.installation)
      return res.data.repos
    },
    syncGithubApp: async () => {
      setStore('isSyncing', true)
      try {
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
        setStore('installation', res.data.installation)
      } finally {
        setStore('isSyncing', false)
      }
    },
    getBranches: async (repoName: string) => {
      const res = await httpClient.getBranches({ repoName: repoName })
      if ('error' in res) return []
      return (res.data.branches as string[]) || []
    },
  })
}

export const reposStore = createReposStore()
