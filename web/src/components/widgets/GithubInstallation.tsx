import { Button } from '@/components/ui/Button'
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
        when={reposStore.installation}
        fallback={<Button onclick={onClick}>Integrate Github</Button>}
      >
        <Button variant="outline" onclick={onClick}>
          Update Github Credentials
        </Button>
      </Show>
    </div>
  )
}
