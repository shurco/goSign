<script lang="ts">
  import type { Snippet } from "svelte";
  import Modal from "@/components/ui/Modal.svelte";
  import Button from "@/components/ui/Button.svelte";

  interface Props {
    open: boolean;
    title: string;
    size?: "sm" | "md" | "lg" | "xl";
    submitText?: string;
    cancelText?: string;
    showCancel?: boolean;
    closeButton?: boolean;
    closeOnOutsideClick?: boolean;
    closeOnEscape?: boolean;
    validateOnMount?: boolean;
    onSubmit?: (formData: Record<string, unknown>) => void | Promise<void>;
    onSubmitEvent?: (formData: Record<string, unknown>) => void;
    onCancel?: () => void;
    onClose?: () => void;
    children?: Snippet<[formData: Record<string, any>, errors: Record<string, string>]>;
    footer?: Snippet<[submit: () => void, cancel: () => void, isSubmitting: boolean]>;
  }

  let {
    open: isOpen = $bindable(),
    title,
    size = "md",
    submitText = "Submit",
    cancelText = "Cancel",
    showCancel = true,
    closeButton = true,
    closeOnOutsideClick = true,
    closeOnEscape = true,
    validateOnMount = false,
    onSubmit,
    onSubmitEvent,
    onCancel,
    onClose,
    children: formChildren,
    footer: footerSnippet
  }: Props = $props();

  let formData = $state<Record<string, any>>({});
  let errors = $state<Record<string, string>>({});
  let isSubmitting = $state(false);

  const formIsValid = $derived(Object.keys(errors).length === 0);

  $effect(() => {
    if (isOpen) {
      if (validateOnMount) {
        validateForm();
      }
      // Initialize formData when modal opens
      // Initialize formData with default values if empty
      // Individual fields will be initialized by v-model in the template
      if (!formData.name) {
        formData.name = "";
      }
      if (!formData.email) {
        formData.email = "";
      }
      if (!formData.role) {
        formData.role = "member";
      }
    } else {
      // Reset form when modal closes
      resetForm();
    }
  });

  function handleClose(): void {
    if (!isSubmitting) {
      isOpen = false;
      onClose?.();
      onCancel?.();
      resetForm();
    }
  }

  async function handleSubmit(): Promise<void> {
    if (isSubmitting || !formIsValid) {
      return;
    }

    validateForm();

    if (!formIsValid) {
      return;
    }

    isSubmitting = true;

    try {
      if (onSubmit) {
        const result = onSubmit(formData);
        // If we got a Promise, wait for it, then reset the submitting state.
        // For sync handlers, keep isSubmitting true: the parent closes the
        // modal or resets it on success/error, preserving the loading state.
        if (result instanceof Promise) {
          await result;
          isSubmitting = false;
        }
      } else {
        // Emit submit event - parent handler should handle async operations
        onSubmitEvent?.(formData);
      }
    } catch (error) {
      // Re-throw error so parent component can handle it
      console.error("FormModal submit error:", error);
      isSubmitting = false;
      throw error;
    }
  }

  // Expose method to reset submitting state (for parent components)
  export function resetSubmitting(): void {
    isSubmitting = false;
  }

  export function validateForm(): void {
    errors = {};
  }

  export function resetForm(): void {
    formData = {};
    errors = {};
  }

  export function setFormData(data: Record<string, unknown>): void {
    formData = { ...data };
  }

  export function getFormData(): Record<string, any> {
    return formData;
  }

  export function setError(field: string, message: string): void {
    errors[field] = message;
  }

  export function clearError(field: string): void {
    const { [field]: _removed, ...rest } = errors;
    errors = rest;
  }

  export function clearAllErrors(): void {
    errors = {};
  }

  export function open(): void {
    isOpen = true;
  }

  export function close(): void {
    handleClose();
  }

  export function isValid(): boolean {
    return formIsValid;
  }
</script>

<Modal
  bind:open={isOpen}
  {title}
  {size}
  showClose={closeButton}
  closeOnBackdrop={closeOnOutsideClick}
  {closeOnEscape}
  onClose={handleClose}
>
  <div class="py-4">
    {@render formChildren?.(formData, errors)}
  </div>

  {#snippet footer()}
    {#if footerSnippet}
      {@render footerSnippet(handleSubmit, handleClose, isSubmitting)}
    {:else}
      <div class="flex justify-end gap-3">
        {#if showCancel}
          <Button variant="ghost" disabled={isSubmitting} onclick={handleClose}>
            {cancelText}
          </Button>
        {/if}
        <Button variant="primary" loading={isSubmitting} disabled={isSubmitting || !formIsValid} onclick={handleSubmit}>
          {submitText}
        </Button>
      </div>
    {/if}
  {/snippet}
</Modal>
