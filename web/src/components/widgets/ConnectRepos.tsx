import { Button } from '@/components/ui/Button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/Card'
import { TextFieldInput } from '@/components/ui/Input'
import { For, JSX, Show, createSignal } from 'solid-js'

type ConnectReposItem = {
  name: string
  branch: string
}

function RepoItem(props: { repo: ConnectReposItem }) {
  const [branch, setBranch] = createSignal('')
  const [isEditing, setIsEditing] = createSignal(false)

  function onConnect(id: string, branch: string) {
    console.log(id, branch)
  }
  function onDisconnect(id: string) {
    console.log(id)
  }

  function handleConnect() {
    onConnect(props.repo.name, branch())
    setIsEditing(false)
    setBranch('')
  }

  return (
    <Card class="border-muted-foreground/20 mx-auto w-full max-w-2xl transition-shadow focus-within:shadow-lg hover:shadow-lg">
      <CardHeader class="flex-row items-center justify-between gap-4 p-6 pb-2">
        <div class="min-w-0 flex-1">
          <CardTitle class="truncate">{props.repo.name}</CardTitle>
          <CardDescription>
            <span class="inline-flex items-center gap-1">
              {' '}
              <Show when={props.repo.branch}>
                ✅ branch: <b>{props.repo.branch}</b>
              </Show>
            </span>
          </CardDescription>
        </div>
        <Show
          when={props.repo.branch}
          fallback={
            <div class="flex items-center gap-2">
              <Show
                when={isEditing()}
                fallback={
                  <Button onClick={() => setIsEditing(true)} size="sm">
                    Connect
                  </Button>
                }
              >
                <TextFieldInput
                  placeholder="Branch name"
                  value={branch()}
                  onInput={(e) => setBranch(e.currentTarget.value)}
                  class="w-40"
                />
                <Button onClick={handleConnect} size="sm">
                  Confirm
                </Button>
              </Show>
            </div>
          }
        >
          <Button variant="destructive" onClick={() => onDisconnect(props.repo.name)} size="sm">
            Disconnect
          </Button>
        </Show>
      </CardHeader>
    </Card>
  )
}

export function ConnectRepos(): JSX.Element {
  const items = [
    {
      name: 'a connected one',
      branch: 'main',
    },
    {
      name: 'not connected one',
      branch: '',
    },
  ]
  return (
    <section class="flex w-full flex-col items-center justify-center py-8">
      <div class="border-muted w-full max-w-2xl space-y-6 rounded-lg border p-6">
        <h2 class="mb-2 text-2xl font-bold">Connected Repositories</h2>
        <For each={items}>{(repo) => <RepoItem repo={repo} />}</For>
      </div>
    </section>
  )
}
