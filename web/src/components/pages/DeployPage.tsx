import { useParams } from '@solidjs/router'

export default function DeployPage() {
  const params = useParams()
  const deploymentID = params.id

  return (
    <div>
      <h1>Deployment Details</h1>
      <p>Deployment ID: {deploymentID}</p>
    </div>
  )
}
