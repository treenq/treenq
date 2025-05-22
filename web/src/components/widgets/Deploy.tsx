import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { Separator } from '@/components/ui/Separator'
import { deployStore } from '@/store/deployStore'
import { createSignal, For, Show } from 'solid-js' // Added Show
import { useNavigate } from '@solidjs/router';

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

export default function Deploy(props: DeployProps) {
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
  ]);
  const [isLoadingDeployment, setIsLoadingDeployment] = createSignal(false); // Added back
  const navigate = useNavigate();

  const deploy = async () => {
    setIsLoadingDeployment(true); // Set loading true
    try {
      const deployID = await deployStore.deploy(props.repoID);
      if (deployID && deployStore.recentDeploy) {
        navigate(`/deployment-view/${deployStore.recentDeploy.id}`, { // Use specified path
          state: {
            id: deployStore.recentDeploy.id,
            status: deployStore.recentDeploy.status,
            createdAt: deployStore.recentDeploy.createdAt, // This is a string from deployStore
          },
        });
      } else {
        console.error("Deployment failed or data not available from store.");
        // Consider user feedback for failure, e.g., a toast notification
      }
    } catch (error) {
      console.error("Deployment failed:", error);
      // User feedback
    } finally {
      setIsLoadingDeployment(false); // Set loading false
    }
  };

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
          <Button variant="outline" class="hover:bg-primary" onclick={deploy}>
            Deploy Now
            <div class="bg-muted ml-2 h-4 w-4 rounded" />
          </Button>
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
          {/* Skeleton Loader ADDED BACK */}
          <Show when={isLoadingDeployment()}>
            <div class="flex items-start gap-4 p-6 animate-pulse">
              <div class="bg-muted flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full">
                <div class="bg-muted-foreground h-4 w-4" />
              </div>
              <div class="flex-1">
                <div class="flex items-center justify-between">
                  <div>
                    <div class="bg-muted-foreground mb-2 h-4 w-1/2 rounded"></div> {/* Title */}
                    <div class="bg-muted-foreground mb-2 h-3 w-1/3 rounded"></div> {/* Sub-info */}
                    <div class="bg-muted-foreground h-3 w-1/4 rounded"></div>      {/* Timestamp */}
                  </div>
                </div>
              </div>
            </div>
            {/* Separator is optional, but often good for consistency if items have separators */}
            <Separator /> 
          </Show>
          
          {/* Existing Deployments List */}
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
