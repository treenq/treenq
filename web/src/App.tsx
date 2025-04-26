import { Navigate, Route, Router } from '@solidjs/router'
import type { Component, JSX } from 'solid-js'
import { ParentProps } from 'solid-js'

import Auth from '@/components/pages/Auth'
import Main from '@/components/pages/Main'
import { Show } from 'solid-js'

type ProtectedRouterProps = {
  satisfies: boolean
  redirectTo: string
} & ParentProps

function ProtectedRouter(props: ProtectedRouterProps) {
  return (
    <Show when={props.satisfies} fallback={<Navigate href={props.redirectTo} />}>
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
  return (
    <>
      <Router>
        <Route
          path="/"
          component={MakeProtectedComponent({
            satisfies: false,
            redirectTo: '/auth',
            children: <Main />,
          })}
        />
        <Route path="/auth" component={Auth} />
      </Router>
    </>
  )
}

export default App
