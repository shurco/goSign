<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    required?: boolean;
    children?: Snippet;
  }

  let { required = false, children }: Props = $props();

  const labelClasses = "block text-sm font-medium text-gray-700 mb-1";
</script>

<label class={labelClasses}>
  {@render children?.()}
  {#if required}
    <span class="ml-1 text-[var(--color-error)]">*</span>
  {/if}
</label>
