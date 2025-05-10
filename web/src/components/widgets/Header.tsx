import { Button } from '@/components/ui/Button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/Popover'
import { userStore } from '@/store/userStore'
import { Show } from 'solid-js'

export function Header() {
  return (
    <header class="bg-background">
      <div class="flex h-16 items-center justify-between sm:px-6 lg:px-8">
        <img src="/logo.png" alt="Logo" width="48" height="48" />

        <div class="flex">
          <Show when={userStore.user}>
            <Popover>
              <PopoverTrigger>
                <Button size="lg" variant="default">
                  {userStore.user?.displayName}
                </Button>
              </PopoverTrigger>
              <PopoverContent>
                <Button variant="secondary" onClick={userStore.logout}>
                  Logout
                </Button>
              </PopoverContent>
            </Popover>
          </Show>
        </div>
      </div>
      <hr />
    </header>
  )
}
