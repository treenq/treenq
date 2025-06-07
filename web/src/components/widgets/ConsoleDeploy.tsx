import { Badge, type badgeVariants } from '@/components/ui/Badge'
import { Card, CardDescription, CardTitle } from '@/components/ui/Card'
import Console from '@/components/ui/Console'
import {
  BuildProgressMessage,
  type DeploymentStatus,
  type GetBuildProgressMessage,
  type GetDeploymentResponse,
  httpClient,
} from '@/services/client'
import { userStore } from '@/store/userStore'

import { useSolidRoute } from '@/hooks/useSolidRoutre'
import { deployStore } from '@/store/deployStore'
import { VariantProps } from 'class-variance-authority'
import { createEffect, createSignal } from 'solid-js'

const STATUS_DEPLOYMENT: Record<DeploymentStatus, BadgeVariant> = {
  run: 'default',
  failed: 'error',
  done: 'success',
}

type BadgeVariant = VariantProps<typeof badgeVariants>['variant']

export default function ConsoleDeploy() {
  const [logs, setLogs] = createSignal<BuildProgressMessage[]>([])
  const userName = userStore.user?.displayName
  const { params } = useSolidRoute<GetDeploymentResponse>()

  createEffect(() => {
    if (!deployStore.deployment.id) {
      deployStore.getDeployment(params.id)
    }
  })

  httpClient.listenProgress(params.id, (data: GetBuildProgressMessage) => {
    setLogs((listMessage) => {
      return [...listMessage, data.message]
    })
  })

  return (
    <Card class="p-6">
      <div class="mb-3 flex items-center gap-2">
        <CardTitle>Logs</CardTitle>
        <Badge variant={STATUS_DEPLOYMENT[deployStore.deployment.status || 'run'] as BadgeVariant}>
          Running
        </Badge>
      </div>
      <div class="border-b-border mb-3 grid grid-cols-4 justify-between border-b border-solid pb-3">
        <div>
          <CardDescription>Duration</CardDescription>
          <CardDescription class="mt-0">{0}</CardDescription>
        </div>
        <div>
          <CardDescription>Branch</CardDescription>
          <CardDescription class="mt-0">{deployStore.deployment.branch}</CardDescription>
        </div>
        <div>
          <CardDescription>Main</CardDescription>
          <CardDescription class="mt-0">{deployStore.deployment.sha.slice(0, 7)}</CardDescription>
        </div>
        <div>
          <CardDescription>Triggered by</CardDescription>
          <CardDescription class="mt-0">{userName}</CardDescription>
        </div>
      </div>
      <div class="mb-3 text-sm">
        <CardDescription>Commit message</CardDescription>
        <CardDescription class="mt-0">{deployStore.deployment.commitMessage}</CardDescription>
      </div>

      <Console classNames="mb-3" logs={logs()} />
    </Card>
  )
}
