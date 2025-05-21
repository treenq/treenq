export type IconProps = {
  size?: number
  class?: string
}

import { cn } from '@/components/ui/utils'
import { JSX } from 'solid-js'

export function Icon(children: JSX.Element): (props: IconProps) => JSX.Element {
  return function IconComponent(props: IconProps): JSX.Element {
    const size = props.size || DefaultSize

    return (
      <svg
        viewBox="0 0 24 24"
        class={cn('fill-foreground', props.class)}
        style={{ width: String(size), height: String(size) }}
      >
        {children}
      </svg>
    )
  }
}

export const DefaultSize = 24
