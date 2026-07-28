<script lang="ts">
  import type { Snippet } from "svelte";
  import Label from "./Label.svelte";

  interface Props {
    label?: string;
    required?: boolean;
    error?: string;
    hint?: string;
    children?: Snippet;
  }

  let { label = "", required = false, error = "", hint = "", children }: Props = $props();
</script>

<div class="ui-form-control">
  {#if label}
    <Label {required}>{label}</Label>
  {/if}
  {@render children?.()}
  {#if error}
    <p class="ui-form-error">{error}</p>
  {:else if hint}
    <p class="ui-form-hint">{hint}</p>
  {/if}
</div>

<style>
  .ui-form-control {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }
  .ui-form-error {
    margin: var(--space-2) 0 0;
    font-size: var(--font-size-12);
    color: var(--base-txt-alert-minor);
  }
  .ui-form-hint {
    margin: var(--space-2) 0 0;
    font-size: var(--font-size-12);
    color: var(--base-txt-muted);
  }
</style>
