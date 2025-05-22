import { useLocation } from '@solidjs/router';
import { Show } from 'solid-js';

interface DeploymentDetailsState {
  id: string;
  status: string;
  createdAt: string; // This will be a string from route state
}

export default function DeploymentDetailsPage() {
  const location = useLocation();
  // Attempt to cast state. Solid's location.state is 'unknown' by default.
  const details = location.state as DeploymentDetailsState | undefined;

  return (
    <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
      <h1>Deployment Triggered</h1> {/* Changed heading */}
      <Show
        when={details && details.id} // Check if details and details.id exist for robustness
        fallback={<p>Deployment details not available or state not passed correctly.</p>} // Updated fallback
      >
        {(d) => ( // d is an Accessor to details, so d().id etc.
          <div>
            <p><strong>Deployment ID:</strong> {d().id}</p>
            <p><strong>Status:</strong> {d().status}</p>
            <p><strong>Created At:</strong> {d().createdAt}</p>
            {/* Corrected p_ex to p and added the informational message */}
            <p><em>This page displays details for the deployment just initiated.</em></p>
          </div>
        )}
      </Show>
    </div>
  );
}
