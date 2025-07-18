import { A } from '@/components/ui/A'
import { Button } from '@/components/ui/Button'
import { Card, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import {
  Combobox,
  ComboboxContent,
  ComboboxControl,
  ComboboxErrorMessage,
  ComboboxInput,
  ComboboxItem,
  ComboboxItemIndicator,
  ComboboxItemLabel,
  ComboboxTrigger,
} from '@/components/ui/Combobox'
import { Skeleton } from '@/components/ui/Skeleton'
import { Routes } from '@/routes'
import { reposStore } from '@/store/repoStore'
import { For, Show, createEffect, createSignal, onMount, type JSX } from 'solid-js'

type ConnectReposItem = {
  id: string
  fullName: string
  branch: string
}

function RepoItem(props: ConnectReposItem) {
  const repoHref = Routes.repos.makeHref({ id: props.id })

  return (
    <Card class="mx-auto w-full max-w-2xl">
      <CardHeader class="flex-row items-center justify-between gap-4 p-6 pb-2">
        <div class="min-w-0 flex-1">
          <CardTitle class="truncate">
            <Show when={props.branch !== ''} fallback={props.fullName}>
              <A variant="light" href={repoHref}>
                {props.fullName}
              </A>
            </Show>
          </CardTitle>
          <CardDescription>
            <span class="inline-flex items-center gap-1">
              {' '}
              <Show when={props.branch}>
                ✅ branch: <b>{props.branch}</b>
              </Show>
            </span>
          </CardDescription>
        </div>
        <ConnectionAction {...props} />
      </CardHeader>
    </Card>
  )
}

function ConnectionAction(props: ConnectReposItem): JSX.Element {
  const [branch, setBranch] = createSignal('')
  const [branches, setBranches] = createSignal<string[]>([])
  const [isEditing, setIsEditing] = createSignal(false)
  const [editingStarted, setEditingStarted] = createSignal(false)

  function onBranchInput(text: string) {
    setBranch(text)
    setEditingStarted(true)
  }

  function onConnect() {
    setIsEditing(true)
    setBranch('')
  }

  function onConnectConfirm(id: string, branch: string) {
    if (!branches().includes(branch)) {
      setEditingStarted(true)
      return
    }

    reposStore.connectRepo(id, branch)
    setEditingStarted(false)
    setIsEditing(false)
    setBranch('')
  }

  function onDisconnect(id: string) {
    reposStore.connectRepo(id, '')
  }

  createEffect(() => {
    if (isEditing()) {
      reposStore.getBranches(props.fullName).then(setBranches)
    }
  })

  return (
    <Show
      when={props.branch}
      fallback={
        <div class="flex items-center gap-2">
          <Show
            when={isEditing()}
            fallback={
              <Button class="w-28" onClick={onConnect}>
                Connect
              </Button>
            }
          >
            <div class="flex items-center gap-2">
              <Combobox
                options={branches()}
                value={branch()}
                onChange={setBranch}
                placeholder="Branch name"
                validationState={branches().includes(branch()) ? 'valid' : 'invalid'}
                itemComponent={(props) => (
                  <ComboboxItem item={props.item}>
                    <ComboboxItemLabel>{props.item.rawValue}</ComboboxItemLabel>
                    <ComboboxItemIndicator />
                  </ComboboxItem>
                )}
              >
                <ComboboxControl aria-label="Branch">
                  <ComboboxInput
                    value={branch()}
                    onInput={(e) => onBranchInput(e.currentTarget.value)}
                  />
                  <ComboboxTrigger />
                </ComboboxControl>
                <div class="relative">
                  <Show when={editingStarted()}>
                    <ComboboxErrorMessage class="text-destructive absolute">
                      Branch must not be empty
                    </ComboboxErrorMessage>
                  </Show>
                </div>
                <ComboboxContent />
              </Combobox>
              <Button class="w-28" onClick={() => onConnectConfirm(props.id, branch())}>
                Confirm
              </Button>
            </div>
          </Show>
        </div>
      }
    >
      <Button variant="destructive" onClick={() => onDisconnect(props.id)}>
        Disconnect
      </Button>
    </Show>
  )
}

export function ConnectRepos(): JSX.Element {
  onMount(() => {
    reposStore.getRepos()
  })

  return (
    <section class="flex w-full flex-col items-center justify-center py-8">
      <div class="w-full max-w-2xl space-y-6 p-6">
        <h2 class="mb-2 text-2xl font-bold">Connected Repositories</h2>
        <Show
          when={!reposStore.isSyncing}
          fallback={
            <div class="space-y-6">
              <For each={Array(3).fill(0)}>
                {() => (
                  <Card class="mx-auto w-full max-w-2xl">
                    <CardHeader class="flex-row items-center justify-between gap-4 p-6 pb-2">
                      <div class="min-w-0 flex-1 space-y-2">
                        <Skeleton class="h-6 w-3/4" />
                        <Skeleton class="h-4 w-1/2" />
                      </div>
                      <Skeleton class="h-10 w-28" />
                    </CardHeader>
                  </Card>
                )}
              </For>
            </div>
          }
        >
          <For each={reposStore.repos}>
            {(repo) => (
              <RepoItem id={repo.treenqID} fullName={repo.fullName} branch={repo.branch} />
            )}
          </For>
        </Show>
      </div>
    </section>
  )
}
