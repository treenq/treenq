<script lang="ts">
  let { onClick } = $props()
</script>

<div class="flex min-h-screen items-center justify-center bg-background">
  <div class="flex flex-col items-start space-y-6">
    <h1>Kuku!</h1>
    <button onclick={onClick}></button>
  </div>
</div>
