import { A } from '@/components/ui/A'
import { Button } from '@/components/ui/Button'

export default function NotFound() {
  return (
    <div class="flex h-screen w-screen flex-col items-center justify-center px-4 text-center">
      <h1 class="mb-2 text-3xl font-bold">Page Not Found</h1>
      <p class="text-muted-foreground mb-6">Sorry, we couldn’t find the page you’re looking for.</p>
      <A href="/">
        <Button>Go to Home</Button>
      </A>
    </div>
  )
}
