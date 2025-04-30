import { createUserStore } from '@/store/userStore'
import { createAsync } from '@solidjs/router'

const userStore = createUserStore()

export default function Main() {
  const profile = createAsync(() => userStore.getProfile())

  return (
    <div class="bg-background flex min-h-screen items-center justify-center">
      <div class="flex flex-col items-start space-y-6">
        <h1>Kuku!</h1>
        <p>{profile()?.email}</p>
      </div>
    </div>
  )
}
