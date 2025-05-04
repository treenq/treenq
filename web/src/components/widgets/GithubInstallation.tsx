import { Button } from '@/components/ui/Button'

const installationLink = `https://github.com/apps/${import.meta.env.APP_GITHUB_APP_NAME}/installations/select_target`

export function GithubInstallation() {
  function onClick() {
    window.open(installationLink, '_blank', 'width=400,height=500')
  }
  return (
    <div class="flex items-center justify-center">
      <Button onclick={onClick}>Integrate Github</Button>
    </div>
  )
}
