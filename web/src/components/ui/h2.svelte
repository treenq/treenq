<script lang="ts">
  const { children } = $props()
</script>

<h2
  class="scroll-m-20 border-b pb-2 text-3xl font-semibold tracking-tight transition-colors first:mt-0"
>
  {@render children()}
</h2>
