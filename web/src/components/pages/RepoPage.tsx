import { useParams } from '@solidjs/router'
import { createEffect } from 'solid-js'
import Deploy from '../widgets/Deploy'

export default function RepoPage() {
  const params = useParams()
  createEffect(() => {
    console.log(params.id)
  })
  // get installation
  // no installation ? offer an installation button
  // has installation ? show list of available repositories
  // implement a connect button
  // show list of connected repositories

  return (
    <main class="bg-background flex min-h-screen flex-col items-center justify-center py-12">
      <div class="flex w-full max-w-3xl flex-col items-center gap-10">
        <Deploy repoID={params.id} />
      </div>
    </main>
  )
}
