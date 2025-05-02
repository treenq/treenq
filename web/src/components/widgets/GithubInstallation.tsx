import { Button } from '@/components/ui/Button'

const installationLink = `https://github.com/apps/${import.meta.env.APP_GITHUB_APP_NAME}/installations/select_target`

export function GithubInstallation() {
  //TODO: add github app setup redirect
  function onClick() {
    window.location.href = installationLink
  }
  return (
    <div class="flex items-center justify-center">
      <Button onclick={onClick}>Integrate Github</Button>
    </div>
  )
}
