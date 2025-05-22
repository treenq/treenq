import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { Separator } from '@/components/ui/Separator'
import { deployStore } from '@/store/deployStore'
import { createSignal, For } from 'solid-js'
import { useNavigate } from 'solid-js/router'

interface DeploymentProps {
  id: string
  hash: string
  description: string
  status: 'success' | 'failed' | 'started' | 'first'
  timestamp: string
  additionalInfo?: string
  canRollback?: boolean
}

type DeployProps = {
  repoID: string
}

const SkeletonLoading = () => <div class="text-center p-4">Loading deployment details...</div>

export default function Deploy(props: DeployProps) {
  const navigate = useNavigate()
  const [isLoading, setIsLoading] = createSignal(false)
  const [deployments] = createSignal<DeploymentProps[]>([
    {
      id: '1',
      hash: '7144e6c',
      description: 'feat(web): add router, signIn page, fix layout',
      status: 'success',
      timestamp: 'March 28, 2025 at 10:01 AM',
      canRollback: false,
    },
    {
      id: '2',
      hash: '7144e6c',
      description: 'feat(web): add router, signIn page, fix layout',
      status: 'started',
      timestamp: 'March 28, 2025 at 10:01 AM',
      additionalInfo: 'New commit via Auto-Deploy',
    },
    {
      id: '3',
      hash: 'f47c499',
      description: 'fix sed command in Makefile for cross platform compatibility (#51)',
      status: 'success',
      timestamp: 'March 26, 2025 at 11:20 PM',
      canRollback: true,
    },
    {
      id: '4',
      hash: 'f47c499',
      description: 'fix sed command in Makefile for cross platform compatibility (#51)',
      status: 'started',
      timestamp: 'March 26, 2025 at 11:19 PM',
      additionalInfo: 'Build command updated',
    },
    {
      id: '5',
      hash: 'f47c499',
      description: 'fix sed command in Makefile for cross platform compatibility (#51)',
      status: 'failed',
      timestamp: 'March 26, 2025 at 11:18 PM',
      additionalInfo:
        'Exited with status 254 while building your code. Check your deploy logs for more information.',
      canRollback: true,
    },
    {
      id: '6',
      hash: 'f47c499',
      description: 'fix sed command in Makefile for cross platform compatibility (#51)',
      status: 'first',
      timestamp: 'March 26, 2025 at 11:17 PM',
    },
  ])

  const deploy = async () => {
    setIsLoading(true)
    try {
      // Assume deployStore.deploy returns { deploymentID: string, status: string, createdAt: string }
      const { deploymentID, status, createdAt } = await deployStore.deploy(props.repoID)
      if (deploymentID) {
        navigate(`/deploy/${deploymentID}`, { state: { status, createdAt, deployID: deploymentID } })
      } else {
        // Handle the case where deploymentID might be missing, though the task assumes it's returned
        console.error('Deployment failed or did not return a deployment ID.')
      }
    } catch (error) {
      console.error('Error deploying:', error)
      // Potentially show an error message to the user
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div class="mx-auto flex w-full max-w-5xl flex-col">
      <div class="bg-background text-foreground p-6">
        <div class="mb-4 flex items-center justify-between">
          <div>
            <div class="text-muted-foreground mb-1 flex items-center gap-2 text-sm">
              <span class="inline-flex items-center gap-1">
                <div class="bg-muted h-4 w-4 rounded" />
                SERVICE
              </span>
            </div>
            <h1 class="text-3xl font-bold">treenq</h1>
          </div>
          {isLoading() ? (
            <SkeletonLoading />
          ) : (
            <Button variant="outline" class="hover:bg-primary" onclick={deploy}>
              Deploy Now
              <div class="bg-muted ml-2 h-4 w-4 rounded" />
            </Button>
          )}
        </div>

        <div class="mb-2 flex items-center gap-4">
          <div class="flex items-center gap-2">
            <div class="bg-muted h-4 w-4 rounded-full" />
            <span class="text-sm">treenq / treenq</span>
          </div>
          <div class="flex items-center gap-2 text-sm">
            <div class="bg-muted h-4 w-4 rounded-full" />
            <span>main</span>
          </div>
        </div>

        <div class="text-muted-foreground flex items-center gap-2 text-sm">
          <span>treenq.onrender.com</span>
          <div class="bg-muted h-4 w-4 rounded" />
        </div>
      </div>

      <Card class="rounded-none border-0">
        <CardContent class="p-0">
          <For each={deployments()}>
            {(deployment, index) => (
              <div>
                <div class="flex items-start gap-4 p-6">
                  <div class="bg-muted flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full">
                    <div class="bg-muted-foreground h-4 w-4" />
                  </div>

                  <div class="flex-1">
                    <div class="flex items-center justify-between">
                      <div>
                        <p class="text-base font-medium">
                          {deployment.status === 'first'
                            ? 'First deploy'
                            : `Deploy ${deployment.status === 'success' ? 'live' : deployment.status}`}{' '}
                          for{' '}
                          <a href="#" class="text-primary hover:underline">
                            {deployment.hash}
                          </a>
                          : {deployment.description}
                        </p>
                        {deployment.additionalInfo && (
                          <p class="text-muted-foreground mt-1 text-sm">
                            {deployment.additionalInfo}
                          </p>
                        )}
                        <p class="text-muted-foreground mt-1 text-sm">{deployment.timestamp}</p>
                      </div>

                      {deployment.canRollback && (
                        <Button variant="outline" size="sm" class="gap-1">
                          <div class="bg-muted h-4 w-4 rounded" />
                          Rollback
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
                {index() < deployments().length - 1 && <Separator />}
              </div>
            )}
          </For>
        </CardContent>
      </Card>
    </div>
  )
}
