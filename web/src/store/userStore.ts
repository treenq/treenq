import { HttpClient, UserInfo } from '@/services/client'
import { redirect } from '@solidjs/router'

import { makePersisted } from '@solid-primitives/storage'
import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type UserState = {
  user?: UserInfo
}

type UserStateMutator = {
  logout: () => void
  getProfile: () => Promise<UserInfo>
}

type UserStore = UserState & UserStateMutator

const newDefaultAuthState = (): UserState => ({ user: undefined })

function createUserStore(): UserStore {
  const client = new HttpClient(import.meta.env.APP_API_HOST)

  const [store, setStore] = makePersisted(createStore(newDefaultAuthState()), {
    name: 'tq-auth',
  })

  const getProfile = async () => {
    if (store.user) return store.user

    const res = await client.getProfile()
    if ('error' in res) throw redirect('/auth')
    setStore({ user: res.data.userInfo })
    return res.data.userInfo
  }

  return mergeProps(store, {
    logout: () => setStore(newDefaultAuthState()),
    getProfile: getProfile,
  })
}

export const userStore = createUserStore()
