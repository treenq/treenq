import { GithubInstallation } from '../widgets/GithubInstallation'

export default function Main() {
  // get installation
  // no installation ? offer an installation button
  // has installation ? show list of available repositories
  // implement a connect button
  // show list of connected repositories

  return (
    <div class="bg-background flex min-h-screen items-center justify-center">
      <div class="flex flex-col items-start space-y-6">
        <GithubInstallation />
      </div>
    </div>
  )
}
