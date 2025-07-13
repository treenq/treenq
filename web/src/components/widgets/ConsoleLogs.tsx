import { Card, CardTitle } from '@/components/ui/Card'
import Console from '@/components/ui/Console'
import {
  type BuildProgressMessage,
  type GetBuildProgressMessage,
  httpClient,
} from '@/services/client'
import { createEffect, createSignal } from 'solid-js'

type ConsoleLogsProps = {
  repoID: string
}

export default function ConsoleLogs(props: ConsoleLogsProps) {
  const [logs, setLogs] = createSignal<BuildProgressMessage[]>([])
  const [showEmptyState, setShowEmptyState] = createSignal(false)

  createEffect(() => {
    httpClient.listenLogs(props.repoID, (data: GetBuildProgressMessage) => {
      if (data.message.errorCode == 'NO_PODS_RUNNING') {
        setShowEmptyState(true)
        return
      }

      setLogs((listMessage) => {
        return [...listMessage, data.message]
      })
    })
  })

  return (
    <Card class="p-6">
      <div class="mb-3 flex items-center gap-2">
        <CardTitle>Service Logs</CardTitle>
      </div>

      <Console
        classNames="mb-3"
        logs={logs()}
        emptyStateMessage={showEmptyState() ? 'No pods are running' : undefined}
        emptyStateDescription={
          showEmptyState() ? 'Deploy a workload to start reading the logs' : undefined
        }
      />
    </Card>
  )
}
