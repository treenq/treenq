import { SpriteIcon } from '@/components/icons/SpriteIcon'
import { A } from '@/components/ui/A'
import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/Popover'
import { Separator } from '@/components/ui/Separator'
import { TextField, TextFieldInput, TextFieldLabel } from '@/components/ui/TextField'
import { Routes } from '@/routes'
import { createEffect, createSignal, For, Show, type Accessor, type Setter } from 'solid-js'

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
                <A href={Routes.logs.makeHref({ id: props.repoID })} class="text-primary">
                  (logs)
                </A>
              </span>
            </div>
            <h3 class="font-bold">{repo()?.fullName}</h3>
          </div>
          <Popover open={popoverOpen()} onOpenChange={setPopoverOpen}>
            <PopoverTrigger>
              <Button variant="outline" class="hover:bg-primary" aria-expanded={popoverOpen()}>
                Manual Deploy
                <SpriteIcon name="rocket" />
              </Button>
            </PopoverTrigger>
            <PopoverContent>
              <Show when={deployMode() === 'menu'}>
                <div class="flex flex-col gap-2">
                  <Button
                    variant="ghost"
                    class="justify-start"
                    onClick={() => setDeployMode('branch')}
                  >
                    Deploy a branch
                  </Button>
                  <Button
                    variant="ghost"
                    class="justify-start"
                    onClick={() => setDeployMode('commit')}
                  >
                    Deploy a specific commit
                  </Button>
                  <Button
                    variant="ghost"
                    class="justify-start"
                    onClick={() => setDeployMode('tag')}
                  >
                    Deploy a tag
                  </Button>
                  <Button variant="outline" class="mt-2" onClick={() => doDeploy()}>
                    Deploy <SpriteIcon name="branch" />
                    {repo()?.branch}
                  </Button>
                </div>
              </Show>
              <Show when={deployMode() === 'branch'}>
                <DeployAction
                  inputText="Branch"
                  inputPlaceholder="Enter branch name"
                  getter={branchInput}
                  setter={setBranchInput}
                  loading={loading}
                  onDeploy={() => doDeploy('', branchInput(), '', '')}
                  onBack={() => setDeployMode('menu')}
                  deployText="Deploy branch"
                />
              </Show>
              <Show when={deployMode() === 'commit'}>
                <DeployAction
                  inputText="Commit SHA"
                  inputPlaceholder="Enter commit SHA"
                  getter={commitInput}
                  setter={setCommitInput}
                  loading={loading}
                  onDeploy={() => doDeploy('', '', commitInput(), '')}
                  onBack={() => setDeployMode('menu')}
                  deployText="Deploy commit"
                />
              </Show>
              <Show when={deployMode() === 'tag'}>
                <DeployAction
                  inputText="Tag"
                  inputPlaceholder="Enter tag"
                  getter={tagInput}
                  setter={setTagInput}
                  loading={loading}
                  onDeploy={() => doDeploy('', '', '', tagInput())}
                  onBack={() => setDeployMode('menu')}
                  deployText="Deploy tag"
                />
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

                      {deployment.status === 'done' && index() != 0 && (
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

type DeployActionProps = {
  inputText: string
  inputPlaceholder: string
  getter: Accessor<string>
  setter: Setter<string>
  loading: Accessor<boolean>
  onDeploy: () => void
  onBack: () => void
  deployText: string
}
function DeployAction(props: DeployActionProps) {
  return (
    <div class="flex flex-col gap-2">
      <TextField>
        <TextFieldLabel>{props.inputText}</TextFieldLabel>
        <TextFieldInput
          value={props.getter()}
          onInput={(e) => props.setter(e.currentTarget.value)}
          placeholder={props.inputPlaceholder}
        />
      </TextField>
      <div class="mt-2 flex gap-2">
        <Button disabled={props.loading() || !props.getter()} onClick={props.onDeploy}>
          {props.deployText}
        </Button>
        <Button variant="ghost" onClick={props.onBack}>
          Back
        </Button>
      </div>
    </div>
  )
}
