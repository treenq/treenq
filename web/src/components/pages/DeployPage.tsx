import { Button } from '@/components/ui/Button'
import ConsoleDeploy from '@/components/widgets/ConsoleDeploy'
import { useSolidRoute } from '@/hooks/useSolidRoutre'

export default function DeploymentDetailsPage() {
  const { backPage } = useSolidRoute()

  return (
    <div class="px-8">
      <div class="mb-6 w-full">
        <Button onClick={() => backPage()} textContent="Back" variant="outline"></Button>
      </div>
      <h1 class="mb-6">Deployment Triggered</h1>
      <div class="">
        <ConsoleDeploy />
      </div>
    </div>
  )
}
