<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    variant?: "ghost" | "primary" | "success" | "warning" | "error" | "info";
    size?: "sm" | "md";
    children?: Snippet;
  }

  let { variant = "ghost", size = "md", children }: Props = $props();
</script>

<span class="badge badge-{variant} badge-{size}">
  {@render children?.()}
</span>

<style>
  /* WS2 chips: soft tinted fill + colored text */
  .badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-weight: var(--font-weight-semibold);
    border: none;
    white-space: nowrap;
  }

  .badge-sm {
    height: 18px;
    padding: 0 var(--space-6);
    font-size: var(--font-size-11);
    line-height: var(--line-height-16);
    border-radius: var(--radius-4);
  }
  .badge-md {
    height: 20px;
    padding: 0 var(--space-8);
    font-size: var(--font-size-12);
    line-height: var(--line-height-16);
    border-radius: var(--radius-6);
  }

  .badge-ghost {
    background: var(--base-hlt-g-selected);
    color: var(--base-txt-secondary);
  }
  .badge-primary {
    background: var(--base-hlt-notr-easy);
    color: var(--base-txt-act-major);
  }
  .badge-success {
    background: var(--color-lime-50);
    color: var(--color-green-750);
  }
  .badge-warning {
    background: var(--color-amber-50);
    color: var(--base-txt-notice-major);
  }
  .badge-error {
    background: var(--base-hlt-w-hover);
    color: var(--base-txt-alert-major);
  }
  .badge-info {
    background: var(--base-hlt-notr-hover);
    color: var(--base-txt-link-minor);
  }
</style>
