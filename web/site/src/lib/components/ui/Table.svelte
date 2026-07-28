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
  <table class="w-full border-collapse">
    {#if header}
      <thead class="border-b border-gray-200 bg-gray-50">
        {@render header()}
      </thead>
    {/if}
    <tbody class={zebra ? "divide-y divide-gray-200" : ""}>
      {@render children?.()}
    </tbody>
    {#if footer}
      <tfoot class="border-t border-gray-200 bg-gray-50">
        {@render footer()}
      </tfoot>
    {/if}
  </table>
</div>

<style>
  tbody.divide-y :global(> tr:nth-child(even)) {
    background-color: #f9fafb;
  }

  tbody :global(> tr:hover) {
    background-color: #f3f4f6;
  }
</style>
