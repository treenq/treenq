import { JSX } from 'solid-js'
import { DefaultSize, IconProps } from './props'

export function LayoutGrid(props: IconProps): JSX.Element {
  const size = props.size || DefaultSize
  return (
    <svg
      viewBox="0 0 24 24"
      class="fill-foreground"
      style={{ width: String(size), height: String(size) }}
    >
      <path stroke="none" d="M0 0h24v24H0z" fill="none" />
      <path d="M9 3a2 2 0 0 1 2 2v4a2 2 0 0 1 -2 2h-4a2 2 0 0 1 -2 -2v-4a2 2 0 0 1 2 -2z" />
      <path d="M19 3a2 2 0 0 1 2 2v4a2 2 0 0 1 -2 2h-4a2 2 0 0 1 -2 -2v-4a2 2 0 0 1 2 -2z" />
      <path d="M9 13a2 2 0 0 1 2 2v4a2 2 0 0 1 -2 2h-4a2 2 0 0 1 -2 -2v-4a2 2 0 0 1 2 -2z" />
      <path d="M19 13a2 2 0 0 1 2 2v4a2 2 0 0 1 -2 2h-4a2 2 0 0 1 -2 -2v-4a2 2 0 0 1 2 -2z" />{' '}
    </svg>
  )
}
