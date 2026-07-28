<script lang="ts">
  import type { Snippet } from "svelte";

  export interface DropdownMenuItem {
    label: string;
    onclick: () => void;
    danger?: boolean;
    /** SVG icon name (SvgIcon) */
    icon?: string;
    divider?: boolean;
    disabled?: boolean;
  }

  import SvgIcon from "@/components/SvgIcon.svelte";

  let {
    items,
    children
  }: {
    items: DropdownMenuItem[];
    children: Snippet;
  } = $props();

  let open = $state(false);

  function toggle() {
    open = !open;
  }

  function handleClick(item: DropdownMenuItem) {
    if (item.disabled) {
      return;
    }
    item.onclick();
    open = false;
  }

  function handleOutsideClick() {
    if (open) {
      open = false;
    }
  }
</script>

<svelte:window onclick={handleOutsideClick} />

<div
  class="dd"
  role="menu"
  tabindex="-1"
  onclick={(e) => e.stopPropagation()}
  onkeydown={(e) => {
    if (e.key === "Escape") {
      open = false;
    }
  }}
>
  <button type="button" class="dd-trigger" onclick={toggle}>
    {@render children()}
  </button>
  {#if open}
    <div class="dd-menu">
      {#each items as item (item.label)}
        {#if item.divider}
          <div class="dd-divider"></div>
        {/if}
        <button
          type="button"
          class="dd-item"
          class:danger={item.danger}
          class:disabled={item.disabled}
          onclick={() => handleClick(item)}
        >
          {#if item.icon}
            <span class="dd-icon">
              <SvgIcon name={item.icon} width="14" height="14" />
            </span>
          {/if}
          {item.label}
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .dd {
    position: relative;
    display: inline-flex;
  }
  .dd-trigger {
    display: inline-flex;
    align-items: center;
    gap: var(--space-6);
    height: 34px;
    padding: 0 14px;
    border-radius: var(--radius-8);
    font-size: var(--font-size-13);
    font-weight: var(--font-weight-semibold);
    cursor: pointer;
    white-space: nowrap;
    font-family: inherit;
    background: var(--base-cont-top);
    color: var(--base-txt-primary);
    border: 1px solid transparent;
    box-shadow: 0 0 0 1px var(--base-btn-brd-minor) inset;
    transition: var(--transition-colors), var(--transition-shadow);
  }
  .dd-trigger:hover {
    background: var(--base-hlt-g-easy);
    box-shadow: 0 0 0 1px var(--base-line-accent) inset;
  }
  .dd-trigger:focus-visible {
    outline: none;
    box-shadow: var(--shadow-focus);
  }
  .dd-menu {
    position: absolute;
    top: 100%;
    right: 0;
    margin-top: var(--space-4);
    min-width: 200px;
    background: var(--base-cont-top);
    border-radius: var(--radius-12);
    box-shadow: var(--shadow-mod-minor);
    z-index: 100;
    padding: var(--space-4);
  }
  .dd-divider {
    height: 1px;
    background: var(--base-line-ghost);
    margin: var(--space-4) 0;
  }
  .dd-item {
    display: flex;
    align-items: center;
    gap: var(--space-8);
    width: 100%;
    padding: var(--space-8) var(--space-12);
    border-radius: var(--radius-6);
    font-size: var(--font-size-13);
    color: var(--base-txt-secondary);
    background: none;
    border: none;
    text-align: left;
    cursor: pointer;
    font-family: inherit;
    transition: var(--transition-colors);
  }
  .dd-item:hover:not(.disabled) {
    background: var(--base-hlt-g-hover);
  }
  .dd-item.danger {
    color: var(--base-txt-alert-minor);
  }
  .dd-item.danger:hover:not(.disabled) {
    background: var(--base-hlt-w-hover);
  }
  .dd-item.disabled {
    color: var(--base-txt-ghost);
    cursor: not-allowed;
    opacity: 0.5;
  }
  .dd-icon {
    display: flex;
    color: var(--base-txt-ghost);
  }
  .dd-item.danger .dd-icon {
    color: var(--base-txt-alert-minor);
  }
</style>
