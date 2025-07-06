import { Button } from '@/components/ui/Button'
import { cn } from '@/components/ui/utils'
import { type BuildProgressMessage } from '@/services/client'
import { cva } from 'class-variance-authority'
import { For, createEffect, createSignal, onCleanup } from 'solid-js'

const MAX_LINES = 50

const messageVariants = cva('mx-1', {
  variants: {
    variant: {
      INFO: 'text-primary',
      DEBUG: 'text-muted-foreground',
      ERROR: 'text-destructive',
    },
  },
  defaultVariants: {
    variant: 'INFO',
  },
})

type PropsConsole = {
  logs: BuildProgressMessage[]
  classNames: string
  emptyStateMessage?: string
  emptyStateDescription?: string
}

export default function Console(props: PropsConsole) {
  const [isExpanded, setIsExpanded] = createSignal(false)
  let scrollRef: HTMLUListElement | undefined
  let isPaused = false
  let pauseTimeout: number

  const scrollToBottom = () => {
    if (!isPaused && scrollRef) {
      scrollRef.scrollTop = scrollRef.scrollHeight
    }
  }

  createEffect(() => {
    //TODO: make it onMount, then autoscroll down when the last line is visible
    // eslint-disable-next-line @typescript-eslint/no-unused-expressions
    props.logs.length
    scrollToBottom()
  })

  const toggleShowFull = () => {
    isPaused = true
    clearTimeout(pauseTimeout)
    pauseTimeout = window.setTimeout(() => {
      isPaused = false
    }, 1000)
  }
  const handleLogsLess = () => {
    setIsExpanded((prev) => !prev)
  }

  onCleanup(() => toggleShowFull)

  return (
    <>
      <ul
        ref={scrollRef}
        class={cn(
          'rounded-radius-xl bg-background border-border h-[400px] overflow-y-auto border border-solid p-4',
          props.classNames,
        )}
        onWheel={toggleShowFull}
      >
        {props.logs.length === 0 ? (
          <div class="flex h-full flex-col items-center justify-center text-center">
            <div class="text-foreground mb-2 text-lg font-medium">
              {props.emptyStateMessage || 'No logs to show'}
            </div>
            {props.emptyStateDescription && (
              <div class="text-muted-foreground max-w-md text-sm">
                {props.emptyStateDescription}
              </div>
            )}
          </div>
        ) : (
          <For each={isExpanded() ? props.logs.slice(-MAX_LINES) : props.logs}>
            {(log) => (
              <li class="mb-1 text-sm">
                <span class="text-gray-400">{`[${new Date(log.timestamp).toISOString()}]`}</span>

                <span class={cn(messageVariants({ variant: log.level }))}>{`[${log.level}]`}</span>
                {log.payload}
              </li>
            )}
          </For>
        )}
      </ul>
      <div class="flex justify-between">
        <div class="text-muted-foreground text-sm">
          {`Showing ${isExpanded() ? `latest ${MAX_LINES} of` : 'all  of'} ${props.logs.length} log lines`}
        </div>
        <Button
          size="sm"
          class="h-6"
          onClick={handleLogsLess}
          textContent={isExpanded() ? 'Show all' : 'Show less'}
        />
      </div>
    </>
  )
}
