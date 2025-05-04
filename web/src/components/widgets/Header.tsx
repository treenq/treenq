import { userStore } from '@/store/userStore'
import { Show } from 'solid-js'

export function Header() {
  return (
    <header class="bg-background">
      <div class="flex h-16 max-w-screen-xl items-center justify-between sm:px-6 lg:px-8">
        <img src="/logo.png" alt="Logo" width="48" height="48" />

        <div class="flex">
          <Show when={userStore.user && true}>
            <span>{userStore.user?.displayName}</span>
          </Show>
        </div>
      </div>
      <hr />
    </header>
  )
}
