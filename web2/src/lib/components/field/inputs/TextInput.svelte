<script lang="ts">
  import Input from "@/components/ui/Input.svelte";

  interface Props {
    value?: string;
    type?: string;
    placeholder?: string;
    required?: boolean;
    readonly?: boolean;
    disabled?: boolean;
    error?: string;
    min?: number | string;
    max?: number | string;
    step?: string;
    class?: string;
    onBlur?: () => void;
  }

  let {
    value = $bindable(""),
    type = "text",
    placeholder = "",
    required = false,
    readonly = false,
    disabled = false,
    error = "",
    min,
    max,
    step,
    class: className = "",
    onBlur
  }: Props = $props();

  const inputType = $derived.by(() => {
    if (type === "number") {
      return "number" as const;
    }
    if (type === "phone") {
      return "tel" as const;
    }
    return "text" as const;
  });

  function handleBlur(): void {
    onBlur?.();
  }
</script>

<div class="field-input-wrapper {className}">
  <Input
    bind:value
    type={inputType}
    {placeholder}
    {required}
    {readonly}
    {disabled}
    {min}
    {max}
    {step}
    onblur={handleBlur}
  />
  {#if error}
    <div class="mt-1 text-sm text-[var(--color-error)]">{error}</div>
  {/if}
</div>
