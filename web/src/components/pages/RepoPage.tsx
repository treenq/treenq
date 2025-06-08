import { Button } from '@/components/ui/Button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/Tabs'
import Deploy from '@/components/widgets/Deploy'
import Secrets from '@/components/widgets/Secrets'
import { useNavigate } from '@solidjs/router'

import { Routes } from '@/routes'

export default function RepoPage() {
  const reposRoute = Routes.repos
  const repoID = reposRoute.params().id
  const nav = useNavigate()

  // get installation
  // no installation ? offer an installation button
  // has installation ? show list of available repositories
  // implement a connect button
  // show list of connected repositories

  return (
    <main class="bg-background flex min-h-screen w-full flex-col px-8 py-12">
      <div class="mb-6">
        <Button onClick={() => nav(-1)} textContent="Back" variant="outline"></Button>
      </div>
      <div class="flex w-full flex-col gap-10">
        <Tabs defaultValue="deployments" class="w-full">
          <TabsList class="grid w-full grid-cols-2">
            <TabsTrigger value="deployments">Deployments</TabsTrigger>
            <TabsTrigger value="secrets">Secrets</TabsTrigger>
          </TabsList>
          <TabsContent value="deployments">
            <Deploy repoID={repoID} />
          </TabsContent>
          <TabsContent value="secrets">
            <Secrets repoID={repoID} />
          </TabsContent>
        </Tabs>
      </div>
    </main>
  )
}
