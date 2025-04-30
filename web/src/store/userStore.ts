import { HttpClient, UserInfo } from '@/services/client'
import { redirect } from '@solidjs/router'

import { makePersisted } from '@solid-primitives/storage'
import { createStore } from 'solid-js/store'

type AuthState = {
  user?: UserInfo | undefined
}

const defaultAuthState: AuthState = {
  user: undefined,
}

export function createUserStore() {
  const client = new HttpClient(import.meta.env.APP_API_HOST)

  const [store, setStore] = makePersisted(createStore(defaultAuthState), {
    name: 'tq-auth',
  })

  const login = (user: UserInfo) => {
    setStore({ user: user })
  }

  const getProfile = async () => {
    if (store.user) return store.user

    const res = await client.getProfile()
    if ('error' in res) throw redirect('/auth')
    login(res.data.userInfo)
  }

  return {
    login: login,
    logout: () => setStore(defaultAuthState),
    getProfile: getProfile,
    ...store,
  }
}
