import { createSignal, onMount, Show } from 'solid-js';
import { useLocation, useParams } from '@solidjs/router';
// import { httpClient } from '@/services/client'; // REMOVED
import { deployStore, DeploymentDetailsData } from '@/store/deployStore'; // Added deployStore and DeploymentDetailsData

// interface DeploymentData { // REMOVED - using DeploymentDetailsData from store
//   id: string;
//   status: string;
//   createdAt: Date;
// }

// interface LocationState { // REMOVED as location.state is no longer used
//   id: string;
//   status: string;
//   createdAt: string | Date; 
// }

export default function DeployPage() {
  const [deployment, setDeployment] = createSignal<DeploymentDetailsData | null>(null);
  const [isLoading, setIsLoading] = createSignal<boolean>(true);
  // const location = useLocation(); // REMOVED if not used elsewhere (seems it's not)
  const params = useParams();

  onMount(async () => {
    const deploymentIdFromUrl = params.id;
    // Always fetch based on URL param, location.state is not used for this page's primary role.
    console.log("DeployPage: Initializing, will fetch data for ID from URL:", deploymentIdFromUrl);
    await fetchDeploymentDetails(deploymentIdFromUrl, true);
  });

  async function fetchDeploymentDetails(id: string, setLoading: boolean) {
    if (setLoading) setIsLoading(true);
    try {
      const data = await deployStore.fetchDeploymentById(id); // Use the store method
      if (data) {
        setDeployment(data); // data should already have createdAt as Date
      } else {
        // If data is null (e.g., 404 or other non-exception error from store)
        setDeployment(null);
      }
    } catch (error) {
      // Error is already logged by the store method (if it throws)
      console.error("DeployPage: Error caught from fetchDeploymentById:", error);
      setDeployment(null);
    } finally {
      // Only stop loading if we were the ones to start it explicitly with setLoading=true,
      // or if deployment is still null (error case from this page's perspective)
      if (setLoading || !deployment()) {
        setIsLoading(false);
      }
    }
  }

  return (
    <div>
      <Show when={isLoading()}>
        <div class="animate-pulse p-6 space-y-3">
          {/* Adjusted skeleton to use bg-muted-foreground to match project style */}
          <div class="bg-muted-foreground h-6 w-3/4 rounded"></div> {/* Skeleton for Heading/ID */}
          <div class="bg-muted-foreground h-5 w-1/2 rounded"></div> {/* Skeleton for Status */}
          <div class="bg-muted-foreground h-5 w-1/3 rounded"></div> {/* Skeleton for CreatedAt */}
        </div>
      </Show>

      <Show 
        when={!isLoading() && deployment()} 
        fallback={
          <Show when={!isLoading()}> {/* Only show "not found" if not loading */}
            <div class="p-6">
              <h1 class="text-2xl font-bold mb-4">Deployment Information</h1>
              <p>Deployment not found for ID: {params.id}.</p>
            </div>
          </Show>
        }
      >
        {(dep) => ( // dep is an Accessor<DeploymentData>
          <div class="p-6">
            <h1 class="text-2xl font-bold mb-4">Deployment Details</h1>
            <p><strong>Deployment ID:</strong> {dep().id}</p>
            <p><strong>Status:</strong> {dep().status}</p>
            <p><strong>Created At:</strong> {dep().createdAt.toLocaleString()}</p>
          </div>
        )}
      </Show>
    </div>
  );
}
