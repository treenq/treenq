import { httpClient } from '@/services/client'
import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type DeployState = object

const newDefaultDeployState = (): DeployState => ({})

function createDeployStore() {
  const [store] = createStore(newDefaultDeployState())

  return mergeProps(store, {
    deploy: async (repoID: string) => {
      const res = await httpClient.deploy({ repoID })
      if ('error' in res) return ''
      return res.data.deploymentID
    },
  })
}

export const deployStore = createDeployStore()
