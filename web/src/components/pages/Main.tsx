import { createUserStore } from '@/store/userStore'
import { createAsync } from '@solidjs/router'
import { GithubInstallation } from '../widgets/GithubInstallation'

const userStore = createUserStore()

export default function Main() {
  const profile = createAsync(() => userStore.getProfile())
   //TODO: do something with profile and remove it, here it's just for test that auth works
    console.log(profile)
  
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
