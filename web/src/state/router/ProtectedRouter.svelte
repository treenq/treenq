<script lang="ts">
  import { location } from '@wjfe/n-savant'
  let { children, satisfies } = $props()

  $effect(() => {
    if (!satisfies) location.navigate('/auth', { replace: true })
  })
</script>

{@render children()}
