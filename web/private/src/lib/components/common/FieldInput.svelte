<script lang="ts">
  import TextInput from "@/components/field/inputs/TextInput.svelte";
  import DateInput from "@/components/field/inputs/DateInput.svelte";
  import SelectInput from "@/components/field/inputs/SelectInput.svelte";
  import FileInput from "@/components/field/inputs/FileInput.svelte";
  import SignatureInput from "@/components/field/inputs/SignatureInput.svelte";
  import CellsInput from "@/components/field/inputs/CellsInput.svelte";

  interface Option {
    id?: string;
    value?: string;
    label?: string;
  }

  interface Props {
    type: string;
    value?: string | boolean | string[];
    placeholder?: string;
    required?: boolean;
    readonly?: boolean;
    disabled?: boolean;
    options?: Option[];
    error?: string;
    formula?: string;
    calculationType?: "number" | "currency";
    calculatedValue?: number;
    cellCount?: number;
    price?: number;
    currency?: string;
    /** Signature field format: '', drawn, typed, drawn_or_typed, drawn_or_upload, upload */
    signatureFormat?: string;
    /** Number field: min, max, step for validation and input attributes */
    numberMin?: number;
    numberMax?: number;
    numberStep?: string;
    /** Date field: display format (e.g. DD/MM/YYYY) */
    dateFormat?: string;
    /** Prefill for typed signature/initials tab (DocuSeal-style auto initials from submitter name). */
    defaultTypedText?: string;
    onBlur?: () => void;
  }

  let {
    type,
    // No fallback: parents may bind not-yet-initialized values (undefined);
    // a fallback here would make Svelte throw props_invalid_value.
    value = $bindable(),
    placeholder = "",
    required = false,
    readonly = false,
    disabled = false,
    options = [],
    error = "",
    formula,
    calculationType,
    calculatedValue,
    cellCount = 6,
    price = 0,
    currency = "USD",
    signatureFormat = "",
    numberMin = undefined,
    numberMax = undefined,
    numberStep = undefined,
    dateFormat = "",
    defaultTypedText = "",
    onBlur
  }: Props = $props();

  // Writable derived: resets to the incoming prop, still locally assignable (Vue: ref + watch)
  let localValue = $derived(value);

  const stringValue = $derived(typeof localValue === "string" ? localValue : "");

  const selectModelValue = $derived.by(() => {
    if (type === "checkbox") {
      const v = localValue;
      return v === true || v === "true";
    }
    if (type === "multiple" || type === "multi_select") {
      return Array.isArray(localValue) ? localValue : [];
    }
    if (type === "radio") {
      const v = localValue;
      return v != null && v !== "" ? String(v) : "";
    }
    return stringValue;
  });

  const isTextType = $derived(["text", "number", "phone"].includes(type));

  const isSelectType = $derived(["select", "radio", "checkbox", "multiple", "multi_select"].includes(type));

  const isCalculated = $derived(!!formula);

  const isSignatureType = $derived(["signature", "initials"].includes(type));

  function formatCalculated(val: number | undefined): string {
    if (val === undefined || val === null) {
      return "";
    }

    if (calculationType === "currency") {
      return new Intl.NumberFormat("en-US", {
        style: "currency",
        currency: "USD"
      }).format(val);
    }

    return val.toFixed(2);
  }

  function formatPaymentPrice(paymentPrice: number | undefined, paymentCurrency: string | undefined): string {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: paymentCurrency || "USD"
    }).format(paymentPrice ?? 0);
  }

  // Watch calculatedValue to update display for calculated fields
  // Note: For calculated fields, the value is displayed directly from calculatedValue prop
  // No need to update localValue as it's read-only

  function handleUpdate(newValue: string | boolean | string[]): void {
    localValue = newValue;
    value = newValue;
  }
</script>

<!-- Calculated field (read-only) -->
{#if isCalculated}
  <TextInput
    bind:value={() => formatCalculated(calculatedValue), () => {}}
    type="text"
    {placeholder}
    {required}
    readonly
    {disabled}
    {error}
    class="calculated-field"
    onBlur={() => onBlur?.()}
  />
{:else if isTextType}
  <TextInput
    bind:value={() => stringValue, (v) => handleUpdate(v)}
    {type}
    {placeholder}
    {required}
    {readonly}
    {disabled}
    {error}
    min={type === "number" ? numberMin : undefined}
    max={type === "number" ? numberMax : undefined}
    step={type === "number" ? numberStep : undefined}
    onBlur={() => onBlur?.()}
  />
{:else if type === "date"}
  <DateInput
    bind:value={() => stringValue, (v) => handleUpdate(v)}
    {dateFormat}
    {placeholder}
    {required}
    {readonly}
    {disabled}
    {error}
    onBlur={() => onBlur?.()}
  />
{:else if isSelectType}
  <SelectInput
    bind:value={() => selectModelValue, (v) => handleUpdate(v)}
    type={type as "select" | "radio" | "checkbox" | "multiple"}
    {placeholder}
    {required}
    {disabled}
    {options}
    {error}
    onBlur={() => onBlur?.()}
  />
{:else if type === "file" || type === "image"}
  <FileInput
    bind:value={() => stringValue, (v) => handleUpdate(v)}
    type={type as "file" | "image"}
    {disabled}
    {error}
    onBlur={() => onBlur?.()}
  />
{:else if isSignatureType || type === "stamp"}
  <SignatureInput
    bind:value={() => stringValue, (v) => handleUpdate(v)}
    mode={type === "initials" ? "initials" : type === "stamp" ? "stamp" : "signature"}
    format={signatureFormat}
    {defaultTypedText}
    {placeholder}
    disabled={disabled || (type === "stamp" && readonly)}
    {error}
    onBlur={() => onBlur?.()}
  />
{:else if type === "cells"}
  <CellsInput bind:value={() => stringValue, (v) => handleUpdate(v)} {cellCount} {readonly} {disabled} {error} />
{:else if type === "payment"}
  <div class="field-input-wrapper">
    <div class="text-lg font-semibold">
      {formatPaymentPrice(price, currency)}
    </div>
    <div class="mt-1 text-sm text-gray-500">
      {placeholder || "Payment required"}
    </div>
    {#if error}
      <div class="mt-1 text-sm text-[var(--color-error)]">{error}</div>
    {/if}
  </div>
{:else}
  <div class="field-input-wrapper">
    <div class="text-sm text-gray-500">Field type "{type}" not yet implemented</div>
  </div>
{/if}
