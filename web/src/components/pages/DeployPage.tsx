import { DeployResponse } from '@/services/client'
import { useLocation } from '@solidjs/router'

import ConsoleDeploy from '../widgets/consoleDeploy'

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
      <ConsoleDeploy />
    </div>
  )
}
