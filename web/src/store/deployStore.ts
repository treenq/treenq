import { httpClient } from '@/services/client'
import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type DeployState = object

const newDefaultDeployState = (): DeployState => ({})

function createDeployStore() {
  const [store] = createStore(newDefaultDeployState())

  return mergeProps(store, {
    getDeployments: async (repoID: string) => {
      const res = await httpClient.getDeployments({ repoID })
      if ('error' in res) return []
      return res.data.deployments
    },
    deploy: async (repoID: string) => {
      const res = await httpClient.deploy({ repoID })
      if ('error' in res) return ''
      return res.data
    },
  })
}

export const deployStore = createDeployStore()
