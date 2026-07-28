<script lang="ts">
  import Input from "@/components/ui/Input.svelte";
  import { t } from "@/i18n/index.svelte";
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

  function setToday(): void {
    const now = new Date();
    const month = String(now.getMonth() + 1).padStart(2, "0");
    const day = String(now.getDate()).padStart(2, "0");
    value = `${now.getFullYear()}-${month}-${day}`;
    onBlur?.();
  }

  function handleBlur(): void {
    onBlur?.();
  }
</script>

<div class="field-input-wrapper">
  <div class="flex items-center gap-2">
    <div class="min-w-0 flex-1">
      <Input bind:value type="date" {placeholder} {required} {readonly} {disabled} onblur={handleBlur} />
    </div>
    {#if !readonly && !disabled}
      <button type="button" class="btn btn-ghost btn-sm whitespace-nowrap" onclick={setToday}>
        {t("signing.setToday")}
      </button>
    {/if}
  </div>
  {#if dateFormat && formattedDisplay}
    <div class="mt-1 text-sm text-[--color-base-content]/70">
      {formattedDisplay}
    </div>
  {/if}
  {#if error}
    <div class="mt-1 text-sm text-[var(--color-error)]">{error}</div>
  {/if}
</div>
