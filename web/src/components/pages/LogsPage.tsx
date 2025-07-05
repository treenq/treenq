import { Button } from '@/components/ui/Button'
import ConsoleLogs from '@/components/widgets/ConsoleLogs'
import { useNavigate } from '@solidjs/router'

import { Routes } from '@/routes'

export default function LogsPage() {
  const logsRoute = Routes.logs
  const repoID = logsRoute.params().id
  const nav = useNavigate()

  return (
    <main class="bg-background flex min-h-screen w-full flex-col px-8 py-12">
      <div class="mb-6">
        <Button onClick={() => nav(-1)} textContent="Back" variant="outline"></Button>
      </div>
      <div class="flex w-full flex-col gap-6">
        <div class="flex items-center gap-2">
          <h1 class="text-2xl font-bold">Logs</h1>
          <span class="text-muted-foreground">for {repoID}</span>
        </div>
        <ConsoleLogs repoID={repoID} />
      </div>
    </main>
  )
}