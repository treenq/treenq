import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import Deploy from '@/components/widgets/Deploy'
import Secrets from '@/components/widgets/Secrets'
import { useParams } from '@solidjs/router'
import { createEffect } from 'solid-js'

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
    <main class="bg-background flex min-h-screen flex-col items-center py-12">
      <div class="flex w-full max-w-3xl flex-col items-center gap-10">
        <Tabs defaultValue="deployments" class="w-full">
          <TabsList class="grid w-full grid-cols-2">
            <TabsTrigger value="deployments">Deployments</TabsTrigger>
            <TabsTrigger value="secrets">Secrets</TabsTrigger>
          </TabsList>
          <TabsContent value="deployments">
            <Deploy repoID={params.id} />
          </TabsContent>
          <TabsContent value="secrets">
            <Secrets repoID={params.id} />
          </TabsContent>
        </Tabs>
      </div>
    </main>
  )
}
