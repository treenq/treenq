import { JSX } from 'solid-js'
import { DefaultSize, IconProps } from './props'

export function CreditCard(props: IconProps): JSX.Element {
  const size = props.size || DefaultSize
  return (
    <svg
      viewBox="0 0 24 24"
      class="fill-foreground"
      style={{ width: String(size), height: String(size) }}
    >
      <path stroke="none" d="M0 0h24v24H0z" fill="none" />
      <path d="M22 10v6a4 4 0 0 1 -4 4h-12a4 4 0 0 1 -4 -4v-6h20zm-14.99 4h-.01a1 1 0 1 0 .01 2a1 1 0 0 0 0 -2zm5.99 0h-2a1 1 0 0 0 0 2h2a1 1 0 0 0 0 -2zm5 -10a4 4 0 0 1 4 4h-20a4 4 0 0 1 4 -4h12z" />
    </svg>
  )
}
