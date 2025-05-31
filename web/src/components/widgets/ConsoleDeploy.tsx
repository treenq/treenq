import { Badge, badgeVariants } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card, CardDescription, CardTitle } from '@/components/ui/Card'
import Console from '@/components/ui/Console'
import {
  BuildProgressMessage,
  DeployResponse,
  GetBuildProgressMessage,
  httpClient,
} from '@/services/client'
import { userStore } from '@/store/userStore'
import { useLocation } from '@solidjs/router'
import { VariantProps } from 'class-variance-authority'
import { createSignal } from 'solid-js'

const MAX_LINES = 100

const STATUS_DEPLOYMENT = {
  run: 'default',
  failed: 'error',
  done: 'success',
}
interface DeploymentState {
  deployment: DeployResponse
}

type BadgeVariant = VariantProps<typeof badgeVariants>['variant']
export default function ConsoleDeploy() {
  const [isExpanded, setIsExpanded] = createSignal(false)
  const [logs, setLogs] = createSignal<BuildProgressMessage[]>([])
  const userName = userStore.user?.displayName

  const location = useLocation<DeploymentState>()

  httpClient.listenProgress(
    location.state?.deployment?.deploymentID || '',
    (data: GetBuildProgressMessage) => {
      setLogs((listMessage) => {
        return [...listMessage, data.message]
      })
    },
  )

  const handleLogsLess = () => {
    setIsExpanded((prev) => !prev)
  }

  return (
    <Card class="p-6">
      <div class="mb-3 flex items-center gap-2">
        <CardTitle>Logs</CardTitle>
        <Badge
          round={true}
          variant={STATUS_DEPLOYMENT[location.state?.deployment?.status || 'run'] as BadgeVariant}
        >
          Running
        </Badge>
      </div>
      <div class="border-b-border mb-3 grid grid-cols-4 justify-between border-b border-solid pb-3">
        <div>
          <CardDescription>Duration</CardDescription>
          <CardDescription class="mt-0">5m 0s</CardDescription>
        </div>
        <div>
          <CardDescription>Branch</CardDescription>
          <CardDescription class="mt-0">Main</CardDescription>
        </div>
        <div>
          <CardDescription>Main</CardDescription>
          <CardDescription class="mt-0">a1b2c3d</CardDescription>
        </div>
        <div>
          <CardDescription>Triggered by</CardDescription>
          <CardDescription class="mt-0">{userName}</CardDescription>
        </div>
      </div>
      <div class="mb-3 text-sm">
        <CardDescription>Commit message</CardDescription>
        <CardDescription class="mt-0">
          feat: Add new authentication flow and improve error handling
        </CardDescription>
      </div>

      <Console classNames="mb-3" logs={isExpanded() ? logs().slice(-MAX_LINES) : logs()} />
      <div class="flex justify-between">
        <CardDescription>
          {`Showing ${isExpanded() ? `latest ${MAX_LINES} of` : 'all  of'} ${logs().length} log lines`}
        </CardDescription>
        <Button
          size="sm"
          class="h-6"
          onClick={handleLogsLess}
          textContent={isExpanded() ? 'Show all' : 'Show less'}
        />
      </div>
    </Card>
  )
}
