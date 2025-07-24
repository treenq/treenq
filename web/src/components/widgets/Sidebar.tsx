// File: web/src/components/widgets/Sidebar.tsx
import { createEffect, createSignal, For, Show } from 'solid-js'

import { SpriteIcon } from '@/components/icons/SpriteIcon'

/*
1) fix style (text, colors, position)
2) add A link on href elements

*/
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/Collapsible'
import {
  Sidebar,
  SidebarContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
} from '@/components/ui/Sidebar'
import { reposStore } from '@/store/repoStore'

interface SidebarChild {
  label: string
  href: string
}

interface SidebarItemProps {
  icon: string
  label: string
  isActive?: boolean
  href?: string
  children?: SidebarChild[]
}

export function AppSidebar() {
  const sidebarItemsSkeleton: (SidebarItemProps | { type: 'divider'; label: string })[] = [
    {
      type: 'divider',
      label: 'workspace',
    },
    {
      icon: 'layout-grid',
      label: 'Projects',
      isActive: true,
      href: '/projects',
      children: [],
    },
    {
      type: 'divider',
      label: 'account',
    },
    {
      icon: 'settings',
      label: 'Settings',
      href: '#',
    },
  ]
  const [sidebarItems, setSidebarItems] = createSignal(sidebarItemsSkeleton)

  createEffect(() => {
    const reposList = reposStore.repos.map((it) => ({ label: it.fullName, href: '#' }))
    const updated = sidebarItemsSkeleton.map((item) => {
      if (item.label === 'Projects' && 'children' in item) {
        return { ...item, children: reposList }
      }
      return item
    })
    setSidebarItems(updated)
  })

  return (
    <SidebarProvider>
      <Sidebar>
        <SidebarContent>
          <SidebarMenu>
            <For each={sidebarItems()}>
              {(item) =>
                'type' in item && item.type === 'divider' ? (
                  <div class="text-muted-foreground px-4 py-2 text-sm font-medium">
                    {item.label}
                  </div>
                ) : (
                  <Show
                    when={(item as SidebarItemProps).children}
                    fallback={
                      <SidebarMenuItem>
                        <SidebarMenuButton>
                          <a href={(item as SidebarItemProps).href} class="flex items-center">
                            <SpriteIcon name={(item as SidebarItemProps).icon} />
                            <span>{(item as SidebarItemProps).label}</span>
                          </a>
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    }
                  >
                    <Collapsible class="w-full">
                      <SidebarMenuItem>
                        <CollapsibleTrigger class="group flex w-full">
                          <SidebarMenuButton>
                            <div class="flex items-center">
                              <SpriteIcon name={(item as SidebarItemProps).icon} />
                              <span>{(item as SidebarItemProps).label}</span>
                            </div>
                            <SpriteIcon name="chevron-right" class="group-data-expanded:rotate-90" />
                          </SidebarMenuButton>
                        </CollapsibleTrigger>
                        <CollapsibleContent>
                          <div class="ml-6 mt-1 space-y-1">
                            <For each={(item as SidebarItemProps).children}>
                              {(child) => (
                                <SidebarMenuItem>
                                  <SidebarMenuButton class="hover:bg-sidebar-primary">
                                    <a href={child.href}>{child.label}</a>
                                  </SidebarMenuButton>
                                </SidebarMenuItem>
                              )}
                            </For>
                          </div>
                        </CollapsibleContent>
                      </SidebarMenuItem>
                    </Collapsible>
                  </Show>
                )
              }
            </For>
          </SidebarMenu>
        </SidebarContent>
      </Sidebar>
    </SidebarProvider>
  )
}
