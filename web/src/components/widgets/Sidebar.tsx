// File: web/src/components/widgets/Sidebar.tsx
import { For, Show } from 'solid-js'

import { ChevronRight } from '@/components/icons/ChevronRight'
// import { CreditCard } from '@/components/icons/CreditCard'
import { LayoutGrid } from '@/components/icons/LayoutGrid'
// import { Settings } from '@/components/icons/Settings'

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
import { cn } from '@/components/ui/utils'

interface SidebarChild {
  label: string
  href: string
}

interface SidebarItemProps {
  icon: typeof ChevronRight
  label: string
  isActive?: boolean
  href?: string
  children?: SidebarChild[]
}

const sidebarItems: (SidebarItemProps | { type: 'divider'; label: string })[] = [
  {
    type: 'divider',
    label: 'workspace',
  },
  {
    icon: LayoutGrid,
    label: 'Projects',
    isActive: true,
    href: '/projects',
    children: [
      { label: 'Frontend', href: '/blueprints/frontend' },
      { label: 'Backend', href: '/blueprints/backend' },
      { label: 'Full Stack', href: '/blueprints/full-stack' },
    ],
  },
  // {
  //   type: 'divider',
  //   label: 'account',
  // },
  // {
  //   icon: CreditCard,
  //   label: 'Billing',
  //   href: '/billing',
  // },
  // {
  //   icon: Settings,
  //   label: 'Settings',
  //   href: '/settings',
  // },
]

export function AppSidebar() {
  return (
    <SidebarProvider>
      <Sidebar class="text-sidebar-foreground bg-sidebar border-r">
        <SidebarContent>
          <SidebarMenu>
            <For each={sidebarItems}>
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
                        <SidebarMenuButton
                          class={cn(
                            'hover:bg-sidebar-accent',
                            (item as SidebarItemProps).isActive &&
                              'text-sidebar-primary-foreground bg-sidebar-primary',
                          )}
                        >
                          <a href={(item as SidebarItemProps).href} class="flex items-center">
                            {(item as SidebarItemProps).icon({})}
                            <span>{(item as SidebarItemProps).label}</span>
                          </a>
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    }
                  >
                    <Collapsible class="w-full">
                      <SidebarMenuItem>
                        <CollapsibleTrigger class="flex w-full">
                          <SidebarMenuButton
                            class={cn(
                              'hover:bg-sidebar-accent w-full justify-between',
                              (item as SidebarItemProps).isActive &&
                                'bg-sidebar-primary text-sidebar-primary-foreground',
                            )}
                          >
                            <div class="flex items-center">
                              {(item as SidebarItemProps).icon({})}
                              <span>{(item as SidebarItemProps).label}</span>
                            </div>
                            <ChevronRight />
                          </SidebarMenuButton>
                        </CollapsibleTrigger>
                        <CollapsibleContent>
                          <div class="ml-6 mt-1 space-y-1">
                            <For each={(item as SidebarItemProps).children}>
                              {(child) => (
                                <SidebarMenuItem>
                                  <SidebarMenuButton class="hover:bg-sidebar-accent">
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
