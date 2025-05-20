import { JSX } from 'solid-js'
import { DefaultSize, IconProps } from './props'

export function ChevronRight(props: IconProps): JSX.Element {
  const size = props.size || DefaultSize
  return (
    <svg
      viewBox="0 0 24 24"
      class="fill-foreground"
      style={{ width: String(size), height: String(size) }}
    >
      <path d="M9 6l6 6l-6 6" />
    </svg>
  )
}
