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
import { createEffect, createSignal, onCleanup } from 'solid-js'

const STATUS_DEPLOYMENT: Record<DeploymentStatus, BadgeVariant> = {
  run: 'default',
  failed: 'error',
  done: 'success',
}

type BadgeVariant = VariantProps<typeof badgeVariants>['variant']

export default function ConsoleDeploy() {
  const [logs, setLogs] = createSignal<BuildProgressMessage[]>([])
  const [showEmptyState, setShowEmptyState] = createSignal(false)
  const [actualDuration, setActualDuration] = createSignal(0)
  const userName = userStore.user?.displayName
  const params = Routes.deploy.params()
  const { startTimer, time, finishTimer } = useTimer()
  createEffect(() => {
    if (!deployStore.deployment.id) {
      deployStore.getDeployment(params.id)
    } else if (deployStore.deployment.status === 'run' && !actualDuration()) {
      startTimer()
    }
  })

  httpClient.listenProgress(params.id, (data: GetBuildProgressMessage, isFinish: boolean) => {
    if (data.message.errorCode == 'NO_LOGS') {
      finishTimer()
      setShowEmptyState(true)
      return
    }

    setLogs((listMessage) => {
      const newLogs = [...listMessage, data.message]

      if (isFinish && newLogs.length > 0) {
        finishTimer()
        const startTime = new Date(newLogs[0].timestamp)
        const endTime = new Date(newLogs[newLogs.length - 1].timestamp)
        const totalSeconds = Math.floor((endTime.getTime() - startTime.getTime()) / 1000)
        setActualDuration(totalSeconds)
      }

      return newLogs
    })
  })

  onCleanup(() => {
    finishTimer()
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
          <CardDescription class="mt-0">
            {actualDuration()
              ? `${Math.floor(actualDuration() / 60)} m ${actualDuration() % 60} s`
              : `${time().minute} m ${time().second} s`}
          </CardDescription>
        </div>
        <div>
          <CardDescription>Branch</CardDescription>
          <CardDescription class="mt-0">{deployStore.deployment.branch}</CardDescription>
        </div>
        <div>
          <CardDescription>Commit</CardDescription>
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
