<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    zebra?: boolean;
    header?: Snippet;
    footer?: Snippet;
    children?: Snippet;
  }

  let { zebra = false, header, footer, children }: Props = $props();
</script>

<div class="w-full overflow-x-auto">
  <table class="data-grid">
    {#if header}
      <thead>
        {@render header()}
      </thead>
    {/if}
    <tbody class={zebra ? "zebra" : ""}>
      {@render children?.()}
    </tbody>
    {#if footer}
      <tfoot>
        {@render footer()}
      </tfoot>
    {/if}
  </table>
</div>

<style>
  /* Base table look — global table.data-grid (app.css) */
  tbody.zebra :global(> tr:nth-child(even) td) {
    background-color: var(--base-cont-mid);
  }
  tbody.zebra :global(> tr:hover td) {
    background-color: var(--base-hlt-g-easy);
  }
</style>
