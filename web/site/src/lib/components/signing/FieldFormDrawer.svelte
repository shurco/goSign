<script lang="ts">
  import { tick } from "svelte";
  import { slide } from "svelte/transition";
  import { t } from "@/i18n/index.svelte";
  import FieldInput from "@/components/common/FieldInput.svelte";
  import FieldProgressDots from "@/components/common/FieldProgressDots.svelte";
  import Button from "@/components/ui/Button.svelte";
  import type { Field } from "@/models/template";

  interface Props {
    isOpen: boolean;
    field: Field | null;
    value?: string | boolean | string[];
    allFields: Field[];
    filledFieldIds: string[];
    fieldStates: Record<string, { visible?: boolean; required?: boolean; disabled?: boolean }>;
    fieldErrors: Record<string, string>;
    calculatedValues: Record<string, number | undefined>;
    signatureIds: Record<string, string>;
    getFieldLabel: (field: Field) => string;
    getCellCount: (field: Field) => number;
    getSignatureFormat: (field: Field) => string;
    hasWithSignatureId: (field: Field) => boolean;
    isFieldFilled: (field: Field) => boolean;
    /** Auto-initials text derived from submitter name (DocuSeal-style prefill for initials fields). */
    initialsDefault?: string;
    canGoPrev: boolean;
    canGoNext: boolean;
    isFormValid: boolean;
    isSubmitting: boolean;
    prevUnfilledIndex: number;
    nextUnfilledIndex: number;
    onClose?: () => void;
    onNavigate?: (direction: "prev" | "next") => void;
    onFieldSelect?: (fieldId: string) => void;
    onBlur?: (field: Field) => void;
    onReset?: () => void;
    onSubmit?: () => void;
  }

  let {
    isOpen,
    field,
    value = $bindable<string | boolean | string[]>(),
    allFields,
    filledFieldIds,
    fieldStates,
    fieldErrors,
    calculatedValues,
    signatureIds,
    getFieldLabel,
    getCellCount,
    getSignatureFormat,
    hasWithSignatureId,
    isFieldFilled,
    initialsDefault = "",
    canGoPrev: _canGoPrev,
    canGoNext: _canGoNext,
    isFormValid,
    isSubmitting,
    prevUnfilledIndex,
    nextUnfilledIndex,
    onClose,
    onNavigate,
    onFieldSelect,
    onBlur,
    onReset,
    onSubmit
  }: Props = $props();

  let drawerEl = $state<HTMLElement | null>(null);

  $effect(() => {
    if (isOpen) {
      tick().then(() => {
        drawerEl?.focus({ preventScroll: true });
      });
    }
  });

  function getCalculationType(fieldItem: Field): "number" | "currency" | undefined {
    return (fieldItem as { calculationType?: "number" | "currency" }).calculationType;
  }

  function handleClose(): void {
    onClose?.();
  }

  function handleNavigate(direction: "prev" | "next"): void {
    onNavigate?.(direction);
  }

  function handleFieldSelect(fieldId: string): void {
    onFieldSelect?.(fieldId);
  }

  function handleBlur(fieldItem: Field): void {
    onBlur?.(fieldItem);
  }

  function onKeydown(e: KeyboardEvent): void {
    if (e.key === "Escape") {
      handleClose();
    }
  }

  function handleReset(): void {
    onReset?.();
  }

  function handleSubmit(): void {
    onSubmit?.();
  }
</script>

<div
  bind:this={drawerEl}
  tabindex="-1"
  class="field-form-drawer pointer-events-none fixed inset-x-0 bottom-0 z-50 flex flex-col ring-0 outline-none focus:ring-0 focus:outline-none focus-visible:ring-0 focus-visible:outline-none"
  role="region"
  aria-label="Signing actions"
  onkeydown={onKeydown}
>
  <div class="field-form-drawer__panel relative flex flex-col">
    <div class="pb-safe container mx-auto w-full max-w-4xl px-4">
      <div
        class="field-form-drawer__card pointer-events-auto overflow-hidden rounded-t-lg border border-b-0 border-[var(--color-base-300)] bg-white shadow-[0_-2px_10px_rgba(0,0,0,0.06)] dark:border-neutral-600 dark:bg-neutral-800"
      >
        <!-- Expandable: dots + form (no handle, no extra nav row) -->
        {#if isOpen}
          <div class="flex max-h-[45vh] flex-col overflow-y-auto" transition:slide={{ duration: 300 }}>
            <!-- Current field form (keyed so each field mounts a fresh input: onMount prefill, no state leaks) -->
            {#if field}
              {#key field.id}
                <div class="border-t border-[var(--color-base-200)] px-3 py-3 first:border-t-0 dark:border-neutral-600">
                  <div class="mx-auto max-w-md">
                    <label class="mb-1.5 block text-xs font-medium text-[--color-base-content]/80">
                      {getFieldLabel(field)}
                      {#if fieldStates[field.id]?.required || field.required}
                        <span class="text-error">*</span>
                      {/if}
                    </label>
                    <FieldInput
                      bind:value
                      type={field.type}
                      required={fieldStates[field.id]?.required || field.required}
                      readonly={field.readonly}
                      disabled={fieldStates[field.id]?.disabled}
                      options={field.options}
                      placeholder={getFieldLabel(field)}
                      error={fieldErrors[field.id]}
                      formula={(field as { formula?: string }).formula ?? field.preferences?.formula}
                      calculationType={getCalculationType(field)}
                      calculatedValue={calculatedValues[field.id]}
                      cellCount={getCellCount(field)}
                      price={field.preferences?.price}
                      currency={field.preferences?.currency}
                      signatureFormat={getSignatureFormat(field)}
                      numberMin={field.type === "number"
                        ? (field as { validation?: { min?: number } }).validation?.min
                        : undefined}
                      numberMax={field.type === "number"
                        ? (field as { validation?: { max?: number } }).validation?.max
                        : undefined}
                      numberStep={field.type === "number"
                        ? (field as { validation?: { step?: string } }).validation?.step
                        : undefined}
                      dateFormat={field.type === "date"
                        ? ((field.preferences as { format?: string } | undefined)?.format ?? "")
                        : ""}
                      defaultTypedText={field.type === "initials" ? initialsDefault : ""}
                      onBlur={() => handleBlur(field)}
                    />
                    {#if hasWithSignatureId(field) && isFieldFilled(field) && signatureIds[field.id]}
                      <p class="mt-1.5 text-[11px] text-[--color-base-content]/60">
                        {field.type === "stamp" ? t("signing.stampId") : t("signing.signatureId")}:
                        <span class="font-mono">{signatureIds[field.id]}</span>
                      </p>
                    {/if}
                  </div>
                </div>
              {/key}
            {/if}
          </div>
        {/if}

        <!-- Action bar: compact, always visible -->
        <div
          class="flex flex-shrink-0 flex-wrap items-center justify-center gap-2 border-t border-[var(--color-base-200)] px-3 py-2 dark:border-neutral-600"
        >
          <Button
            type="button"
            variant="outline"
            size="sm"
            class="min-w-0"
            disabled={prevUnfilledIndex < 0}
            onclick={() => handleNavigate("prev")}
          >
            {t("common.previous")}
          </Button>
          <div class="flex min-w-0 flex-1 justify-center py-0.5">
            <FieldProgressDots
              fields={allFields}
              {filledFieldIds}
              currentFieldId={field?.id ?? null}
              {getFieldLabel}
              onFieldSelect={handleFieldSelect}
            />
          </div>
          {#if isFormValid}
            <Button
              type="button"
              variant="ghost"
              size="sm"
              class="min-w-0"
              disabled={isSubmitting}
              onclick={handleReset}
            >
              {t("common.reset")}
            </Button>
            <Button
              type="button"
              variant="primary"
              size="sm"
              class="min-w-0"
              loading={isSubmitting}
              disabled={isSubmitting}
              onclick={handleSubmit}
            >
              {t("signing.sign")}
            </Button>
          {/if}
          <Button
            type="button"
            variant="outline"
            size="sm"
            class="min-w-0"
            disabled={nextUnfilledIndex < 0}
            onclick={() => handleNavigate("next")}
          >
            {t("common.next")}
          </Button>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .field-form-drawer:focus,
  .field-form-drawer:focus-visible {
    outline: none;
    box-shadow: none;
  }
  .field-form-drawer__panel {
    background: transparent;
  }
  .field-form-drawer__card {
    min-height: 0;
  }
  .pb-safe {
    padding-bottom: env(safe-area-inset-bottom, 0);
  }
</style>
