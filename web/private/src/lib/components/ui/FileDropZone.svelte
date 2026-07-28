<script lang="ts">
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { t } from "@/i18n/index.svelte";

  interface Props {
    accept?: string;
    disabled?: boolean;
    /** Display when a file is selected (e.g. file name). */
    selectedLabel?: string;
    /** Height of the zone, e.g. 'h-32' or '128px'. Default h-32. */
    height?: string;
    /** Override "Click to upload" (otherwise from i18n templates.clickToUpload). */
    clickLabel?: string;
    /** Override "or drag and drop here" (otherwise from i18n templates.dragAndDrop). */
    dragLabel?: string;
    /** Override "Remove file" (otherwise from i18n templates.removeFile). */
    removeLabel?: string;
    onChange?: (file: File) => void;
    onClear?: () => void;
  }

  let {
    accept = undefined,
    disabled = false,
    selectedLabel = "",
    height = "h-32",
    clickLabel: clickLabelProp = "",
    dragLabel: dragLabelProp = "",
    removeLabel: removeLabelProp = "",
    onChange,
    onClear
  }: Props = $props();

  let inputRef = $state<HTMLInputElement | null>(null);

  const inputId = `file-drop-zone-${Math.random().toString(36).slice(2, 9)}`;

  const heightClass = $derived(height || "h-32");

  const clickLabel = $derived(clickLabelProp || t("templates.clickToUpload"));
  const dragLabel = $derived(dragLabelProp || t("templates.dragAndDrop"));
  const removeLabel = $derived(removeLabelProp || t("templates.removeFile"));

  function onInputChange(e: Event): void {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (file) {
      onChange?.(file);
    }
    input.value = "";
  }

  function onDrop(e: DragEvent): void {
    if (disabled) {
      return;
    }
    const file = e.dataTransfer?.files?.[0];
    if (file) {
      onChange?.(file);
    }
  }

  function clear(e: MouseEvent): void {
    e.stopPropagation();
    if (inputRef) {
      inputRef.value = "";
    }
    onClear?.();
  }
</script>

<label
  for={inputId}
  class="relative block w-full cursor-pointer rounded-xl border-2 border-dashed border-gray-300 transition-colors hover:bg-gray-50 {heightClass} {selectedLabel
    ? 'border-blue-400 bg-gray-50'
    : ''} {disabled ? 'cursor-not-allowed opacity-60' : ''}"
  ondragover={(e) => e.preventDefault()}
  ondrop={(e) => {
    e.preventDefault();
    onDrop(e);
  }}
>
  <div class="absolute inset-0 flex items-center justify-center p-2">
    <div class="flex flex-col items-center text-center">
      {#if !selectedLabel}
        <SvgIcon name="cloud-upload" class="h-8 w-8 shrink-0 text-gray-400" />
        <div class="mt-2 text-sm font-medium text-gray-700">{clickLabel}</div>
        <div class="text-xs text-gray-500">{dragLabel}</div>
      {:else}
        <SvgIcon name="document" class="h-8 w-8 shrink-0 text-blue-500" />
        <div class="mt-2 max-w-full truncate text-sm font-medium text-gray-900">{selectedLabel}</div>
        {#if !disabled}
          <button type="button" class="mt-1 text-xs text-red-600 hover:text-red-800" onclick={clear}>
            {removeLabel}
          </button>
        {/if}
      {/if}
    </div>
  </div>

  <input id={inputId} bind:this={inputRef} type="file" class="hidden" {accept} {disabled} onchange={onInputChange} />
</label>
