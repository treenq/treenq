import { httpClient } from '@/services/client'
import { mergeProps } from 'solid-js'
import { createStore, SetStoreFunction } from 'solid-js/store' // Ensure SetStoreFunction is imported

// Define the shape of a recent deployment
interface RecentDeploy {
  id: string;
  status: string;
  createdAt: string; // Ensure this is string
}

// Update DeployState to include recentDeploy
interface DeployState {
  recentDeploy: RecentDeploy | null;
  // Removed other fields like currentDeploymentDetails
}

// Initialize recentDeploy to null
const newDefaultDeployState = (): DeployState => ({
  recentDeploy: null,
});

function createDeployStore() {
  const [store, setStore] = createStore<DeployState>(newDefaultDeployState());

  return mergeProps(store, {
    deploy: async (repoID: string) => {
      const res = await httpClient.deploy({ repoID });
      if ('error' in res) {
        setStore('recentDeploy', null);
        return '';
      }

      // Assuming res.data contains { deploymentID, status, createdAt }
      const { deploymentID, status, createdAt } = res.data; // createdAt here is a string

      setStore('recentDeploy', {
        id: deploymentID,
        status: status,
        createdAt: createdAt, // Store as string
      });

      return deploymentID;
    },
    // Removed fetchDeploymentById and other actions if they were not part of the original simple state
  })
}

export const deployStore = createDeployStore()
