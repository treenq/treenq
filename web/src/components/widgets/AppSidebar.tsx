import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/Sidebar'
import { A } from '@solidjs/router'
import { FontAwesomeIcon } from 'solid-fontawesome'
import { For } from 'solid-js'

const items = [
  { title: 'Home', icon: () => <FontAwesomeIcon icon="fa-home" />, url: '/' },
  {
    title: 'Manage Repositories',
    icon: () => <FontAwesomeIcon icon="fab fa-github" />,
    url: '/repositories',
  },
]

export const AppSidebar = () => (
  <Sidebar>
    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>App</SidebarGroupLabel>
        <SidebarGroupContent>
          <SidebarMenu>
            <For each={items}>
              {(item) => (
                <SidebarMenuItem>
                  <SidebarMenuButton as={A} href={item.url}>
                    <item.icon />
                    <span>{item.title}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )}
            </For>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>
  </Sidebar>
)
