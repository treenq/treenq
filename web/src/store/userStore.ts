import { HttpClient, type UserInfo } from '@/services/client'
import { redirect } from '@solidjs/router'

import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type UserState = {
  user?: UserInfo
}

const newDefaultAuthState = (): UserState => ({ user: undefined })

function createUserStore() {
  const client = new HttpClient(import.meta.env.APP_API_HOST)

  const [store, setStore] = createStore(newDefaultAuthState())

  const getProfile = async () => {
    const res = await client.getProfile()
    if ('error' in res) throw redirect('/auth')
    setStore({ user: res.data.userInfo })
    return res.data.userInfo
  }

  return mergeProps(store, {
    logout: async () => {
      setStore(newDefaultAuthState())

      await client.logout()
    },
    getProfile: getProfile,
  })
}

export const userStore = createUserStore()
