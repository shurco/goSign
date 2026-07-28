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

  const modalClasses = $derived(
    {
      sm: "max-w-sm",
      md: "max-w-2xl",
      lg: "max-w-4xl",
      xl: "max-w-6xl"
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
    class="fixed inset-0 z-50 flex items-center justify-center p-4"
    transition:fade={{ duration: 200 }}
    use:portal
    use:escapeKey={handleEscape}
    onclick={handleBackdropClick}
    role="presentation"
  >
    <div class="fixed inset-0 bg-black/50 transition-opacity"></div>
    <div
      use:focusTrap={open}
      class="{modalClasses} relative z-10 flex max-h-[90vh] w-full flex-col rounded-lg border border-gray-200 bg-white transition-colors"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      {#if header || title}
        <div class="flex-shrink-0 border-b border-gray-200 px-6 py-4">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">
              {#if header}{@render header()}{:else}{title}{/if}
            </h3>
            {#if showClose}
              <button type="button" class="text-gray-400 transition-colors hover:text-gray-600" onclick={handleClose}>
                <SvgIcon name="x" class="h-5 w-5" />
              </button>
            {/if}
          </div>
        </div>
      {/if}
      <div class="flex-1 overflow-y-auto px-6 py-4">
        {@render children?.()}
      </div>
      {#if footer}
        <div class="flex-shrink-0 border-t border-gray-200 px-6 py-4">
          {@render footer()}
        </div>
      {/if}
    </div>
  </div>
{/if}
