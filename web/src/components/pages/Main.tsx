import { ConnectRepos } from '@/components/widgets/ConnectRepos'
import { GithubInstallation } from '@/components/widgets/GithubInstallation'

export default function Main() {
  // get installation
  // no installation ? offer an installation button
  // has installation ? show list of available repositories
  // implement a connect button
  // show list of connected repositories

  return (
    <main class="bg-background min-h-screen flex flex-col items-center justify-center py-12">
      <div class="w-full max-w-3xl flex flex-col gap-10 items-center">
        <GithubInstallation />
        <ConnectRepos />
      </div>
    </main>
  )
}
