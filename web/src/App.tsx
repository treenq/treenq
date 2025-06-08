import { Navigate, Route, Router } from '@solidjs/router'
import type { Component, JSX } from 'solid-js'

import Auth from '@/components/pages/Auth'
import Main from '@/components/pages/Main'
import { createSignal, onMount, Show } from 'solid-js'

import DeployPage from '@/components/pages/DeployPage'
import NotFound from '@/components/pages/NotFound'
import RedirectPage from '@/components/pages/RedirectPage'
import RepoPage from '@/components/pages/RepoPage'
import { Skeleton } from '@/components/ui/Skeleton'
import { Header } from '@/components/widgets/Header'
import { AppSidebar } from '@/components/widgets/Sidebar'
import { Routes } from '@/routes'
import { userStore } from '@/store/userStore'

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
  const [isLoading, setIsLoading] = createSignal(true)

  onMount(() => {
    const fetchUser = async () => {
      try {
        await userStore.getProfile()
      } finally {
        setIsLoading(false)
      }
    }
    fetchUser()
  })

  const isAuthenticated = () => !!userStore.user
  const isNotAuthenticated = () => !userStore.user

  const requiresAuth = (component: () => JSX.Element) =>
    MakeProtectedComponent({
      satisfies: isAuthenticated,
      redirectTo: Routes.auth.path,
      component: component,
    })

  return (
    <>
      <Header />
      <div class="flex min-h-screen w-full">
        <Show when={!isLoading()} fallback={<Skeleton />}>
          <Show when={userStore.user}>
            <AppSidebar />
          </Show>
          <div class="min-h-screen flex-1">
            <Router>
              <Route path="/">
                <Route path="/" component={requiresAuth(Main)} />
                <Route path={Routes.repos.path} component={requiresAuth(RepoPage)} />
                <Route path={Routes.deploy.path} component={requiresAuth(DeployPage)} />
              </Route>
              <Route
                path={Routes.auth.path}
                component={MakeProtectedComponent({
                  satisfies: isNotAuthenticated,
                  redirectTo: '/',
                  component: Auth,
                })}
              />
              <Route path="/githubPostInstall" component={RedirectPage} />
              <Route path={`*${Routes.notFound.path}`} component={NotFound} />
            </Router>
          </div>
        </Show>
      </div>
    </>
  )
}

export default App
