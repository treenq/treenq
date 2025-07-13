import { Badge, type badgeVariants } from '@/components/ui/Badge'
import { Card, CardDescription, CardTitle } from '@/components/ui/Card'
import { VariantProps } from 'class-variance-authority'
import { createEffect, Show } from 'solid-js'

import { Routes } from '@/routes'
import { deployStore } from '@/store/deployStore'

const STATUS_WORKLOAD: Record<string, BadgeVariant> = {
  healthy: 'success',
  degraded: 'warning',
  failing: 'error',
}

type BadgeVariant = VariantProps<typeof badgeVariants>['variant']

export default function WorkloadInfo() {
  const repoID = Routes.deploy.params().id

  createEffect(() => {
    if (repoID) {
      deployStore.getWorkloadStats(repoID)
    }
  })

  return (
    <Show
      when={!deployStore.workloadStatsError}
      fallback={
        <Card class="p-6">
          <CardTitle>Workload Status</CardTitle>
          <CardDescription class="mt-3">{deployStore.workloadStatsError}</CardDescription>
        </Card>
      }
    >
      <Show
        when={deployStore.workloadStats}
        fallback={
          <Card class="p-6">
            <CardTitle>Loading workload stats...</CardTitle>
          </Card>
        }
      >
        {(workload) => (
          <Card class="p-6">
            <div class="mb-3 flex items-center gap-2">
              <CardTitle>Workload Status</CardTitle>
              <Badge variant={STATUS_WORKLOAD[workload().overallStatus] as BadgeVariant}>
                {workload().overallStatus}
              </Badge>
            </div>

            <div class="border-b-border mb-3 grid grid-cols-4 justify-between border-b border-solid pb-3">
              <div>
                <CardDescription>Total Replicas</CardDescription>
                <CardDescription class="mt-0">
                  {workload().replicas.running +
                    workload().replicas.pending +
                    workload().replicas.failed}
                  /{workload().replicas.desired}
                </CardDescription>
              </div>
              <div>
                <CardDescription>Running</CardDescription>
                <CardDescription class="mt-0 flex items-center gap-1">
                  {workload().replicas.running}
                  <Badge variant="success">running</Badge>
                </CardDescription>
              </div>
              <div>
                <CardDescription>Pending</CardDescription>
                <CardDescription class="mt-0 flex items-center gap-1">
                  {workload().replicas.pending}
                  <Badge variant="warning">pending</Badge>
                </CardDescription>
              </div>
              <div>
                <CardDescription>Failed</CardDescription>
                <CardDescription class="mt-0 flex items-center gap-1">
                  {workload().replicas.failed}
                  <Badge variant="error">failed</Badge>
                </CardDescription>
              </div>
            </div>

            <div class="mb-3">
              <CardDescription class="mb-2">Replica Distribution by Version</CardDescription>
              <div class="space-y-2">
                {workload().versions.map((version) => (
                  <div class="bg-muted/50 flex items-center justify-between rounded-md border border-solid p-2">
                    <div class="font-mono text-sm">{version.version}</div>
                    <div class="flex gap-2">
                      {version.replicas.running > 0 && (
                        <Badge variant="success">{version.replicas.running} running</Badge>
                      )}
                      {version.replicas.pending > 0 && (
                        <Badge variant="warning">{version.replicas.pending} pending</Badge>
                      )}
                      {version.replicas.failed > 0 && (
                        <Badge variant="error">{version.replicas.failed} failed</Badge>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        )}
      </Show>
    </Show>
  )
}
