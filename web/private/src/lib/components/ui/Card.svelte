<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    header?: Snippet;
    footer?: Snippet;
    children?: Snippet;
  }

  let { header, footer, children }: Props = $props();
</script>

<!-- Card panel (twing-m .table-card): white, r12, tertiary border -->
<div class="card">
  {#if header}
    <div class="card-header">
      {@render header()}
    </div>
  {/if}
  <div class="card-body">
    {@render children?.()}
  </div>
  {#if footer}
    <div class="card-footer">
      {@render footer()}
    </div>
  {/if}
</div>

<style>
  .card {
    background: var(--base-cont-top);
    border: 1px solid var(--base-line-tertiary);
    border-radius: var(--radius-12);
  }
  .card-header {
    padding: var(--space-16) var(--space-20);
    border-bottom: 1px solid var(--base-line-ghost);
  }
  .card-body {
    padding: var(--space-20);
  }
  .card-footer {
    padding: var(--space-12) var(--space-20);
    border-top: 1px solid var(--base-line-ghost);
    background: var(--base-cont-mid);
    border-radius: 0 0 var(--radius-12) var(--radius-12);
  }
</style>
