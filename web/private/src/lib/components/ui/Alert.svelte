<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    variant?: "info" | "success" | "warning" | "error";
    class?: string;
    icon?: Snippet;
    children?: Snippet;
  }

  let { variant = "info", class: className = "", icon, children }: Props = $props();
</script>

<div class="alert alert-{variant} {className}" role="alert">
  {#if icon}
    <div class="alert-icon">
      {@render icon()}
    </div>
  {/if}
  <div class="alert-body">
    {@render children?.()}
  </div>
</div>

<style>
  /* WS2 note blocks: soft tinted panel with accent stripe */
  .alert {
    display: flex;
    align-items: flex-start;
    gap: var(--space-12);
    padding: var(--space-12) var(--space-16);
    border-radius: var(--radius-8);
    border-left: 3px solid transparent;
    font-size: var(--font-size-13);
    line-height: var(--line-height-20);
  }
  .alert-icon {
    flex-shrink: 0;
  }
  .alert-body {
    flex: 1;
    min-width: 0;
  }

  .alert-info {
    background: var(--base-hlt-notr-easy);
    border-left-color: var(--base-line-act-minor);
    color: var(--color-interblue-920);
  }
  .alert-success {
    background: var(--color-lime-50);
    border-left-color: var(--color-lime-600);
    color: var(--color-green-750);
  }
  .alert-warning {
    background: var(--base-cont-note);
    border-left-color: var(--color-amber-500);
    color: var(--base-txt-notice-major);
  }
  .alert-error {
    background: var(--base-hlt-w-easy);
    border-left-color: var(--base-line-alert);
    color: var(--base-txt-alert-major);
  }
</style>
