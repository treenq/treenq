import { A } from '@/components/ui/A'
import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { Separator } from '@/components/ui/Separator'
import { Routes } from '@/routes'

import { Deployment } from '@/services/client'
import { deployStore } from '@/store/deployStore'
import { reposStore, type Repo } from '@/store/repoStore'
import { useNavigate } from '@solidjs/router'
import { createEffect, createSignal, For } from 'solid-js'

type DeployProps = {
  repoID: string
}

export default function Deploy(props: DeployProps) {
  const [deployments, setDeployments] = createSignal<Deployment[]>([])
  const [repo, setRepo] = createSignal<Repo | undefined>()

  const deploy = async (fromDeploymentID: string = '') => {
    const deployment = await deployStore.deploy(props.repoID, fromDeploymentID)
    if (deployment) {
      deployStore.setDeployment(deployment)
      Routes.deploy.navigate(useNavigate, { id: deployment.id })
    }
  }

  createEffect(() => {
    deployStore.getDeployments(props.repoID).then((res) => {
      setDeployments(res)
    })
    reposStore.getRepos().then(() => {
      const repo = reposStore.repos.find((it) => it.treenqID === props.repoID)
      setRepo(repo)
    })
  })

  return (
    <div class="mx-auto flex w-full flex-col">
      <div class="bg-background text-foreground p-6">
        <div class="mb-4 flex items-center justify-between">
          <div>
            <div class="text-muted-foreground mb-1 flex items-center gap-2 text-sm">
              <span class="inline-flex items-center gap-1">
                <div class="bg-muted h-4 w-4 rounded" />
                SERVICE
              </span>
            </div>
            <h3 class="font-bold">{repo()?.fullName}</h3>
          </div>
          <Button variant="outline" class="hover:bg-primary" onclick={() => deploy()}>
            Deploy Now
            <div class="bg-muted ml-2 h-4 w-4 rounded" />
          </Button>
        </div>

        <div class="mb-2 flex items-center gap-4">
          <div class="flex items-center gap-2">
            <div class="bg-muted h-4 w-4 rounded-full" />
            <span class="text-sm">{repo()?.fullName}</span>
          </div>
          <div class="flex items-center gap-2 text-sm">
            <div class="bg-muted h-4 w-4 rounded-full" />
            <span>{repo()?.branch}</span>
          </div>
        </div>

        <div class="text-muted-foreground flex items-center gap-2 text-sm">
          <span>repoID: {props.repoID}</span>
          <div class="bg-muted h-4 w-4 rounded" />
        </div>
      </div>

      <Card class="rounded-none border-0">
        <CardContent class="p-0">
          <For each={deployments()}>
            {(deployment, index) => (
              <div>
                <div class="flex items-start gap-4 p-6">
                  <div class="bg-muted flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full">
                    <div class="bg-muted-foreground h-4 w-4" />
                  </div>

                  <div class="flex-1">
                    <div class="flex items-center justify-between">
                      <div>
                        <span class="text-base font-medium">
                          Deploy
                          <A
                            class="text-primary"
                            href={Routes.deploy.makeHref({ id: deployment.id })}
                          >
                            #${deployment.id}
                          </A>{' '}
                          {deployment.status === 'run' ? 'live' : deployment.status}
                          for {deployment.sha.slice(0, 7)}
                          {': '}
                          {deployment.commitMessage}
                        </span>
                        <p class="text-muted-foreground mt-1 text-sm">{deployment.createdAt}</p>
                      </div>

                      {deployment.status === 'done' && (
                        <Button
                          variant="outline"
                          size="sm"
                          class="gap-1"
                          onClick={() => deploy(deployment.id)}
                        >
                          <div class="bg-muted h-4 w-4 rounded" />
                          Rollback
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
                {index() < deployments().length - 1 && <Separator />}
              </div>
            )}
          </For>
        </CardContent>
      </Card>
    </div>
  )
}
