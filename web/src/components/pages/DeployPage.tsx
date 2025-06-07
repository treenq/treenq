import { DeployResponse, httpClient } from '@/services/client'
import { useLocation } from '@solidjs/router'

import ConsoleDeploy from '@/components/widgets/ConsoleDeploy'

export default function DeploymentDetailsPage() {
  const location = useLocation()
  const deploy = location.state as DeployResponse | undefined

  if (!deploy) {
    // const deployID = useParams().id
    //TODO: get deployment by ID
  }

  return (
    <div>
      <h1 class="px-8 py-4">Deployment Triggered</h1>
      <div class="px-8">
        <ConsoleDeploy />
      </div>
    </div>
  )
}
