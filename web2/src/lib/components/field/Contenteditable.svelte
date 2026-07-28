<script lang="ts">
  import SvgIcon from "@/components/SvgIcon.svelte";

  interface Props {
    value?: string;
    iconInline?: boolean;
    iconWidth?: number;
    withRequired?: boolean;
    selectOnEditClick?: boolean;
    editable?: boolean;
    iconStrokeWidth?: number;
    onFocus?: (event: FocusEvent) => void;
    onBlur?: (event: FocusEvent) => void;
    onValueChange?: (value: string) => void;
    class?: string;
    style?: string;
  }

  let {
    value = $bindable(""),
    iconInline = false,
    iconWidth = 30,
    withRequired = false,
    selectOnEditClick = false,
    editable = true,
    iconStrokeWidth = 2,
    onFocus,
    onBlur,
    onValueChange,
    class: className = "",
    style
  }: Props = $props();

  let contenteditableEl: HTMLSpanElement | null = $state(null);
  // Writable derived: resets to the incoming prop, still locally assignable (Vue: ref + watch)
  let internalValue = $derived(value);

  export function getContenteditable(): HTMLSpanElement | null {
    return contenteditableEl;
  }

  function selectContent(): void {
    setTimeout(() => {
      const el = contenteditableEl;
      if (!el) {
        return;
      }

      const range = document.createRange();
      range.selectNodeContents(el);
      const sel = window.getSelection();
      sel?.removeAllRanges();
      sel?.addRange(range);
    }, 10);
  }

  function handleBlur(e: FocusEvent): void {
    setTimeout(() => {
      if (contenteditableEl) {
        internalValue = contenteditableEl.innerText.trim() || value;
        value = internalValue;
        onValueChange?.(internalValue);
      }
      onBlur?.(e);
    }, 1);
  }

  function focusContenteditable(): void {
    contenteditableEl?.focus();
  }

  function blurContenteditable(): void {
    contenteditableEl?.blur();
  }

  function handlePencilClick(): void {
    focusContenteditable();
    if (selectOnEditClick) {
      selectContent();
    }
  }
</script>

<div
  class="group/contenteditable relative overflow-visible {className} {iconInline ? 'flex items-center' : ''}"
  {style}
>
  <span
    bind:this={contenteditableEl}
    dir="auto"
    contenteditable={editable}
    role="textbox"
    style="min-width: 2px"
    class="peer outline-none focus:block {iconInline ? 'inline' : 'block'}"
    onkeydown={(e) => {
      if (e.key === "Enter") {
        e.preventDefault();
        blurContenteditable();
      }
    }}
    onfocus={(e) => onFocus?.(e)}
    onblur={handleBlur}
  >
    {internalValue}
  </span>
  {#if withRequired}
    <span title="Required" class="text-red-500 peer-focus:hidden" onclick={focusContenteditable} role="presentation"
      >*</span
    >
  {/if}
  <SvgIcon
    name="pencil"
    class="flex-none cursor-pointer align-middle opacity-0 group-hover/contenteditable:opacity-100 group-hover/contenteditable-container:opacity-100 peer-focus:hidden {editable
      ? ''
      : 'invisible'} {withRequired ? '' : 'ml-1'} {iconInline ? 'inline align-bottom' : 'absolute'}"
    style={iconInline ? undefined : `right: ${-(1.1 * iconWidth)}px`}
    title="Edit"
    width={iconWidth}
    height={iconWidth}
    stroke-width={iconStrokeWidth}
    onclick={handlePencilClick}
  />
</div>
