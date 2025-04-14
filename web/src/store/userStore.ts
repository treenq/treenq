import { reactive } from 'vue'

export type User = {
  id: string
  displayName: string
  email: string
}

export type UserState = {
  user?: User
  token: string

  isAuthenticated(): boolean
}

export const userStore: UserState = reactive({
  user: undefined,
  token: '',
  isAuthenticated(): boolean {
    return this.token !== '' && this.user !== undefined
  },
})
