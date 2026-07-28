<script lang="ts">
  import Input from "@/components/ui/Input.svelte";
  import { formatDateByPattern } from "@/utils/time";

  interface Props {
    value?: string;
    dateFormat?: string;
    placeholder?: string;
    required?: boolean;
    readonly?: boolean;
    disabled?: boolean;
    error?: string;
    onBlur?: () => void;
  }

  let {
    value = $bindable(""),
    dateFormat = "",
    placeholder = "",
    required = false,
    readonly = false,
    disabled = false,
    error = "",
    onBlur
  }: Props = $props();

  const formattedDisplay = $derived.by(() => {
    if (!dateFormat || !value) {
      return "";
    }
    return formatDateByPattern(value, dateFormat);
  });

  function handleBlur(): void {
    onBlur?.();
  }
</script>

<div class="field-input-wrapper">
  <Input bind:value type="date" {placeholder} {required} {readonly} {disabled} onblur={handleBlur} />
  {#if dateFormat && formattedDisplay}
    <div class="mt-1 text-sm text-[--color-base-content]/70">
      {formattedDisplay}
    </div>
  {/if}
  {#if error}
    <div class="mt-1 text-sm text-[var(--color-error)]">{error}</div>
  {/if}
</div>
