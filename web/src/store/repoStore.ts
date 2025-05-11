import { HttpClient } from '@/services/client'
import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type ReposState = {
  installation: string
  repos: Repo[]
}

type Repo = {
  treenqID: string
  fullName: string
  branch: string
}

const newDefaultAuthState = (): ReposState => ({ installation: '', repos: [] })

function createReposStore() {
  const client = new HttpClient(import.meta.env.APP_API_HOST)

  const [store, setStore] = createStore(newDefaultAuthState())

  return mergeProps(store, {
    connectRepo: async (id: string, branch: string) => {
      await client.connectBranch({ repoID: id, branch: branch })

      setStore('repos', (it) => it.treenqID === id, 'branch', branch)
    },
    getRepos: async () => {
      const res = await client.getRepos()
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
      const res = await client.syncGithubApp()
      if ('error' in res) return

      setStore('installation', res.data.installationID)
    },
  })
}

export const reposStore = createReposStore()
