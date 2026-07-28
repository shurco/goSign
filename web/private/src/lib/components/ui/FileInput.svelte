<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";

  interface Props extends Omit<HTMLInputAttributes, "size" | "type"> {
    size?: "sm" | "md" | "lg";
  }

  let { size = "md", ...rest }: Props = $props();
</script>

<input type="file" class="ui-file-input ui-file-input-{size}" {...rest} />

<style>
  .ui-file-input {
    width: 100%;
    color: var(--base-txt-secondary);
    background: var(--base-cont-top);
    border: 1px solid var(--base-line-tertiary);
    border-radius: var(--radius-6);
    cursor: pointer;
    transition: var(--transition-colors);
  }
  .ui-file-input:focus {
    outline: none;
    border-color: var(--base-hlt-invert);
    box-shadow: 0 0 0 2px var(--base-hlt-easy);
  }
  .ui-file-input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .ui-file-input::file-selector-button {
    margin-right: var(--space-12);
    border: none;
    background: var(--base-hlt-g-hover);
    padding: var(--space-6) var(--space-12);
    font-size: var(--font-size-12);
    font-weight: var(--font-weight-medium);
    font-family: inherit;
    color: var(--base-txt-secondary);
    cursor: pointer;
    transition: var(--transition-colors);
  }
  .ui-file-input:hover::file-selector-button {
    background: var(--base-hlt-g-selected);
  }

  .ui-file-input-sm {
    font-size: var(--font-size-12);
  }
  .ui-file-input-md {
    font-size: var(--font-size-13);
  }
  .ui-file-input-lg {
    font-size: var(--font-size-14);
  }
</style>
