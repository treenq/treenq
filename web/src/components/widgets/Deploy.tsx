import { A } from '@/components/ui/A'
import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { Separator } from '@/components/ui/Separator'
import { Routes } from '@/routes'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/Popover'
import { TextField, TextFieldInput, TextFieldLabel } from '@/components/ui/TextField'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/Tooltip'
import { createEffect, createSignal, For, Show } from 'solid-js'

import { Deployment } from '@/services/client'
import { deployStore } from '@/store/deployStore'
import { reposStore, type Repo } from '@/store/repoStore'

type DeployProps = {
  repoID: string
}

export default function Deploy(props: DeployProps) {
  const [deployments, setDeployments] = createSignal<Deployment[]>([])
  const [repo, setRepo] = createSignal<Repo | undefined>()

  const [popoverOpen, setPopoverOpen] = createSignal(false)
  const [deployMode, setDeployMode] = createSignal<'menu' | 'branch' | 'commit' | 'tag'>('menu')
  const [branchInput, setBranchInput] = createSignal('')
  const [commitInput, setCommitInput] = createSignal('')
  const [tagInput, setTagInput] = createSignal('')
  const [loading, setLoading] = createSignal(false)

  const navigateToDeploy = Routes.deploy.navigate()

  const doDeploy = async (fromDeploymentID = '', branch = '', sha = '', tag = '') => {
    setLoading(true)
    const deployment = await deployStore.deploy(props.repoID, fromDeploymentID, branch, sha, tag)
    setLoading(false)
    if (deployment) {
      deployStore.setDeployment(deployment)
      navigateToDeploy({ id: deployment.id })
      setPopoverOpen(false)
      setDeployMode('menu')
      setBranchInput('')
      setCommitInput('')
      setTagInput('')
    }
  }

  createEffect(() => {
    deployStore.getDeployments(props.repoID).then((res) => {
      setDeployments(res)
    })
    reposStore.getRepos().then(() => {
      const repo = reposStore.repos.find((it) => it.treenqID === props.repoID)
      setRepo(repo)
      setBranchInput(repo?.branch || '')
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
          <Popover open={popoverOpen()} onOpenChange={setPopoverOpen}>
            <PopoverTrigger asChild>
              <Button variant="outline" class="hover:bg-primary" aria-expanded={popoverOpen()}>
                Manual Deploy
                <svg class="ml-2 h-4 w-4" viewBox="0 0 20 20" fill="none"><path d="M6 8l4 4 4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </Button>
            </PopoverTrigger>
            <PopoverContent>
              <Show when={deployMode() === 'menu'}>
                <div class="flex flex-col gap-2">
                  <Button variant="ghost" class="justify-start" onClick={() => setDeployMode('branch')}>
                    Deploy a branch
                  </Button>
                  <Button variant="ghost" class="justify-start" onClick={() => setDeployMode('commit')}>
                    Deploy a specific commit
                  </Button>
                  <Button variant="ghost" class="justify-start" onClick={() => setDeployMode('tag')}>
                    Deploy a tag
                  </Button>
                  <Button variant="outline" class="mt-2" onClick={() => doDeploy()}>Deploy from default branch</Button>
                </div>
              </Show>
              <Show when={deployMode() === 'branch'}>
                <div class="flex flex-col gap-2">
                  <TextField>
                    <TextFieldLabel>Branch</TextFieldLabel>
                    <div class="flex items-center gap-2">
                      <TextFieldInput
                        value={branchInput()}
                        onInput={e => setBranchInput(e.currentTarget.value)}
                        placeholder="Enter branch name"
                        class="flex-1"
                      />
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <span class="text-xs text-muted-foreground cursor-pointer">{repo()?.branch}</span>
                        </TooltipTrigger>
                        <TooltipContent>{repo()?.branch} branch</TooltipContent>
                      </Tooltip>
                    </div>
                  </TextField>
                  <div class="flex gap-2 mt-2">
                    <Button
                      variant="primary"
                      disabled={loading() || !branchInput()}
                      onClick={() => doDeploy('', branchInput(), '', '')}
                    >
                      Deploy branch
                    </Button>
                    <Button variant="ghost" onClick={() => setDeployMode('menu')}>Back</Button>
                  </div>
                </div>
              </Show>
              <Show when={deployMode() === 'commit'}>
                <div class="flex flex-col gap-2">
                  <TextField>
                    <TextFieldLabel>Commit SHA</TextFieldLabel>
                    <TextFieldInput
                      value={commitInput()}
                      onInput={e => setCommitInput(e.currentTarget.value)}
                      placeholder="Enter commit SHA"
                    />
                  </TextField>
                  <div class="flex gap-2 mt-2">
                    <Button
                      variant="primary"
                      disabled={loading() || !commitInput()}
                      onClick={() => doDeploy('', '', commitInput(), '')}
                    >
                      Deploy commit
                    </Button>
                    <Button variant="ghost" onClick={() => setDeployMode('menu')}>Back</Button>
                  </div>
                </div>
              </Show>
              <Show when={deployMode() === 'tag'}>
                <div class="flex flex-col gap-2">
                  <TextField>
                    <TextFieldLabel>Tag</TextFieldLabel>
                    <TextFieldInput
                      value={tagInput()}
                      onInput={e => setTagInput(e.currentTarget.value)}
                      placeholder="Enter tag"
                    />
                  </TextField>
                  <div class="flex gap-2 mt-2">
                    <Button
                      variant="primary"
                      disabled={loading() || !tagInput()}
                      onClick={() => doDeploy('', '', '', tagInput())}
                    >
                      Deploy tag
                    </Button>
                    <Button variant="ghost" onClick={() => setDeployMode('menu')}>Back</Button>
                  </div>
                </div>
              </Show>
            </PopoverContent>
          </Popover>
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
                            #{deployment.id}
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
                          onClick={() => doDeploy(deployment.id, '', '', '')}
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
