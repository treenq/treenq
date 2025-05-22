import { useLocation } from 'solid-js/router';
import { Match, Switch } from 'solid-js';

export default function DeployPage() {
  const location = useLocation();
  // It's good practice to cast the state and handle potential undefined cases.
  const state = location.state as { deployID: string; status: string; createdAt: string } | undefined;

  return (
    <div class="container mx-auto p-4">
      <h1 class="text-2xl font-bold mb-4">Deployment Details</h1>
      <Switch>
        <Match when={state}>
          {(resolvedState) => ( // resolvedState is guaranteed to be non-null here due to the 'when' condition
            <div>
              <p><strong>ID:</strong> {resolvedState().deployID}</p>
              <p><strong>Status:</strong> {resolvedState().status}</p>
              <p><strong>Created At:</strong> {new Date(resolvedState().createdAt).toLocaleString()}</p>
            </div>
          )}
        </Match>
        <Match when={!state}>
          <div>
            <p>Deployment details are not available. Please initiate a deployment to see details here.</p>
          </div>
        </Match>
      </Switch>
    </div>
  );
}
