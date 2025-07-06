import { Button } from '@/components/ui/Button'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/Tooltip'
import { reposStore } from '@/store/repoStore'
import { Show } from 'solid-js'

const installationLink = `https://github.com/apps/${import.meta.env.APP_GITHUB_APP_NAME}/installations/select_target`

type PopupOptions = {
  url: string
  title: string
  w: number
  h: number
}

const createPopup = ({ url, title, w, h }: PopupOptions) => {
  const dualScreenLeft =
    window.screenLeft || // Most browsers
    window.screenX // Firefox
  const dualScreenTop =
    window.screenTop || // Most browsers
    window.screenY // Firefox

  const width = window.innerWidth || document.documentElement.clientWidth || screen.width
  const height = window.innerHeight || document.documentElement.clientHeight || screen.height

  const systemZoom = width / window.screen.availWidth
  const left = (width - w) / 2 / systemZoom + dualScreenLeft
  const top = (height - h) / 2 / systemZoom + dualScreenTop

  window.open(
    url,
    title,
    `
    scrollbars=yes,
    width=${w / systemZoom},
    height=${h / systemZoom},
    top=${top},
    left=${left}
    `,
  )
}

export function GithubInstallation() {
  function onClick() {
    createPopup({ url: installationLink, title: 'Install Treenq', w: 800, h: 600 })
  }

  return (
    <div class="flex items-center justify-center">
      <Show
        when={!reposStore.installation}
        fallback={
          <IntegrateGithubAction
            text="Update Github Credentials"
            variant="outline"
            onClick={onClick}
          />
        }
      >
        <IntegrateGithubAction text="Integrate GitHub" variant="default" onClick={onClick} />
      </Show>
    </div>
  )
}

type GithubAppActionProps = {
  text: string
  variant: 'default' | 'outline'
  onClick: () => void
}

function IntegrateGithubAction(props: GithubAppActionProps) {
  return (
    <>
      <Button variant={props.variant} onclick={props.onClick}>
        {props.text}
      </Button>
      <SyncGithubAppAction />
    </>
  )
}

function SyncGithubAppAction() {
  function syncGithubApp() {
    reposStore.syncGithubApp()
  }

  return (
    <Tooltip>
      <TooltipTrigger as={Button} variant="outline" onClick={syncGithubApp} disabled={reposStore.isSyncing}>
        {reposStore.isSyncing ? 'Syncing...' : 'Sync Github Repos'}
      </TooltipTrigger>
      <TooltipContent>
        <p>App installation webhook may fail, you can sync it manually</p>
      </TooltipContent>
    </Tooltip>
  )
}
