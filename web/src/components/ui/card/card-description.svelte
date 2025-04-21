<script lang="ts">
  import type { WithElementRef } from 'bits-ui'
  import type { HTMLAttributes } from 'svelte/elements'
  import { cn } from '@/components/ui/utils'

  let {
    ref = $bindable(),
    class: className,
    children,
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLParagraphElement>> = $props()
</script>

<p
  bind:this={ref}
  data-slot="card-description"
  class={cn('text-muted-foreground text-sm', className)}
  {...restProps}
>
  {@render children?.()}
</p>
