<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    children?: Snippet;
  }

  let { children }: Props = $props();
</script>

<!-- Simple wrapper for stat items -->
<div class="stats-wrapper rounded-lg border border-[var(--color-base-300)] bg-white">
  {@render children?.()}
</div>
