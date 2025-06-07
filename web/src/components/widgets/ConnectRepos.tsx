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
import { reposStore } from '@/store/repoStore'
import { For, Show, createEffect, createSignal, onMount, type JSX } from 'solid-js'

type ConnectReposItem = {
  id: string
  name: string
  fullName: string
  branch: string
}

function RepoItem(props: ConnectReposItem) {
  const repoHref = `/repos/${props.id}`

  return (
    <Card class="mx-auto w-full max-w-2xl">
      <CardHeader class="flex-row items-center justify-between gap-4 p-6 pb-2">
        <div class="min-w-0 flex-1">
          <CardTitle class="truncate">
            <A variant="light" href={repoHref}>
              {props.fullName}
            </A>
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
      reposStore.getBranches(props.name).then(setBranches)
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
        <For each={reposStore.repos}>
          {(repo) => (
            <RepoItem
              id={repo.treenqID}
              name={repo.fullName.split('/')[1]}
              fullName={repo.fullName}
              branch={repo.branch}
            />
          )}
        </For>
      </div>
    </section>
  )
}
