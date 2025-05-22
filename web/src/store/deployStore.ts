import { httpClient } from '@/services/client'
import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type DeployState = object

const newDefaultDeployState = (): DeployState => ({})

function createDeployStore() {
  const [store] = createStore(newDefaultDeployState())

  return mergeProps(store, {
    deploy: async (repoID: string): Promise<{ deploymentID: string; status: string; createdAt: string }> => {
      const res = await httpClient.deploy({ repoID })
      if ('error' in res && res.error) {
        // It's better to throw an error that can be caught by the caller
        // and provide more details if available from res.error
        let errorMessage = 'Deployment failed';
        if (typeof res.error === 'string') {
          errorMessage = res.error;
        } else if (typeof res.error === 'object' && res.error !== null && 'message' in res.error) {
          errorMessage = (res.error as { message: string }).message;
        }
        throw new Error(errorMessage);
      }
      // Assuming res.data now holds { deploymentID: string, status: string, createdAt: string }
      // and httpClient.deploy has been updated to reflect this new response structure from the API.
      if (!res.data || typeof res.data.deploymentID === 'undefined') {
        // This case handles if 'error' is not in res, but data is not what we expect.
        throw new Error('Invalid response data from deployment API.');
      }
      return res.data;
    },
  })
}

export const deployStore = createDeployStore()
