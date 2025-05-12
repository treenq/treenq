import { reposStore } from '@/store/repoStore'
import { NavigationMenu } from '@kobalte/core/navigation-menu'
import { createSignal, For } from 'solid-js'

const Sidebar = () => {
  const [isOpen, setIsOpen] = createSignal(false)

  const repos = reposStore.repos.map((repo) => repo.fullName)

  return (
    <>
      <button
        class="bg-accent fixed left-4 top-4 z-10 rounded-full p-2 text-white xl:hidden"
        onClick={() => setIsOpen(!isOpen())}
      >
        - - -
      </button>
      <NavigationMenu
        orientation="vertical"
        class={`${isOpen() ? 'block' : 'hidden'} bg-background fixed top-[64px] h-full w-[250px] border px-2 py-6 xl:absolute xl:block`}
      >
        <NavigationMenu.Menu>
          <NavigationMenu.Trigger as="a" href="/" class="block">
            Main Menu
          </NavigationMenu.Trigger>
        </NavigationMenu.Menu>
        <NavigationMenu.Menu>
          <span class="text-muted-foreground mt-4 block">Repos</span>
          <NavigationMenu.Separator class="my-2" />
          <For each={repos}>
            {(repo) => (
              <NavigationMenu.Trigger as="a" href={`/${repo}`} class="mt-2 block">
                {repo}
              </NavigationMenu.Trigger>
            )}
          </For>
        </NavigationMenu.Menu>
      </NavigationMenu>
    </>
  )
}

export default Sidebar
