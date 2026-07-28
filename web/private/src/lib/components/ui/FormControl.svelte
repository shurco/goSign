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

<div class="form-control">
  {#if label}
    <Label {required}>{label}</Label>
  {/if}
  {@render children?.()}
  {#if error}
    <p class="mt-1 text-sm text-[var(--color-error)]">{error}</p>
  {:else if hint}
    <p class="mt-1 text-sm text-gray-500">{hint}</p>
  {/if}
</div>
