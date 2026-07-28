<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    required?: boolean;
    children?: Snippet;
  }

  let { required = false, children }: Props = $props();
</script>

<label class="ui-label">
  {@render children?.()}
  {#if required}
    <span class="ui-label-required">*</span>
  {/if}
</label>

<style>
  .ui-label {
    display: block;
    font-size: var(--font-size-12);
    font-weight: var(--font-weight-medium);
    color: var(--base-txt-muted);
    margin-bottom: var(--space-4);
  }
  .ui-label-required {
    margin-left: var(--space-2);
    color: var(--base-txt-alert-minor);
  }
</style>
