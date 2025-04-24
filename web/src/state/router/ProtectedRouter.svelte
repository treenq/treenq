<script lang="ts">
  let { component, satisfies, fallback } = $props()
</script>

{#if satisfies}
  {@render component()}
{:else}
  {@render fallback()}
{/if}
