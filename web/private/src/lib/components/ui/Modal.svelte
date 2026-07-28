<script lang="ts">
  import type { Snippet } from "svelte";
  import { fade } from "svelte/transition";
  import { escapeKey, focusTrap, portal } from "@/composables/ui.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

  interface Props {
    open: boolean;
    title?: string;
    size?: "sm" | "md" | "lg" | "xl";
    showClose?: boolean;
    closeOnBackdrop?: boolean;
    closeOnEscape?: boolean;
    onClose?: () => void;
    header?: Snippet;
    footer?: Snippet;
    children?: Snippet;
  }

  let {
    open = $bindable(),
    title = "",
    size = "md",
    showClose = true,
    closeOnBackdrop = true,
    closeOnEscape = true,
    onClose,
    header,
    footer,
    children
  }: Props = $props();

  const modalWidth = $derived(
    {
      sm: "384px",
      md: "672px",
      lg: "896px",
      xl: "1152px"
    }[size]
  );

  $effect(() => {
    if (open) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
    return () => {
      document.body.style.overflow = "";
    };
  });

  function handleClose(): void {
    open = false;
    onClose?.();
  }

  function handleBackdropClick(): void {
    if (closeOnBackdrop) {
      handleClose();
    }
  }

  function handleEscape(): void {
    if (closeOnEscape) {
      handleClose();
    }
  }
</script>

{#if open}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div
    class="modal-overlay"
    transition:fade={{ duration: 160 }}
    use:portal
    use:escapeKey={handleEscape}
    onclick={handleBackdropClick}
    role="presentation"
  >
    <div
      use:focusTrap={open}
      class="modal-window"
      style={`width: min(${modalWidth}, calc(100vw - 32px));`}
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      {#if header || title}
        <div class="modal-header">
          <h3 class="modal-title">
            {#if header}{@render header()}{:else}{title}{/if}
          </h3>
          {#if showClose}
            <button type="button" class="modal-close" onclick={handleClose} aria-label="Close">
              <SvgIcon name="x" class="modal-close-icon" />
            </button>
          {/if}
        </div>
      {/if}
      <div class="modal-body">
        {@render children?.()}
      </div>
      {#if footer}
        <div class="modal-footer">
          {@render footer()}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  /* twing-m Modal: scrim + white window r16 */
  .modal-overlay {
    position: fixed;
    inset: 0;
    z-index: 300;
    background: var(--color-scrim);
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding: 80px var(--space-16) var(--space-16);
    overflow-y: auto;
  }
  .modal-window {
    display: flex;
    flex-direction: column;
    max-height: calc(100vh - 112px);
    background: var(--base-cont-top);
    border-radius: var(--radius-16);
    box-shadow: var(--shadow-mod-major);
  }
  .modal-header {
    flex-shrink: 0;
    padding: var(--space-16) var(--space-20);
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-8);
    border-bottom: 1px solid var(--base-line-tertiary);
  }
  .modal-title {
    font-size: var(--font-size-16);
    font-weight: var(--font-weight-bold);
    color: var(--base-txt-primary);
    margin: 0;
    min-width: 0;
  }
  .modal-close {
    display: flex;
    flex-shrink: 0;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--base-txt-ghost);
    padding: var(--space-4);
    border-radius: var(--radius-6);
    transition: var(--transition-colors);
  }
  .modal-close:hover {
    background: var(--base-hlt-g-hover);
    color: var(--base-txt-secondary);
  }
  .modal-close :global(.modal-close-icon) {
    width: 18px;
    height: 18px;
  }
  .modal-body {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    padding: var(--space-20);
  }
  .modal-footer {
    flex-shrink: 0;
    padding: var(--space-12) var(--space-20);
    border-top: 1px solid var(--base-line-tertiary);
  }
</style>
