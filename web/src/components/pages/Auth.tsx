import { GitHub } from '@/components/icons/Github'
import { Button } from '@/ui/Button'

export default function Auth() {
  function handleAuth() {
    console.log(import.meta.env.APP_API_HOST)
    window.location.href = `${import.meta.env.APP_API_HOST}/auth?redirectUrl=${import.meta.env.APP_HOST}/authCallback`
  }

  return (
    <div class="bg-background flex min-h-screen items-center justify-center">
      <div class="flex flex-col items-start space-y-6">
        <img src="/logo.png" alt="Logo" width="72" height="72" />

        <h1>Sign In to Treenq</h1>

        <Button
          onClick={handleAuth}
          variant="outline"
          size="xxl"
          class="w-85 flex content-end justify-around text-lg"
        >
          <GitHub />
          Sign In with GitHub
        </Button>
      </div>
    </div>
  )
}
