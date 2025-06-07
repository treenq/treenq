import { Badge, type badgeVariants } from '@/components/ui/Badge'
import { Card, CardDescription, CardTitle } from '@/components/ui/Card'
import Console from '@/components/ui/Console'
import {
  BuildProgressMessage,
  type DeploymentStatus,
  DeployResponse,
  type GetBuildProgressMessage,
  httpClient,
} from '@/services/client'
import { userStore } from '@/store/userStore'

import { useSolidRoute } from '@/hooks/useSolidRoutre'
import { ROUTES } from '@/routes'
import { redirect } from '@solidjs/router'
import { VariantProps } from 'class-variance-authority'
import { createEffect, createSignal } from 'solid-js'

const STATUS_DEPLOYMENT: Record<DeploymentStatus, BadgeVariant> = {
  run: 'default',
  failed: 'error',
  done: 'success',
}
interface DeploymentState {
  deployment: DeployResponse
}

type BadgeVariant = VariantProps<typeof badgeVariants>['variant']

export default function ConsoleDeploy() {
  const [logs, setLogs] = createSignal<BuildProgressMessage[]>([])
  const [dataDeployment, setDataDeployment] = createSignal<DeployResponse>()
  const userName = userStore.user?.displayName
  const { params, stateRoute } = useSolidRoute<DeploymentState>()

  const getDeployment = async () => {
    const res = await httpClient.getDeployment(params.id)
    if ('error' in res) return

    setDataDeployment(res.data)
  }

  createEffect(() => {
    if (!stateRoute.deployment.id) {
      getDeployment()
    }
  })

  httpClient.listenProgress(
    stateRoute.deployment.id || params.id,
    (data: GetBuildProgressMessage) => {
      setLogs((listMessage) => {
        return [...listMessage, data.message]
      })
    },
  )

  return (
    <Card class="p-6">
      <div class="mb-3 flex items-center gap-2">
        <CardTitle>Logs</CardTitle>
        <Badge
          variant={
            STATUS_DEPLOYMENT[
              stateRoute?.deployment?.status || dataDeployment()?.status || 'run'
            ] as BadgeVariant
          }
        >
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
          <CardDescription class="mt-0">{dataDeployment()?.branch}</CardDescription>
        </div>
        <div>
          <CardDescription>Main</CardDescription>
          <CardDescription class="mt-0">{dataDeployment()?.sha.slice(0, 7)}</CardDescription>
        </div>
        <div>
          <CardDescription>Triggered by</CardDescription>
          <CardDescription class="mt-0">{userName}</CardDescription>
        </div>
      </div>
      <div class="mb-3 text-sm">
        <CardDescription>Commit message</CardDescription>
        <CardDescription class="mt-0">{dataDeployment()?.commitMessage}</CardDescription>
      </div>

      <Console classNames="mb-3" logs={logs()} />
    </Card>
  )
}
