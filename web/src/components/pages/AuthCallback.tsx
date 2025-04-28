import { createUserStore, User } from '@/store/userStore'
import { useNavigate, useSearchParams } from '@solidjs/router'
import { createEffect } from 'solid-js'

/**
 * getProfile takes a JWT and retrieves a User model from the given claims
 * @param {string} token is a JWT
 * @returns {User} a user in the claims
 */
function getProfile(token: string): User {
  const [, payload] = token.split('.')
  const decodedPayload = JSON.parse(atob(payload))
  return {
    id: decodedPayload.id,
    email: decodedPayload.email,
    displayName: decodedPayload.displayName,
  }
}

type TokenResponse = {
  accessToken: string
}

export default function AuthCallback() {
  // get query code
  const [search] = useSearchParams()
  const navigate = useNavigate()
  const userStore = createUserStore()

  createEffect(async () => {
    let tokenResp: TokenResponse
    try {
      const resp = await fetch(`${import.meta.env.APP_API_HOST}/authCallback?code=${search.code}`)
      tokenResp = (await resp.json()) as TokenResponse
    } catch (e) {
      console.error(e)
      return
    }

    const user = getProfile(tokenResp.accessToken)
    userStore.login(tokenResp.accessToken, user)
    navigate('/', { replace: true })
  })

  return (
    <div class="bg-background flex min-h-screen items-center justify-center">
      <div class="flex flex-col items-start space-y-6">
        <img src="/logo.png" alt="Logo" width="72" height="72" />

        <h1>Just a sec, we are signing you in</h1>
      </div>
    </div>
  )
}
