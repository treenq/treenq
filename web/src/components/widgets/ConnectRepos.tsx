import { Button } from '@/components/ui/Button'
import { Card, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { TextField, TextFieldErrorMessage, TextFieldInput } from '@/components/ui/Input'
import { reposStore } from '@/store/repoStore'
import { For, Show, createSignal, onMount, type JSX } from 'solid-js'

type ConnectReposItem = {
  id: string
  name: string
  branch: string
}

function RepoItem(props: ConnectReposItem) {
  return (
    <Card class="mx-auto w-full max-w-2xl">
      <CardHeader class="flex-row items-center justify-between gap-4 p-6 pb-2">
        <div class="min-w-0 flex-1">
          <CardTitle class="truncate">{props.name}</CardTitle>
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
    if (branch === '') {
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
              <TextField validationState={branch() === '' ? 'invalid' : 'valid'}>
                <TextFieldInput
                  placeholder="Branch name"
                  value={branch()}
                  onInput={(e) => onBranchInput(e.currentTarget.value)}
                  class="w-40"
                />
                <div class="relative">
                  <Show when={editingStarted()}>
                    <TextFieldErrorMessage class="absolute">
                      Branch must not be empty
                    </TextFieldErrorMessage>
                  </Show>
                </div>
              </TextField>
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
          {(repo) => <RepoItem id={repo.treenqID} name={repo.fullName} branch={repo.branch} />}
        </For>
      </div>
    </section>
  )
}
