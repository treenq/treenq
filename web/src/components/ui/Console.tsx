import { type BuildProgressMessage } from '@/services/client'
import { cva } from 'class-variance-authority'
import { For, createEffect, onCleanup } from 'solid-js'
import { cn } from './utils'

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
}

export default function Console(props: PropsConsole) {
  let scrollRef: HTMLUListElement | undefined
  let isPaused = false
  let pauseTimeout: number

  const scrollToBottom = () => {
    if (!isPaused && scrollRef) {
      scrollRef.scrollTop = scrollRef.scrollHeight
    }
  }

  createEffect(() => {
    // eslint-disable-next-line @typescript-eslint/no-unused-expressions
    props.logs.length
    scrollToBottom()
  })

  const handleWheel = () => {
    isPaused = true
    clearTimeout(pauseTimeout)
    pauseTimeout = window.setTimeout(() => {
      isPaused = false
    }, 1000)
  }

  onCleanup(() => handleWheel)

  return (
    <ul
      ref={scrollRef}
      class={cn(
        'rounded-radius-xl bg-background border-border h-[400px] overflow-y-auto border border-solid p-4',
        props.classNames,
      )}
      onWheel={handleWheel}
    >
      <For each={props.logs}>
        {(log) => (
          <li class="mb-1 text-sm">
            <span class="text-gray-400">{`[${new Date(log.timestamp).toISOString()}]`}</span>

            <span class={cn(messageVariants({ variant: log.level }))}>{`[${log.level}]`}</span>
            {log.payload}
          </li>
        )}
      </For>
    </ul>
  )
}
