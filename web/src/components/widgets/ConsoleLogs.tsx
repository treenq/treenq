import { Card, CardTitle } from '@/components/ui/Card'
import Console from '@/components/ui/Console'
import {
  BuildProgressMessage,
  type GetBuildProgressMessage,
  httpClient,
} from '@/services/client'
import { createEffect, createSignal } from 'solid-js'

type ConsoleLogsProps = {
  repoID: string
}

export default function ConsoleLogs(props: ConsoleLogsProps) {
  const [logs, setLogs] = createSignal<BuildProgressMessage[]>([])

  createEffect(() => {
    httpClient.listenLogs(props.repoID, (data: GetBuildProgressMessage) => {
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

      <Console classNames="mb-3" logs={logs()} />
    </Card>
  )
}