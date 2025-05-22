import { Navigate, Route, Router } from '@solidjs/router'
import type { Component, JSX } from 'solid-js'

import Auth from '@/components/pages/Auth'
import Main from '@/components/pages/Main'
import { onMount, Show } from 'solid-js'

import RedirectPage from '@/components/pages/RedirectPage'
import RepoPage from '@/components/pages/RepoPage'
import { Header } from '@/components/widgets/Header'
import { AppSidebar } from '@/components/widgets/Sidebar'
import { userStore } from '@/store/userStore'
import NotFound from './components/pages/NotFound'

type ProtectedRouterProps = {
  component: () => JSX.Element
  satisfies: () => boolean
  redirectTo: string
}

function ProtectedRouter(props: ProtectedRouterProps): JSX.Element {
  return (
    <Show when={props.satisfies()} fallback={<Navigate href={props.redirectTo} />}>
      {props.component()}
    </Show>
  )
}

function MakeProtectedComponent(props: ProtectedRouterProps): Component {
  return function (): JSX.Element {
    return ProtectedRouter(props)
  }
}

function App(): JSX.Element {
  onMount(() => {
    userStore.getProfile()
  })

  return (
    <>
      <Header />
      <div class="flex min-h-screen w-full">
        <Show when={userStore.user}>
          <AppSidebar />
        </Show>
        <div class="min-h-screen flex-1">
          <Router>
            <Route path="/">
              <Route
                path="/"
                component={MakeProtectedComponent({
                  satisfies: () => {
                    return userStore.user ? true : false
                  },
                  redirectTo: '/auth',
                  component: Main,
                })}
              />
              <Route
                path="/repos/:id"
                component={MakeProtectedComponent({
                  satisfies: () => {
                    return userStore.user ? true : false
                  },
                  redirectTo: '/auth',
                  component: RepoPage,
                })}
              />
            </Route>
            <Route
              path="/auth"
              component={MakeProtectedComponent({
                satisfies: () => {
                  return userStore.user ? false : true
                },
                redirectTo: '/',
                component: Auth,
              })}
            />
            <Route path="/githubPostInstall" component={RedirectPage} />
            <Route path="*404" component={NotFound} />
          </Router>
        </div>
      </div>
    </>
  )
}

export default App
