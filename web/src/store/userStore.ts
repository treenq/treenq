import { HttpClient, type UserInfo } from '@/services/client'
import { redirect } from '@solidjs/router'

import { makePersisted } from '@solid-primitives/storage'
import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type UserState = {
  user?: UserInfo
  expiry: Date
}

const minute = 60 * 1000

const newDefaultAuthState = (): UserState => ({ user: undefined, expiry: new Date() })

function createUserStore() {
  const client = new HttpClient(import.meta.env.APP_API_HOST)

  const [store, setStore] = makePersisted(createStore(newDefaultAuthState()), {
    name: 'tq-auth',
  })

  const getProfile = async () => {
    const now = new Date()
    if (store.user && store.expiry > now) return store.user

    const res = await client.getProfile()
    if ('error' in res) throw redirect('/auth')
    const expiry = new Date(now.getTime() + 5 * minute)
    setStore({ user: res.data.userInfo, expiry })
    return res.data.userInfo
  }

  return mergeProps(store, {
    logout: () => setStore(newDefaultAuthState()),
    getProfile: getProfile,
  })
}

export const userStore = createUserStore()
