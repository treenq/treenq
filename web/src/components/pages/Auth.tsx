import { GitHub } from '@/components/icons'
import { Button } from '@/components/ui/Button'
import { ROUTES } from '@/utils/constants/routes'

export default function Auth() {
  function handleAuth() {
    window.location.href = `${import.meta.env.APP_API_HOST}${ROUTES.auth}`
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
