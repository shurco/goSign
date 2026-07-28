<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLSelectAttributes } from "svelte/elements";

  interface Props extends Omit<HTMLSelectAttributes, "size" | "value"> {
    value?: string | number;
    error?: boolean;
    size?: "sm" | "md" | "lg";
    children?: Snippet;
  }

  let { value = $bindable(), error = false, size = "md", class: className = "", children, ...rest }: Props = $props();
</script>

<select
  class="ui-select ui-select-{size} {className}"
  class:error
  {value}
  onchange={(e) => (value = e.currentTarget.value)}
  {...rest}
>
  {@render children?.()}
</select>

<style>
  /* WS2 framed select: tertiary border, token chevron, focus — blue border */
  .ui-select {
    width: 100%;
    min-width: 0;
    color: var(--base-txt-primary);
    background-color: var(--base-cont-top);
    background-image: var(--ui-select-chevron);
    background-repeat: no-repeat;
    background-position: right 10px center;
    background-size: 14px 14px;
    border: 1px solid var(--base-line-tertiary);
    border-radius: var(--radius-6);
    appearance: none;
    -webkit-appearance: none;
    outline: none;
    font-family: inherit;
    cursor: pointer;
    transition: var(--transition-colors);
  }
  .ui-select:hover:not(:disabled):not(:focus) {
    border-color: var(--base-line-secondary);
  }
  .ui-select:focus {
    border-color: var(--base-hlt-invert);
    box-shadow: 0 0 0 2px var(--base-hlt-easy);
  }
  .ui-select:disabled {
    color: var(--base-txt-muted);
    cursor: not-allowed;
    opacity: 0.7;
    background-color: var(--base-cont-trans-low);
  }

  .ui-select.error {
    border-color: var(--base-line-alert);
  }
  .ui-select.error:focus {
    box-shadow: 0 0 0 2px var(--color-red-alpha-100);
  }

  .ui-select-sm {
    height: var(--size-control-28);
    padding: 0 24px 0 var(--space-8);
    font-size: var(--font-size-12);
    background-position: right 6px center;
    background-size: 12px 12px;
  }
  .ui-select-md {
    height: var(--size-control-36);
    padding: 0 32px 0 var(--space-12);
    font-size: var(--font-size-13);
  }
  .ui-select-lg {
    height: var(--size-control-44);
    padding: 0 36px 0 var(--space-12);
    font-size: var(--font-size-14);
  }
</style>
