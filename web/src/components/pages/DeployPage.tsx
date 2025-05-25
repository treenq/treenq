import { DeployResponse } from '@/services/client'
import { useLocation } from '@solidjs/router'
import { Show } from 'solid-js'

export default function DeploymentDetailsPage() {
  const location = useLocation()
  const deploy = location.state as DeployResponse | undefined
  if (!deploy) {
    // const deployID = useParams().id
    //TODO: get deployment by ID
  }

  return (
    <div>
      <h1>Deployment Triggered</h1>
      <Show
        when={deploy}
        fallback={<p>Deployment details not available or state not passed correctly.</p>}
      >
        {(d) => (
          <div>
            <p>
              <strong>Deployment ID:</strong> {d().deploymentID}
            </p>
            <p>
              <strong>Status:</strong> {d().status}
            </p>
            <p>
              <strong>Created At:</strong> {d().createdAt}
            </p>
            <p>
              <em>This page displays details for the deployment just initiated.</em>
            </p>
          </div>
        )}
      </Show>
    </div>
  )
}
