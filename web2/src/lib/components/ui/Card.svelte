<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    header?: Snippet;
    footer?: Snippet;
    children?: Snippet;
  }

  let { header, footer, children }: Props = $props();
</script>

<!-- Simple card component with header, body, and footer slots -->
<div class="rounded-lg border border-[var(--color-base-300)] bg-[var(--color-base-100)]">
  {#if header}
    <div class="border-b border-[var(--color-base-300)] px-6 py-5">
      {@render header()}
    </div>
  {/if}
  <div class="px-6 py-5">
    {@render children?.()}
  </div>
  {#if footer}
    <div class="border-t border-[var(--color-base-300)] bg-[var(--color-base-200)]/30 px-6 py-4">
      {@render footer()}
    </div>
  {/if}
</div>
