import { Badge, type badgeVariants } from '@/components/ui/Badge'
import { Card, CardDescription, CardTitle } from '@/components/ui/Card'
import Console from '@/components/ui/Console'
import {
  BuildProgressMessage,
  type DeploymentStatus,
  type GetBuildProgressMessage,
  httpClient,
} from '@/services/client'
import { userStore } from '@/store/userStore'

import { useTimer } from '@/hooks/useTimer'
import { Routes } from '@/routes'
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
  const [showEmptyState, setShowEmptyState] = createSignal(false)
  const [timeDeploy, setTimeDeploy] = createSignal('0')
  const userName = userStore.user?.displayName
  const params = Routes.deploy.params()
  const { startTimer, time, finishTimer } = useTimer()
  createEffect(() => {
    if (!deployStore.deployment.id) {
      deployStore.getDeployment(params.id)
    }
  })

  httpClient.listenProgress(params.id, (data: GetBuildProgressMessage, isFinish: boolean) => {
    if (data.message.errorCode == 'NO_LOGS') {
      setShowEmptyState(true)
      return
    }
    if (isFinish) {
      getTimeDeploy(logs()[0].timestamp, logs()[logs().length - 1].timestamp)
      return
    }

    setLogs((listMessage) => {
      return [...listMessage, data.message]
    })
  })
  const getTimeDeploy = (start: string, finish: string) => {
    const startDate: Date = new Date(start)
    const finishDate: Date = new Date(finish)

    const totalSeconds: number = Math.floor((finishDate.getTime() - startDate.getTime()) / 1000)

    const seconds = totalSeconds % 60
    const minutes = Math.round(totalSeconds > 60 ? (totalSeconds - seconds) / 60 : 0)

    setTimeDeploy(`${minutes} m ${seconds} s`)
  }

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
          <CardDescription class="mt-0">{`${time().minute} m ${time().second} s`}</CardDescription>
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

      <Console
        classNames="mb-3"
        logs={logs()}
        emptyStateMessage={showEmptyState() ? 'No logs to show' : undefined}
        emptyStateDescription={
          showEmptyState() ? 'The time range is outside your log retention period' : undefined
        }
      />
    </Card>
  )
}
