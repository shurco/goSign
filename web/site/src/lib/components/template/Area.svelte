<script lang="ts">
  import { getContext, tick, untrack } from "svelte";
  import FieldSubmitter from "@/components/field/Submitter.svelte";
  import FieldType from "@/components/field/Type.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { bgColors, borderColors, fieldIcons, fieldNames, subNames } from "@/components/field/constants";
  import { t } from "@/i18n/index.svelte";
  import type { Template } from "@/models/index";
  import type { Area, Field } from "@/models/template";
  import { formatDateByPattern } from "@/utils/time";

  interface Props {
    area: Area;
    isDraw?: boolean;
    defaultField?: Field | null;
    editable?: boolean;
    field: Field | null;
    onStartResize?: (direction: "nwse" | "ew") => void;
    onStopResize?: () => void;
    onStartDrag?: () => void;
    onStopDrag?: () => void;
    onRemove?: () => void;
    onSelectSubmitter?: (submitterId: string) => void;
  }

  let {
    area,
    isDraw = false,
    defaultField = null,
    editable = true,
    field,
    onStartResize,
    onStopResize,
    onStartDrag,
    onStopDrag,
    onRemove,
    onSelectSubmitter
  }: Props = $props();

  const template = getContext<{ value: Template }>("template");
  const selectedAreaRef = getContext<{ value: Area | null }>("selectedAreaRef");
  const save = getContext<() => void>("save");

  if (!template || !selectedAreaRef || !save) {
    throw new Error("Required injectables are missing in Area component");
  }

  let rootRef = $state<HTMLElement | null>(null);
  let name = $state<HTMLElement | null>(null);
  let textContainer = $state<HTMLElement | null>(null);
  let defaultValue = $state<HTMLElement | null>(null);
  let touchTarget = $state<HTMLElement | null>(null);

  let isDragged = $state(false);

  const isValueInput = $derived(field ? ["text", "number", "date", "cells"].includes(field.type) : false);
  const isCheckboxInput = $derived(field?.type === "checkbox");
  const isSelectInput = $derived(field ? ["select", "radio", "multiple"].includes(field.type) : false);
  const isDefaultValuePresent = $derived.by(() => {
    if (!field) {
      return false;
    }
    const v = field.default_value;
    if (field.type === "checkbox") {
      return !!v;
    }
    if (field.type === "multiple") {
      return Array.isArray(v) ? v.length > 0 : !!v;
    }
    if (field.type === "radio" || field.type === "select") {
      return v != null && v !== "";
    }
    if (field.type === "number") {
      return v != null && v !== "" && !Number.isNaN(Number(v));
    }
    return v != null && v !== "";
  });
  const isWFullType = $derived(field ? ["text", "number", "date", "cells", "select"].includes(field.type) : false);

  const areaWiderThanHigh = $derived(area.w > area.h);

  const hasMultipleAreas = $derived((field?.areas?.length ?? 0) > 1);

  const isDatePlaceholder = $derived(field?.default_value === "\u007b\u007bdate\u007d\u007d");

  function buildAreaOptionValue(areaItem: Area): string {
    if (!areaItem.option_id || !field?.options) {
      return "";
    }
    const option = field.options.find((o: { id?: string; value?: string }) => o.id === areaItem.option_id);
    return (option && "value" in option ? option.value : "") ?? "";
  }

  function formatNumber(value: number | string, format?: string): string {
    const num = typeof value === "string" ? parseFloat(value) : value;
    if (Number.isNaN(num)) {
      return String(value);
    }
    if (!format || format === "none") {
      return String(num);
    }
    if (format === "comma") {
      return new Intl.NumberFormat("en-US").format(num);
    }
    if (format === "dot") {
      return new Intl.NumberFormat("de-DE").format(num);
    }
    if (format === "space") {
      return new Intl.NumberFormat("fr-FR").format(num);
    }
    if (format === "usd") {
      return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(num);
    }
    if (format === "eur") {
      return new Intl.NumberFormat("de-DE", { style: "currency", currency: "EUR" }).format(num);
    }
    if (format === "gbp") {
      return new Intl.NumberFormat("en-GB", { style: "currency", currency: "GBP" }).format(num);
    }
    if (format === "percent") {
      return `${num}%`;
    }
    return String(num);
  }

  // Vue bound this as a :style object; Svelte style attributes take a CSS string
  const fontStyle = $derived.by(() => {
    if (!field?.preferences) {
      return "";
    }
    const style: string[] = [];
    if (field.preferences.font) {
      style.push(`font-family: ${field.preferences.font}`);
    }
    if (field.preferences.font_size) {
      style.push(`font-size: ${String(field.preferences.font_size)}px`);
    }
    if (field.preferences.font_type) {
      if (field.preferences.font_type === "bold") {
        style.push("font-weight: bold");
      } else if (field.preferences.font_type === "italic") {
        style.push("font-style: italic");
      } else if (field.preferences.font_type === "bold_italic") {
        style.push("font-weight: bold");
        style.push("font-style: italic");
      }
    }
    if (field.preferences.align) {
      style.push(`text-align: ${field.preferences.align}`);
      style.push(
        `justify-content: ${field.preferences.align === "center" ? "center" : field.preferences.align === "right" ? "flex-end" : "flex-start"}`
      );
    }
    if (field.preferences.color) {
      style.push(`color: ${String(field.preferences.color)}`);
    }
    if (field.preferences.valign) {
      style.push(
        `align-items: ${field.preferences.valign === "top" ? "flex-start" : field.preferences.valign === "bottom" ? "flex-end" : "center"}`
      );
    }
    return style.join("; ");
  });

  let isNameFocus = $state(false);
  let textOverflowChars = $state(0);
  let dragFrom = $state({ x: 0, y: 0 });
  let pointerMode = $state<"drag" | "resize" | "resize-cell" | null>(null);
  let lastClickTime = $state(0);

  const defaultName = $derived.by(() => {
    if (!field) {
      return "Field";
    }

    if (field.type === "payment" && field.preferences?.price) {
      const { price, currency } = field.preferences;
      const formattedPrice = new Intl.NumberFormat([], {
        style: "currency",
        currency
      }).format(price ?? 0);
      return `${fieldNames[field.type]} ${formattedPrice}`;
    }

    const partyName = subNames[submitterIndex]?.replace(" Party", "") || "First";
    const typeName = fieldNames[field.type] || "Field";
    if (!field) {
      return `${partyName} ${typeName} 1`;
    }
    const sameTypeAndPartyFields = template.value.fields.filter(
      (f) => f.type === field?.type && f.submitter_id === field?.submitter_id && f.id !== field?.id
    );
    const fieldNumber = sameTypeAndPartyFields.length + 1;

    return `${partyName} ${typeName} ${fieldNumber}`;
  });

  const optionIndexText = $derived.by(() => {
    const opts = field?.options as Array<{ id?: string }> | undefined;
    if (!area.option_id || !opts?.length) {
      return "";
    }
    const idx = opts.findIndex((o) => o && typeof o === "object" && "id" in o && o.id === area.option_id);
    return idx >= 0 ? `${idx + 1}.` : "";
  });

  function getEffectiveCellW(areaItem: Area): number {
    if (areaItem.cell_w != null && areaItem.cell_w > 0) {
      return areaItem.cell_w;
    }
    if (areaItem.w <= 0) {
      return 0;
    }
    if (areaItem.h > 0) {
      const denom = Math.floor(areaItem.w / areaItem.h);
      return denom > 0 ? (areaItem.w * 2) / denom : areaItem.w / 5;
    }
    return areaItem.w / 5;
  }

  /** Returns number of cells for a cells-type area (same formula as cells computed). */
  function getCellCountFromArea(areaItem: Area): number {
    const cellWidth = getEffectiveCellW(areaItem);
    if (!cellWidth || cellWidth <= 0 || areaItem.w <= 0) {
      return 0;
    }
    let currentWidth = 0;
    let count = 0;
    while (currentWidth + (cellWidth + cellWidth / 4) < areaItem.w) {
      currentWidth += cellWidth;
      count++;
    }
    return Math.max(count, 1);
  }

  const effectiveCellW = $derived(getEffectiveCellW(area));

  const cells = $derived.by(() => {
    const cellsList: number[] = [];
    const cellWidth = getEffectiveCellW(area);
    if (!cellWidth || cellWidth <= 0) {
      return cellsList;
    }

    let currentWidth = 0;
    while (currentWidth + (cellWidth + cellWidth / 4) < area.w) {
      currentWidth += cellWidth;
      cellsList.push(currentWidth);
    }
    return cellsList;
  });

  const submitter = $derived(template.value.submitters.find((s) => s.id === field?.submitter_id));

  const submitterIndex = $derived.by(() => {
    if (!submitter) {
      return 0;
    }
    return template.value.submitters.indexOf(submitter);
  });

  const isSelected = $derived(selectedAreaRef?.value === area);

  const positionStyle = $derived.by(() => {
    const { x, y, w, h } = area;
    return {
      top: y * 100 + "%",
      left: x * 100 + "%",
      width: w * 100 + "%",
      height: h * 100 + "%"
    };
  });

  let defaultValueWatchReady = false;
  $effect(() => {
    // Read the watched source unconditionally so the effect tracks it despite short-circuiting below
    const defaultVal = field?.default_value;
    if (
      defaultValueWatchReady &&
      field?.type === "text" &&
      defaultVal &&
      textContainer &&
      (textOverflowChars === 0 || textOverflowChars - 4 > defaultVal.length)
    ) {
      tick().then(() => {
        const el = document.querySelector(".group.absolute.overflow-visible") as HTMLElement;
        if (el && textContainer && field?.default_value) {
          textOverflowChars = el.clientHeight < textContainer.clientHeight ? field.default_value.length : 0;
        } else {
          textOverflowChars = 0;
        }
      });
    }
    untrack(() => {
      defaultValueWatchReady = true;
    });
  });

  let submitterIdWatchReady = false;
  let prevSubmitterId: string | undefined = undefined;
  $effect(() => {
    const newSubmitterId = field?.submitter_id;
    if (submitterIdWatchReady && newSubmitterId !== prevSubmitterId && isDefaultName(field?.name || "")) {
      if (field) {
        field.name = "";
        save();
      }
    }
    prevSubmitterId = newSubmitterId;
    untrack(() => {
      submitterIdWatchReady = true;
    });
  });

  function isDefaultName(nameText: string): boolean {
    if (!nameText) {
      return true;
    }
    const pattern = /^(First|Second|Third|Fourth|Fifth|Sixth|Seventh|Eighth|Ninth|Tenth)\s+\w+\s+\d+$/;
    return pattern.test(nameText);
  }

  function onNameFocus(): void {
    selectedAreaRef.value = area;
    isNameFocus = true;
    if (name) {
      name.style.minWidth = name.clientWidth + "px";
    }

    if (!field?.name) {
      setTimeout(() => {
        if (name) {
          // eslint-disable-next-line svelte/no-dom-manipulating -- contenteditable text is owned by the user/DOM, not Svelte (ported from Vue)
          name.innerText = " ";
        }
      }, 1);
    }
  }

  function onNameBlur(): void {
    const text = name?.innerText.trim() || "";
    isNameFocus = false;
    if (name) {
      name.style.minWidth = "";
    }

    if (field) {
      field.name = text || "";
      if (!text && name) {
        // eslint-disable-next-line svelte/no-dom-manipulating -- contenteditable text is owned by the user/DOM, not Svelte (ported from Vue)
        name.innerText = defaultName;
      }
      save?.();
    }
  }

  function onNameEnter(): void {
    name?.blur();
  }

  function onDefaultValueBlur(): void {
    const el = defaultValue;
    if (!el || !field) {
      return;
    }
    const text = el.innerText.trim();
    if (field.default_value !== text) {
      field.default_value = text;
      save?.();
    }
  }

  function onDefaultValuePaste(e: ClipboardEvent): void {
    e.preventDefault();
    const text = e.clipboardData?.getData("text") ?? "";
    document.execCommand("insertText", false, text);
  }

  function onDefaultValueEnter(): void {
    defaultValue?.blur();
  }

  function focusArea(a: Area): void {
    selectedAreaRef.value = a;
  }

  function maybeUpdateOptions(): void {
    if (!field) {
      return;
    }

    if (field.type !== "cells") {
      delete field.default_value;
    }

    if (!["radio", "multiple", "select"].includes(field.type)) {
      delete field.options;
    }

    if (["select", "multiple", "radio"].includes(field.type)) {
      type OptionItem = { id: string; value: string };
      const opts = (field as { options?: OptionItem[] }).options;
      (field as { options: OptionItem[] }).options = opts?.length ? opts : [{ value: "", id: crypto.randomUUID() }];
    }

    (field.areas || []).forEach((areaItem: Area) => {
      if (field?.type === "cells") {
        const denom = areaItem.h > 0 ? Math.floor(areaItem.w / areaItem.h) : 0;
        if (denom > 0) {
          areaItem.cell_w = (areaItem.w * 2) / denom;
        } else if (areaItem.w > 0 && (!areaItem.cell_w || areaItem.cell_w <= 0)) {
          areaItem.cell_w = areaItem.w / 5;
        }
        areaItem.cell_count = getCellCountFromArea(areaItem);
      } else {
        delete areaItem.cell_w;
        delete areaItem.cell_count;
      }
    });
  }

  function handleSubmitterUpdate(submitterId: string): void {
    save();
    onSelectSubmitter?.(submitterId);
  }

  function handleTypeUpdate(): void {
    maybeUpdateOptions();
    save();
  }

  // Handle double click to select submitter
  function handleDoubleClick(): void {
    if (field) {
      onSelectSubmitter?.(field.submitter_id);
    }
  }

  // Unified pointer event handlers
  function handleDragStart(e: PointerEvent): void {
    selectedAreaRef.value = area;

    if (!editable) {
      return;
    }

    // Check for double click to prevent drag
    const now = Date.now();
    if (now - lastClickTime < 300) {
      lastClickTime = 0;
      return;
    }
    lastClickTime = now;

    const target = e.target as HTMLElement;
    if (target !== touchTarget && e.pointerType === "touch") {
      return;
    }

    if (e.pointerType === "touch") {
      name?.blur();
      e.preventDefault();
    }

    const el = rootRef || (e.target as HTMLElement);
    const rect = el.getBoundingClientRect();
    dragFrom = { x: e.clientX - rect.left, y: e.clientY - rect.top };
    pointerMode = "drag";

    document.addEventListener("pointermove", handlePointerMove);
    document.addEventListener("pointerup", handlePointerUp);
    onStartDrag?.();
  }

  function handleResizeStart(e: PointerEvent): void {
    selectedAreaRef.value = area;
    if (e.pointerType === "touch") {
      name?.blur();
      e.preventDefault();
    }

    pointerMode = "resize";
    document.addEventListener("pointermove", handlePointerMove);
    document.addEventListener("pointerup", handlePointerUp);
    onStartResize?.("nwse");
  }

  function handleResizeCellStart(): void {
    pointerMode = "resize-cell";
    document.addEventListener("pointermove", handlePointerMove);
    document.addEventListener("pointerup", handlePointerUp);
    onStartResize?.("ew");
  }

  function handlePointerMove(e: PointerEvent): void {
    if (pointerMode === "drag") {
      handleDrag(e);
    } else if (pointerMode === "resize") {
      handleResize(e);
    } else if (pointerMode === "resize-cell") {
      handleResizeCell(e);
    }
  }

  function handlePointerUp(): void {
    document.removeEventListener("pointermove", handlePointerMove);
    document.removeEventListener("pointerup", handlePointerUp);

    if (pointerMode === "drag") {
      if (isDragged) {
        save?.();
      }
      isDragged = false;
      onStopDrag?.();
    } else if (pointerMode === "resize" || pointerMode === "resize-cell") {
      onStopResize?.();
      save?.();
    }

    pointerMode = null;
  }

  function getMaskForArea(): HTMLElement | null {
    let current: HTMLElement | null = touchTarget;
    if (!current) {
      return null;
    }

    while (current && current.parentElement) {
      current = current.parentElement;
      if (current.classList.contains("relative") && current.classList.contains("cursor-crosshair")) {
        const maskEl = current.querySelector("#mask") as HTMLElement | null;
        if (maskEl && typeof maskEl.clientWidth === "number" && maskEl.clientWidth > 0) {
          return maskEl;
        }
        if (!maskEl && import.meta.env.DEV) {
          console.warn("[Area] getMaskForArea: #mask not found in page container");
        }
        return maskEl;
      }
    }

    return null;
  }

  function handleDrag(e: PointerEvent): void {
    const mask: HTMLElement | null =
      getMaskForArea() || ((e.target as HTMLElement).id === "mask" ? (e.target as HTMLElement) : null);
    if (!mask) {
      return;
    }

    isDragged = true;

    const rect = mask.getBoundingClientRect();
    const width = rect.width || mask.clientWidth || 1;
    const height = rect.height || mask.clientHeight || 1;
    const newX = (e.clientX - rect.left - dragFrom.x) / width;
    const newY = (e.clientY - rect.top - dragFrom.y) / height;
    area.x = Math.min(Math.max(newX, 0), 1 - area.w);
    area.y = Math.min(Math.max(newY, 0), 1 - area.h);
  }

  function handleResize(e: PointerEvent): void {
    let mask: HTMLElement | null = getMaskForArea();
    if (!mask && (e.target as HTMLElement).id === "mask") {
      mask = e.target as HTMLElement;
    }
    if (!mask) {
      return;
    }

    if (e.pointerType === "touch") {
      const rect = mask.getBoundingClientRect();
      area.w = (e.clientX - rect.left) / rect.width - area.x;
      area.h = (e.clientY - rect.top) / rect.height - area.y;
    } else {
      if ((e.target as HTMLElement).id === "mask") {
        area.w = e.offsetX / mask.clientWidth - area.x;
        area.h = e.offsetY / mask.clientHeight - area.y;
      } else {
        const rect = mask.getBoundingClientRect();
        area.w = (e.clientX - rect.left) / mask.clientWidth - area.x;
        area.h = (e.clientY - rect.top) / mask.clientHeight - area.y;
      }
    }
  }

  function handleResizeCell(e: PointerEvent): void {
    let mask: HTMLElement | null = getMaskForArea();
    if (!mask && (e.target as HTMLElement).id === "mask") {
      mask = e.target as HTMLElement;
    }
    if (!mask) {
      return;
    }

    let positionX: number;
    if ((e.target as HTMLElement).id === "mask") {
      positionX = e.offsetX / mask.clientWidth;
    } else {
      const rect = mask.getBoundingClientRect();
      positionX = (e.clientX - rect.left) / mask.clientWidth;
    }

    if (positionX > area.x) {
      area.cell_w = positionX - area.x;
      area.cell_count = getCellCountFromArea(area);
    }
  }

  export { rootRef, area, field };
</script>

<div
  bind:this={rootRef}
  class="group absolute overflow-visible {isDragged ? 'z-[1]' : ''}"
  style:top={positionStyle.top}
  style:left={positionStyle.left}
  style:width={positionStyle.width}
  style:height={positionStyle.height}
  onpointerdown={(e) => {
    e.stopPropagation();
    handleDragStart(e);
  }}
  ondblclick={(e) => {
    e.stopPropagation();
    handleDoubleClick();
  }}
>
  {#if isSelected || isDraw}
    <div
      class="border-1.5 pointer-events-none absolute top-0 right-0 bottom-0 left-0 border {borderColors[
        submitterIndex
      ]}"
    ></div>
  {/if}
  {#if !field?.required}
    <div
      class="pointer-events-none absolute top-0 right-0 bottom-0 left-0 border border-dashed {borderColors[
        submitterIndex
      ]}"
    ></div>
  {/if}
  {#if field?.type === "cells" && (isSelected || isDraw)}
    <div class="absolute top-0 right-0 bottom-0 left-0">
      {#each cells as cellW, index (index)}
        <div
          class="absolute top-0 bottom-0 border-r {borderColors[submitterIndex]}"
          style:left="{(cellW / area.w) * 100}%"
        >
          {#if index === 0 && editable}
            <span
              class="absolute -bottom-1 z-10 h-2.5 w-2.5 cursor-ew-resize rounded-full border border-gray-400 bg-white shadow-md"
              style="left: -4px"
              onpointerdown={(e) => {
                e.stopPropagation();
                handleResizeCellStart();
              }}
            ></span>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  {#if field?.type}
    <div
      class="absolute overflow-visible rounded-t border bg-white whitespace-nowrap group-hover:z-10 group-hover:flex {isNameFocus ||
      isSelected
        ? 'z-10 flex'
        : 'invisible'}"
      style="top: -25px; height: 25px"
      onmousedown={(e) => e.stopPropagation()}
      onpointerdown={(e) => e.stopPropagation()}
    >
      {#if field}
        <FieldSubmitter
          bind:value={field.submitter_id}
          class="border-r"
          compact={true}
          editable={editable && !defaultField}
          menuClasses="dropdown-content bg-white menu menu-xs p-2 shadow rounded-box w-52 rounded-t-none -left-[1px]"
          submitters={template.value.submitters as unknown as Record<string, unknown>[]}
          onValueChange={handleSubmitterUpdate}
          onclick={() => focusArea(area)}
        />
        <FieldType
          bind:value={field.type}
          buttonWidth={27}
          editable={editable && !defaultField}
          buttonClasses="px-1"
          menuClasses="bg-white rounded-t-none"
          onValueChange={handleTypeUpdate}
          onclick={() => focusArea(area)}
        />
      {/if}
      {#if field && (field.type !== "checkbox" || field.name)}
        <span
          bind:this={name}
          contenteditable={editable && !defaultField}
          dir="auto"
          class="block cursor-text pr-1 outline-none"
          style="min-width: 2px"
          onkeydown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              onNameEnter();
            }
          }}
          onfocus={onNameFocus}
          onblur={onNameBlur}
        >
          {optionIndexText}
          {field.name || defaultName}
        </span>
      {/if}
      {#if isNameFocus && field && !["checkbox", "phone"].includes(field.type)}
        <div class="ml-1.5 flex items-center gap-1.5 pr-2">
          <input
            id="required-checkbox-{field.id}"
            bind:checked={field.required}
            type="checkbox"
            class="toggle toggle-xs"
            onmousedown={(e) => e.preventDefault()}
          />
        </div>
      {:else if editable}
        <button
          class="pr-1"
          title="Remove"
          onclick={(e) => {
            e.preventDefault();
            onRemove?.();
          }}
        >
          <SvgIcon name="x" class="h-4 w-4" />
        </button>
      {/if}
    </div>
  {/if}

  <div
    class="flex h-full w-full items-center {bgColors[submitterIndex]} {isValueInput || isCheckboxInput || isSelectInput
      ? 'bg-opacity-50'
      : 'bg-opacity-80'} {!isDefaultValuePresent && !isValueInput && !isSelectInput ? 'justify-center' : ''}"
  >
    {#if field}
      <span class="flex h-full items-center justify-center space-x-1 {isWFullType ? 'w-full' : ''}">
        {#if isDefaultValuePresent || isValueInput || isSelectInput || (field.areas?.length && field.type !== "checkbox")}
          <div
            bind:this={textContainer}
            class="flex h-full min-w-0 flex-1 items-center px-0.5 {isWFullType ? 'h-full w-full' : ''}"
            style={fontStyle}
          >
            <div
              class="flex min-w-0 flex-1 items-center {isWFullType ? 'h-full w-full' : ''}"
              style:color={field.preferences?.color}
            >
              {#if field.type === "checkbox" && field.default_value}
                <SvgIcon
                  name="check-circle"
                  class="mx-auto aspect-square flex-shrink-0 {areaWiderThanHigh
                    ? '!h-full !w-auto'
                    : '!h-auto !w-full'}"
                />
              {:else if (field.type === "radio" || field.type === "multiple") && hasMultipleAreas}
                {#if field.type === "multiple" ? Array.isArray(field.default_value) && field.default_value.includes(buildAreaOptionValue(area)) : buildAreaOptionValue(area) === field.default_value}
                  <SvgIcon
                    name="check-circle"
                    class="mx-auto aspect-square flex-shrink-0 {areaWiderThanHigh
                      ? '!h-full !w-auto'
                      : '!h-auto !w-full'}"
                  />
                {/if}
              {:else if field.type === "number" && field.default_value != null && String(field.default_value) !== ""}
                <span class="whitespace-pre-wrap">{formatNumber(field.default_value, field.preferences?.format)}</span>
              {:else if isDatePlaceholder}
                {t("signing.signing_date")}
              {:else if field.type === "cells" && field.default_value}
                <div class="flex w-full items-center">
                  {#each String(field.default_value) as char, index (index)}
                    <div
                      class="flex-none text-center"
                      style:width="{area.w && effectiveCellW ? (effectiveCellW / area.w) * 100 : 0}%"
                    >
                      {char}
                    </div>
                  {/each}
                </div>
              {:else if isSelectInput && field.default_value}
                <span class="whitespace-pre-wrap">{field.default_value}</span>
              {:else if isValueInput}
                <span
                  bind:this={defaultValue}
                  contenteditable={editable && !defaultField}
                  dir="auto"
                  class="min-w-0 flex-1 whitespace-pre-wrap outline-none before:text-base-content/30 empty:before:content-[attr(data-placeholder)] {editable
                    ? 'cursor-text'
                    : ''}"
                  data-placeholder={field.type === "date"
                    ? field.preferences?.format || t("fields.type_value")
                    : t("fields.type_value")}
                  onblur={onDefaultValueBlur}
                  onfocus={() => focusArea(area)}
                  onpaste={onDefaultValuePaste}
                  onkeydown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      onDefaultValueEnter();
                    }
                  }}
                >
                  {field.type === "date" && field.default_value
                    ? formatDateByPattern(String(field.default_value), field.preferences?.format || "DD/MM/YYYY")
                    : field.default_value}
                </span>
              {/if}
            </div>
          </div>
        {:else if field.type}
          <SvgIcon name={fieldIcons[field.type]} width="100%" height="100%" class="max-h-10 flex-shrink-0 opacity-50" />
        {/if}
      </span>
    {/if}
  </div>

  <div bind:this={touchTarget} class="absolute top-0 right-0 bottom-0 left-0 cursor-pointer"></div>
  {#if field?.type && editable}
    <span
      class="absolute -right-1 -bottom-1 h-4 w-4 cursor-nwse-resize rounded-full border border-gray-400 bg-white shadow-md md:h-2.5 md:w-2.5"
      onpointerdown={(e) => {
        e.stopPropagation();
        handleResizeStart(e);
      }}
    ></span>
  {/if}
</div>
