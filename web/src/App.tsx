import { createAsync, Navigate, Route, Router } from '@solidjs/router'
import type { Component, JSX } from 'solid-js'

import Auth from '@/components/pages/Auth'
import Main from '@/components/pages/Main'
import { Show } from 'solid-js'

import { userStore } from '@/store/userStore'

type ProtectedRouterProps = {
  children: JSX.Element
  satisfies: () => void
  redirectTo: string
}

function ProtectedRouter(props: ProtectedRouterProps): JSX.Element {
  return (
    <Show when={props.satisfies()} fallback={<Navigate href={props.redirectTo} />}>
      {props.children}
    </Show>
  )
}

function MakeProtectedComponent(props: ProtectedRouterProps): Component {
  return function (): JSX.Element {
    return ProtectedRouter(props)
  }
}

function App(): JSX.Element {
  const profile = createAsync(userStore.getProfile)

  return (
    <>
      <Router>
        <Route
          path="/"
          component={MakeProtectedComponent({
            satisfies: () => {
              return profile() ? true : false
            },
            redirectTo: '/auth',
            children: <Main />,
          })}
        />
        <Route
          path="/auth"
          component={MakeProtectedComponent({
            satisfies: () => {
              return profile() ? false : true
            },
            redirectTo: '/',
            children: <Auth />,
          })}
        />
      </Router>
    </>
  )
}

export default App
