import { cn } from '@/components/ui/utils'
import { JSX } from 'solid-js'

export type SpriteIconProps = {
  name: string
  size?: number
  class?: string
}

export function SpriteIcon(props: SpriteIconProps): JSX.Element {
  const size = props.size || 24

  return (
    <svg
      viewBox="0 0 24 24"
      class={cn('fill-none stroke-current stroke-2', props.class)}
      style={{ width: String(size), height: String(size) }}
    >
      <use href={`/static/icon-sprite.svg#${props.name}`} />
    </svg>
  )
}
