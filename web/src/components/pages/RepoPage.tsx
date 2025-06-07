import { Button } from '@/components/ui/Button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import Deploy from '@/components/widgets/Deploy'
import Secrets from '@/components/widgets/Secrets'
import { useSolidRoute } from '@/hooks/useSolidRoutre'

import { createEffect } from 'solid-js'

export default function RepoPage() {
  const { id, backPage } = useSolidRoute()

  createEffect(() => {
    console.log(id)
  })

  // get installation
  // no installation ? offer an installation button
  // has installation ? show list of available repositories
  // implement a connect button
  // show list of connected repositories

  return (
    <main class="bg-background flex min-h-screen w-full flex-col px-8 py-12">
      <div class="mb-6">
        <Button onClick={() => backPage()} textContent="Back" variant="outline"></Button>
      </div>
      <div class="flex w-full flex-col gap-10">
        <Tabs defaultValue="deployments" class="w-full">
          <TabsList class="grid w-full grid-cols-2">
            <TabsTrigger value="deployments">Deployments</TabsTrigger>
            <TabsTrigger value="secrets">Secrets</TabsTrigger>
          </TabsList>
          <TabsContent value="deployments">
            <Deploy repoID={id} />
          </TabsContent>
          <TabsContent value="secrets">
            <Secrets repoID={id} />
          </TabsContent>
        </Tabs>
      </div>
    </main>
  )
}
