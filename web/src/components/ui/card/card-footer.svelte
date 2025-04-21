<script lang="ts">
  import type { WithElementRef } from 'bits-ui'
  import type { HTMLAttributes } from 'svelte/elements'
  import { cn } from '@/components/ui/utils'

  let {
    ref = $bindable(),
    class: className,
    children,
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLDivElement>> = $props()
</script>

<div
  bind:this={ref}
  data-slot="card-footer"
  class={cn('[.border-t]:pt-6 flex items-center px-6', className)}
  {...restProps}
>
  {@render children?.()}
</div>
