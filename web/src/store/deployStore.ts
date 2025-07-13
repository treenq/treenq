import { type Deployment, type WorkloadStats, httpClient } from '@/services/client'
import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type DeployState = {
  deployment: Deployment
  workloadStats: WorkloadStats | undefined
  workloadStatsError: string
}

const newDefaultDeployState = (): DeployState => ({
  deployment: {
    id: '',
    fromDeploymentID: '',
    repoID: '',
    space: '',
    sha: '',
    commitMessage: '',
    buildTag: '',
    userDisplayName: '',
    createdAt: '',
    updatedAt: '',
    status: 'run',
    branch: '',
  },
  workloadStats: undefined,
  workloadStatsError: '',
})

function createDeployStore() {
  const [store, setStore] = createStore(newDefaultDeployState())

  return mergeProps(store, {
    getDeployments: async (repoID: string) => {
      const res = await httpClient.getDeployments({ repoID })
      if ('error' in res) return []
      return res.data.deployments
    },
    getDeployment: async (deploymentID: string) => {
      const res = await httpClient.getDeployment({ deploymentID })
      if ('error' in res) return

      setStore('deployment', res.data.deployment)
    },
    setDeployment: (deployment: Deployment) => {
      setStore('deployment', deployment)
    },
    deploy: async (
      repoID: string,
      fromDeploymentID: string,
      branch: string,
      sha: string,
      tag: string,
    ) => {
      const res = await httpClient.deploy({ repoID, fromDeploymentID, branch, sha, tag })
      if ('error' in res) return ''
      return res.data.deployment
    },
    getWorkloadStats: async (repoID: string) => {
      setStore('workloadStatsError', '')
      const res = await httpClient.getWorkloadStats({ repoID })
      if ('error' in res) {
        if (res.error.code === 'NO_PODS_RUNNING') {
          setStore('workloadStatsError', 'No pods are running')
        } else {
          setStore('workloadStatsError', res.error.message || 'Failed to load workload stats')
        }
        return
      }

      setStore('workloadStats', res.data.workloadStats)
      return res.data.workloadStats
    },
  })
}

export const deployStore = createDeployStore()
