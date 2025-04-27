import { makePersisted } from '@solid-primitives/storage'
import { createStore } from 'solid-js/store'

export type User = {
  id: string
  email: string
  displayName: string
}

type AuthState = {
  isAuthenticated: boolean
  token: string
  user?: User | undefined
}

const defaultAuthState: AuthState = {
  isAuthenticated: false,
  token: '',
  user: undefined,
}

export function createUserStore() {
  const [store, setStore] = makePersisted(createStore<AuthState>(defaultAuthState), {
    name: 'tq-auth',
  })

  return {
    store: store,
    login: (token: string, user: User) => {
      setStore({ token: token, user: user, isAuthenticated: true })
    },
    logout: () => setStore(defaultAuthState),
  }
}
