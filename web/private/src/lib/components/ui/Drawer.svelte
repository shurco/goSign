<script lang="ts">
  /**
   * Right-side drawer: detail view over the current screen with a
   * dimmed/blurred backdrop (twing-m scrim). Closes on Esc / backdrop
   * click. Layout matches twing-m Modal.
   */
  import { fade } from "svelte/transition";
  import type { Snippet } from "svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

  let {
    open = false,
    title = "",
    width = "560px",
    onclose,
    actions,
    children
  }: {
    open?: boolean;
    title?: string;
    width?: string;
    onclose?: () => void;
    actions?: Snippet;
    children?: Snippet;
  } = $props();

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Escape" && open) {
      onclose?.();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
  <button class="drawer-backdrop" aria-label="Close" transition:fade={{ duration: 160 }} onclick={() => onclose?.()}
  ></button>
{/if}

<aside class="drawer" class:open style:width>
  <header class="drawer-header">
    <h2 class="drawer-title">{title}</h2>
    <button type="button" class="drawer-close" onclick={() => onclose?.()} aria-label="Close">
      <SvgIcon name="x" class="drawer-close-icon" />
    </button>
  </header>
  <div class="drawer-body">
    {#if children}
      {@render children()}
    {/if}
  </div>
  {#if actions}
    <footer class="drawer-footer">
      <div class="drawer-actions">{@render actions()}</div>
    </footer>
  {/if}
</aside>

<style>
  /* Dimmed and blurred backdrop (twing-m scrim) */
  .drawer-backdrop {
    position: fixed;
    inset: 0;
    z-index: 180;
    background: var(--color-scrim);
    backdrop-filter: blur(3px);
    -webkit-backdrop-filter: blur(3px);
    border: none;
    cursor: default;
    padding: 0;
  }
  .drawer {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    max-width: calc(100vw - var(--layout-rail-width) - var(--space-24));
    background: var(--base-cont-top);
    border-radius: var(--radius-16) 0 0 var(--radius-16);
    box-shadow: var(--shadow-mod-major);
    transform: translateX(105%);
    transition: transform 0.22s ease;
    z-index: 190;
    display: flex;
    flex-direction: column;
  }
  .drawer.open {
    transform: translateX(0);
  }
  .drawer-header {
    display: flex;
    align-items: center;
    gap: var(--space-8);
    flex-shrink: 0;
    padding: var(--space-16) var(--space-20);
    border-bottom: 1px solid var(--base-line-tertiary);
  }
  .drawer-title {
    flex: 1;
    min-width: 0;
    margin: 0;
    font-size: var(--font-size-16);
    font-weight: var(--font-weight-bold);
    color: var(--base-txt-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .drawer-close {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    flex-shrink: 0;
    border: none;
    border-radius: var(--radius-8);
    background: transparent;
    color: var(--base-txt-muted);
    cursor: pointer;
    transition: var(--transition-colors);
  }
  .drawer-close:hover {
    background: var(--base-hlt-g-hover);
    color: var(--base-txt-primary);
  }
  .drawer-close :global(.drawer-close-icon) {
    width: 18px;
    height: 18px;
  }
  .drawer-body {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: var(--space-20);
  }
  /* Action buttons at the bottom, like twing-m form-actions */
  .drawer-footer {
    flex-shrink: 0;
    padding: var(--space-12) var(--space-20);
    border-top: 1px solid var(--base-line-tertiary);
    background: var(--base-cont-top);
  }
  .drawer-actions {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: var(--space-8);
  }
</style>
