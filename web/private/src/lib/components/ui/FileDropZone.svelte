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
  class="drop-zone relative block w-full {heightClass}"
  class:selected={Boolean(selectedLabel)}
  class:disabled
  ondragover={(e) => e.preventDefault()}
  ondrop={(e) => {
    e.preventDefault();
    onDrop(e);
  }}
>
  <div class="absolute inset-0 flex items-center justify-center p-2">
    <div class="flex flex-col items-center text-center">
      {#if !selectedLabel}
        <SvgIcon name="cloud-upload" class="drop-icon h-8 w-8 shrink-0" />
        <div class="drop-click mt-2">{clickLabel}</div>
        <div class="drop-hint">{dragLabel}</div>
      {:else}
        <SvgIcon name="document" class="drop-icon-selected h-8 w-8 shrink-0" />
        <div class="drop-name mt-2 max-w-full truncate">{selectedLabel}</div>
        {#if !disabled}
          <button type="button" class="drop-remove mt-1" onclick={clear}>
            {removeLabel}
          </button>
        {/if}
      {/if}
    </div>
  </div>

  <input id={inputId} bind:this={inputRef} type="file" class="hidden" {accept} {disabled} onchange={onInputChange} />
</label>

<style>
  .drop-zone {
    cursor: pointer;
    border: 2px dashed var(--base-line-secondary);
    border-radius: var(--radius-12);
    background: var(--base-cont-top);
    transition: var(--transition-colors);
  }
  .drop-zone:hover {
    background: var(--base-hlt-easy);
    border-color: var(--base-line-act-minor);
  }
  .drop-zone.selected {
    border-color: var(--base-line-act-minor);
    background: var(--base-hlt-easy);
  }
  .drop-zone.disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }
  .drop-zone :global(.drop-icon) {
    color: var(--base-txt-ghost);
  }
  .drop-zone :global(.drop-icon-selected) {
    color: var(--base-hlt-invert);
  }
  .drop-click {
    font-size: var(--font-size-13);
    font-weight: var(--font-weight-medium);
    color: var(--base-txt-secondary);
  }
  .drop-hint {
    font-size: var(--font-size-12);
    color: var(--base-txt-muted);
  }
  .drop-name {
    font-size: var(--font-size-13);
    font-weight: var(--font-weight-medium);
    color: var(--base-txt-primary);
  }
  .drop-remove {
    border: none;
    background: none;
    padding: 0;
    font-family: inherit;
    font-size: var(--font-size-12);
    color: var(--base-txt-alert-minor);
    cursor: pointer;
    transition: var(--transition-colors);
  }
  .drop-remove:hover {
    color: var(--base-txt-alert-major);
  }
</style>
