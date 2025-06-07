import { Badge, badgeVariants } from '@/components/ui/Badge'
import { Card, CardDescription, CardTitle } from '@/components/ui/Card'
import Console from '@/components/ui/Console'
import {
  BuildProgressMessage,
  DeployResponse,
  GetBuildProgressMessage,
  httpClient,
} from '@/services/client'
import { userStore } from '@/store/userStore'

import { useSolidRoute } from '@/hooks/useSolidRoutre'
import { VariantProps } from 'class-variance-authority'
import { createSignal } from 'solid-js'

const STATUS_DEPLOYMENT = {
  run: 'default',
  failed: 'error',
  done: 'success',
}
interface DeploymentState {
  deployment: DeployResponse
  repoID: string
}

type BadgeVariant = VariantProps<typeof badgeVariants>['variant']

export default function ConsoleDeploy() {
  const [logs, setLogs] = createSignal<BuildProgressMessage[]>([])
  const [dataDeployment, setDataDeployment] = createSignal<DeployResponse>()
  const userName = userStore.user?.displayName
  const { id, stateRoute } = useSolidRoute<DeploymentState>()
  const [dateCereate, setDateCereate] = createSignal({
    start: 0,
    finish: 0,
  })
  const getDeployment = async () => {
    try {
      const res = await httpClient.getDeployment(id)

      setDataDeployment(res.data.deployment)
      dateSecond(res.data.deployment.createdAt)
    } catch (error) {
      console.error('Failed to fetch deployment:', error)
    }
  }

  const dateSecond = (data?: string) => {
    if (!data) return

    const date = new Date(data)
    const seconds = date.getTime() / 1000

    setDateCereate({ start: seconds })
  }

  getDeployment()

  httpClient.listenProgress(id, (data: GetBuildProgressMessage) => {
    if (data.message.final) {
      return
    }

    setLogs((listMessage) => {
      return [...listMessage, data.message]
    })
  })

  return (
    <Card class="p-6">
      <div class="mb-3 flex items-center gap-2">
        <CardTitle>Logs</CardTitle>
        <Badge variant={STATUS_DEPLOYMENT[stateRoute?.deployment?.status || 'run'] as BadgeVariant}>
          Running
        </Badge>
      </div>
      <div class="border-b-border mb-3 grid grid-cols-4 justify-between border-b border-solid pb-3">
        <div>
          <CardDescription>Duration</CardDescription>
          <CardDescription class="mt-0">{dateCereate()}</CardDescription>
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
