import { GitHub } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { Routes } from '@/routes'

export default function Auth() {
  function handleAuth() {
    window.location.replace(`${import.meta.env.APP_API_HOST}${Routes.auth.path}`)
  }

  return (
    <div class="bg-background flex min-h-screen items-center justify-center">
      <div class="flex flex-col items-start space-y-2">
        <img src="/logo.png" alt="Logo" width="48" height="48" />

        <h2>Sign In to Treenq</h2>

        <Button onClick={handleAuth} variant="outline" size="xxl" class="flex w-60 justify-end">
          <GitHub />
          Sign In with GitHub
        </Button>
      </div>
    </div>
  )
}
