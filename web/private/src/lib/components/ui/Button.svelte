<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLButtonAttributes } from "svelte/elements";
  import LoadingSpinner from "./LoadingSpinner.svelte";

  interface Props extends HTMLButtonAttributes {
    variant?: "primary" | "ghost" | "outline" | "success" | "warning" | "error" | "info";
    size?: "sm" | "md" | "lg";
    loading?: boolean;
    disabled?: boolean;
    circle?: boolean;
    children?: Snippet;
  }

  let {
    variant = "primary",
    size = "md",
    loading = false,
    disabled = false,
    circle = false,
    class: className = "",
    children,
    ...rest
  }: Props = $props();

  const spinnerSize = $derived(({ sm: "sm", md: "sm", lg: "md" } as const)[size]);
</script>

<button
  class="btn btn-{variant} btn-{size} {className}"
  class:btn-circle={circle}
  disabled={disabled || loading}
  {...rest}
>
  {#if loading}
    <LoadingSpinner size={spinnerSize} />
  {/if}
  {@render children?.()}
</button>

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-6);
    border: 1px solid transparent;
    font-family: inherit;
    font-weight: var(--font-weight-semibold);
    white-space: nowrap;
    cursor: pointer;
    user-select: none;
    text-decoration: none;
    transition: var(--transition-colors), var(--transition-shadow);
  }
  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    pointer-events: none;
  }
  .btn:focus-visible {
    outline: none;
    box-shadow: var(--shadow-focus);
  }

  /* sizes — twing-m: md 34px, sm 28px, lg 40px, r8 */
  .btn-sm {
    height: var(--size-control-28);
    padding: 0 10px;
    font-size: var(--font-size-12);
    border-radius: var(--radius-8);
  }
  .btn-md {
    height: 34px;
    padding: 0 14px;
    font-size: var(--font-size-13);
    border-radius: var(--radius-8);
  }
  .btn-lg {
    height: var(--size-control-40);
    padding: 0 18px;
    font-size: var(--font-size-14);
    border-radius: var(--radius-8);
  }

  .btn-circle {
    padding: 0;
    border-radius: var(--radius-full);
  }
  .btn-circle.btn-sm {
    width: var(--size-control-28);
  }
  .btn-circle.btn-md {
    width: 34px;
  }
  .btn-circle.btn-lg {
    width: var(--size-control-40);
  }

  /* primary — twing-m solid blue */
  .btn-primary {
    background: var(--base-btn-primary-minor);
    color: var(--base-txt-btn-flip);
    box-shadow: var(--shadow-cont-minor);
  }
  .btn-primary:hover:not(:disabled) {
    background: var(--base-btn-primary-major);
  }

  /* ghost — twing-m secondary: white with border (primary secondary button) */
  .btn-ghost {
    background: var(--base-cont-top);
    color: var(--base-txt-primary);
    box-shadow: 0 0 0 1px var(--base-btn-brd-minor) inset;
  }
  .btn-ghost:hover:not(:disabled) {
    background: var(--base-hlt-g-easy);
    box-shadow: 0 0 0 1px var(--base-line-accent) inset;
  }

  /* outline — transparent with border */
  .btn-outline {
    background: transparent;
    color: var(--base-txt-secondary);
    box-shadow: 0 0 0 1px var(--base-btn-brd-minor) inset;
  }
  .btn-outline:hover:not(:disabled) {
    background: var(--base-hlt-g-hover);
    color: var(--base-txt-primary);
  }

  /* success — WS2 fresh */
  .btn-success {
    background: var(--base-btn-fresh-minor);
    color: var(--base-txt-btn-flip);
  }
  .btn-success:hover:not(:disabled) {
    background: var(--base-btn-fresh-major);
  }

  /* warning — WS2 notice */
  .btn-warning {
    background: var(--base-btn-notice-minor);
    color: var(--base-txt-primary);
  }
  .btn-warning:hover:not(:disabled) {
    background: var(--base-btn-notice-major);
    color: var(--base-txt-btn-flip);
  }

  /* error — twing-m solid red */
  .btn-error {
    background: var(--base-btn-alert-minor);
    color: var(--base-txt-btn-flip);
  }
  .btn-error:hover:not(:disabled) {
    background: var(--base-btn-alert-major);
  }

  /* info — soft blue fill */
  .btn-info {
    background: var(--base-hlt-notr-easy);
    color: var(--base-txt-act-major);
  }
  .btn-info:hover:not(:disabled) {
    background: var(--base-hlt-notr-hover);
  }
</style>
