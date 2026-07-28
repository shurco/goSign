<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";

  interface Props extends Omit<HTMLInputAttributes, "size" | "type" | "value"> {
    value?: string | number;
    type?: "text" | "email" | "password" | "number" | "date" | "tel" | "url" | "search" | "color";
    error?: boolean;
    size?: "sm" | "md" | "lg";
  }

  let {
    value = $bindable(""),
    type = "text",
    error = false,
    size = "md",
    class: className = "",
    ...rest
  }: Props = $props();
</script>

<!-- Manual value sync: Svelte forbids bind:value together with a dynamic `type` -->
<input
  class="ui-input ui-input-{size} {className}"
  class:error
  {type}
  {value}
  oninput={(e) => (value = e.currentTarget.value)}
  {...rest}
/>

<style>
  /* WS2 input: input fill, inset ring border, focus — blue ring */
  .ui-input {
    width: 100%;
    background: var(--base-cont-input);
    color: var(--base-txt-accent);
    border: 1px solid transparent;
    box-shadow: 0 0 0 1px var(--base-line-secondary) inset;
    border-radius: var(--radius-6);
    outline: none;
    font-family: inherit;
    transition: var(--transition-shadow);
  }
  .ui-input::placeholder {
    color: var(--base-txt-ghost);
  }
  .ui-input:hover:not(:focus):not(:disabled) {
    box-shadow: 0 0 0 1px var(--base-line-accent) inset;
  }
  .ui-input:focus {
    box-shadow: var(--shadow-focus);
  }
  .ui-input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    background: var(--base-cont-trans-low);
  }

  .ui-input.error {
    box-shadow: var(--shadow-brd-error);
  }
  .ui-input.error:focus {
    box-shadow:
      0 0 0 1px var(--base-line-alert),
      0 0 0 4px var(--color-red-alpha-200);
  }

  .ui-input-sm {
    height: var(--size-control-28);
    padding: 0 var(--space-8);
    font-size: var(--font-size-12);
  }
  .ui-input-md {
    height: var(--size-control-36);
    padding: 0 var(--space-12);
    font-size: var(--font-size-13);
  }
  .ui-input-lg {
    height: var(--size-control-44);
    padding: 0 var(--space-12);
    font-size: var(--font-size-14);
  }
</style>
