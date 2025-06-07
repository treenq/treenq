import { ROUTES } from '@/routes'
import { httpClient, type UserInfo } from '@/services/client'
import { redirect } from '@solidjs/router'

import { mergeProps } from 'solid-js'
import { createStore } from 'solid-js/store'

type UserState = {
  user?: UserInfo
}

const newDefaultAuthState = (): UserState => ({ user: undefined })

function createUserStore() {
  const [store, setStore] = createStore(newDefaultAuthState())

  const getProfile = async () => {
    const res = await httpClient.getProfile()
    if ('error' in res) throw redirect(ROUTES.auth)
    setStore({ user: res.data.userInfo })

    return res.data.userInfo
  }

  return mergeProps(store, {
    logout: async () => {
      setStore(newDefaultAuthState())

      await httpClient.logout()
    },
    getProfile: getProfile,
  })
}

export const userStore = createUserStore()
