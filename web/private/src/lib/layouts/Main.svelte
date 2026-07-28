<script lang="ts">
  import type { Snippet } from "svelte";

  let { children }: { children?: Snippet } = $props();
</script>

<div class="flex h-screen w-full items-center justify-center gap-8 px-4 sm:px-6 lg:px-8">
  {@render children?.()}
</div>
