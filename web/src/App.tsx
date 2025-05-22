import { Navigate, Route, Router } from '@solidjs/router'
import type { Component, JSX } from 'solid-js'

import Auth from '@/components/pages/Auth'
import Main from '@/components/pages/Main'
import { onMount, Show } from 'solid-js'

import DeployPage from '@/components/pages/DeployPage'
import DeploymentDetailsPage from '@/components/pages/DeploymentDetailsPage' // Import the new page
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

  const isAuthenticated = () => !!userStore.user
  const isNotAuthenticated = () => !userStore.user

  const requiresAuth = (component: () => JSX.Element) =>
    MakeProtectedComponent({
      satisfies: isAuthenticated,
      redirectTo: '/auth',
      component: component,
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
              <Route path="/" component={requiresAuth(Main)} />
              <Route path="/repos/:id" component={requiresAuth(RepoPage)} />
              {/* The existing /deploy/:id route points to DeployPage. */}
              {/* The new route should be /deploy/:deployID and point to DeploymentDetailsPage. */}
              {/* These are distinct if 'id' and 'deployID' are treated as different param names, or if one is more specific. */}
              {/* Given the task, we will add the new route as specified. If there's an overlap, it's usually the first match that wins. */}
              <Route path="/deploy/:id" component={requiresAuth(DeployPage)} /> {/* This page fetches its own data if state is missing */}
              {/* <Route path="/deploy/:deployID" component={requiresAuth(DeploymentDetailsPage)} /> REMOVED old route for DeploymentDetailsPage */}
              <Route path="/deployment-view/:deployID" component={requiresAuth(DeploymentDetailsPage)} /> {/* NEW route for state-driven details page */}
            </Route>
            <Route
              path="/auth"
              component={MakeProtectedComponent({
                satisfies: isNotAuthenticated,
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
