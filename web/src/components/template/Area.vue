<template>
  <div
    ref="rootRef"
    class="group absolute overflow-visible"
    :class="{ 'z-[1]': isDragged }"
    :style="positionStyle"
    @pointerdown.stop="handleDragStart"
    @dblclick.stop="handleDoubleClick"
  >
    <div
      v-if="isSelected || isDraw"
      class="border-1.5 pointer-events-none absolute top-0 right-0 bottom-0 left-0 border"
      :class="borderColors[submitterIndex]"
    />
    <div
      v-if="!field?.required"
      class="pointer-events-none absolute top-0 right-0 bottom-0 left-0 border border-dashed"
      :class="borderColors[submitterIndex]"
    />
    <div v-if="field?.type === 'cells' && (isSelected || isDraw)" class="absolute top-0 right-0 bottom-0 left-0">
      <div
        v-for="(cellW, index) in cells"
        :key="index"
        class="absolute top-0 bottom-0 border-r"
        :class="borderColors[submitterIndex]"
        :style="{ left: (cellW / area.w) * 100 + '%' }"
      >
        <span
          v-if="index === 0 && editable"
          class="absolute -bottom-1 z-10 h-2.5 w-2.5 cursor-ew-resize rounded-full border border-gray-400 bg-white shadow-md"
          style="left: -4px"
          @pointerdown.stop="handleResizeCellStart"
        />
      </div>
    </div>

    <div
      v-if="field?.type"
      class="absolute overflow-visible rounded-t border bg-white whitespace-nowrap group-hover:z-10 group-hover:flex"
      :class="{ 'z-10 flex': isNameFocus || isSelected, invisible: !isNameFocus && !isSelected }"
      style="top: -25px; height: 25px"
      @mousedown.stop
      @pointerdown.stop
    >
      <FieldSubmitter
        v-model="field.submitter_id"
        class="border-r"
        :compact="true"
        :editable="editable && !defaultField"
        :menu-classes="'dropdown-content bg-white menu menu-xs p-2 shadow rounded-box w-52 rounded-t-none -left-[1px]'"
        :submitters="template.submitters"
        @update:model-value="[save(), emit('select-submitter', $event)]"
        @click="focusArea(area)"
      />
      <FieldType
        v-model="field.type"
        :button-width="27"
        :editable="editable && !defaultField"
        :button-classes="'px-1'"
        :menu-classes="'bg-white rounded-t-none'"
        @update:model-value="[maybeUpdateOptions(), save()]"
        @click="focusArea(area)"
      />
      <span
        v-if="field.type !== 'checkbox' || field.name"
        ref="name"
        :contenteditable="editable && !defaultField"
        dir="auto"
        class="block cursor-text pr-1 outline-none"
        style="min-width: 2px"
        @keydown.enter.prevent="onNameEnter"
        @focus="onNameFocus"
        @blur="onNameBlur"
      >
        {{ optionIndexText }} {{ field.name || defaultName }}
      </span>
      <div
        v-if="isNameFocus && !['checkbox', 'phone'].includes(field.type)"
        class="ml-1.5 flex items-center gap-1.5 pr-2"
      >
        <input
          :id="`required-checkbox-${field.id}`"
          v-model="field.required"
          type="checkbox"
          class="toggle toggle-xs"
          @mousedown.prevent
        />
      </div>
      <button v-else-if="editable" class="pr-1" title="Remove" @click.prevent="$emit('remove')">
        <SvgIcon name="x" class="h-4 w-4" />
      </button>
    </div>

    <div
      class="flex h-full w-full items-center"
      :class="[
        bgColors[submitterIndex],
        isValueInput || isCheckboxInput || isSelectInput ? 'bg-opacity-50' : 'bg-opacity-80',
        !isDefaultValuePresent && !isValueInput && !isSelectInput ? 'justify-center' : ''
      ]"
    >
      <span v-if="field" class="flex h-full items-center justify-center space-x-1" :class="{ 'w-full': isWFullType }">
        <div
          v-if="isDefaultValuePresent || isValueInput || isSelectInput || (field.areas?.length && field.type !== 'checkbox')"
          ref="textContainer"
          class="flex items-center px-0.5 flex-1 min-w-0 h-full"
          :class="{ 'w-full h-full': isWFullType }"
          :style="fontStyle"
        >
          <div
            class="flex items-center flex-1 min-w-0"
            :class="{ 'w-full h-full': isWFullType }"
            :style="{ color: field.preferences?.color }"
          >
            <SvgIcon
              v-if="field.type === 'checkbox' && field.default_value"
              name="check-circle"
              class="aspect-square mx-auto flex-shrink-0"
              :class="areaWiderThanHigh ? '!w-auto !h-full' : '!w-full !h-auto'"
            />
            <template v-else-if="(field.type === 'radio' || field.type === 'multiple') && hasMultipleAreas">
              <SvgIcon
                v-if="field.type === 'multiple' ? (Array.isArray(field.default_value) && field.default_value.includes(buildAreaOptionValue(area))) : (buildAreaOptionValue(area) === field.default_value)"
                name="check-circle"
                class="aspect-square mx-auto flex-shrink-0"
                :class="areaWiderThanHigh ? '!w-auto !h-full' : '!w-full !h-auto'"
              />
            </template>
            <span
              v-else-if="field.type === 'number' && (field.default_value != null && String(field.default_value) !== '')"
              class="whitespace-pre-wrap"
            >{{ formatNumber(field.default_value, field.preferences?.format) }}</span>
            <span v-else-if="isDatePlaceholder">{{ t('signing.signing_date') }}</span>
            <div
              v-else-if="field.type === 'cells' && field.default_value"
              class="w-full flex items-center"
            >
              <div
                v-for="(char, index) in String(field.default_value)"
                :key="index"
                class="text-center flex-none"
                :style="{ width: (area.w && effectiveCellW ? (effectiveCellW / area.w * 100) : 0) + '%' }"
              >
                {{ char }}
              </div>
            </div>
            <span
              v-else-if="isSelectInput && field.default_value"
              class="whitespace-pre-wrap"
            >{{ field.default_value }}</span>
            <span
              v-else-if="isValueInput"
              ref="defaultValue"
              :contenteditable="editable && !defaultField"
              dir="auto"
              class="whitespace-pre-wrap outline-none flex-1 min-w-0 empty:before:content-[attr(data-placeholder)] before:text-base-content/30"
              :class="{ 'cursor-text': editable }"
              :data-placeholder="field.type === 'date' ? (field.preferences?.format || t('fields.type_value')) : t('fields.type_value')"
              @blur="onDefaultValueBlur"
              @focus="focusArea(area)"
              @paste.prevent="onDefaultValuePaste"
              @keydown.enter.prevent="onDefaultValueEnter"
            >{{ field.type === 'date' && field.default_value ? formatDateByPattern(String(field.default_value), field.preferences?.format || 'DD/MM/YYYY') : field.default_value }}</span>
          </div>
        </div>
        <SvgIcon
          v-else
          :name="fieldIcons[field.type]"
          width="100%"
          height="100%"
          class="max-h-10 opacity-50 flex-shrink-0"
        />
      </span>
    </div>

    <div ref="touchTarget" class="absolute top-0 right-0 bottom-0 left-0 cursor-pointer"></div>
    <span
      v-if="field?.type && editable"
      class="absolute -right-1 -bottom-1 h-4 w-4 cursor-nwse-resize rounded-full border border-gray-400 bg-white shadow-md md:h-2.5 md:w-2.5"
      @pointerdown.stop="handleResizeStart"
    ></span>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, nextTick, type Ref, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import FieldSubmitter from "@/components/field/Submitter.vue";
import FieldType from "@/components/field/Type.vue";
import { bgColors, borderColors, fieldIcons, fieldNames, subNames } from "@/components/field/constants.ts";
import type { Template } from "@/models/index";
import type { Area, Field } from "@/models/template";
import { formatDateByPattern } from "@/utils/time";

const { t } = useI18n();

interface Props {
  area: Area;
  isDraw?: boolean;
  defaultField?: Field | null;
  editable?: boolean;
  field: Field | null;
}

const props = withDefaults(defineProps<Props>(), {
  isDraw: false,
  defaultField: null,
  editable: true
});

interface Emits {
  "start-resize": [direction: "nwse" | "ew"];
  "stop-resize": [];
  "start-drag": [];
  "stop-drag": [];
  remove: [];
  "select-submitter": [submitterId: string];
}

const emit = defineEmits<Emits>();

const template = inject<Ref<Template>>("template");
const selectedAreaRef = inject<Ref<Area | null>>("selectedAreaRef");
const save = inject<() => void>("save");

if (!template || !selectedAreaRef || !save) {
  throw new Error("Required injectables are missing in Area component");
}

const rootRef = ref<HTMLElement | null>(null);
const name = ref<HTMLElement | null>(null);
const textContainer = ref<HTMLElement | null>(null);
const defaultValue = ref<HTMLElement | null>(null);
const touchTarget = ref<HTMLElement | null>(null);

const isDragged = ref(false);

const isValueInput = computed(() =>
  props.field ? ["text", "number", "date", "cells"].includes(props.field.type) : false
);
const isCheckboxInput = computed(() => props.field?.type === "checkbox");
const isSelectInput = computed(() =>
  props.field ? ["select", "radio", "multiple"].includes(props.field.type) : false
);
const isDefaultValuePresent = computed(() => {
  if (!props.field) return false;
  const v = props.field.default_value;
  if (props.field.type === "checkbox") return !!v;
  if (props.field.type === "multiple") return Array.isArray(v) ? v.length > 0 : !!v;
  if (props.field.type === "radio" || props.field.type === "select") return v != null && v !== "";
  if (props.field.type === "number") return v != null && v !== "" && !Number.isNaN(Number(v));
  return v != null && v !== "";
});
const isWFullType = computed(() =>
  props.field ? ["text", "number", "date", "cells", "select"].includes(props.field.type) : false
);

const areaWiderThanHigh = computed(() => props.area.w > props.area.h);

const hasMultipleAreas = computed(
  () => (props.field?.areas?.length ?? 0) > 1
);

const isDatePlaceholder = computed(
  () => props.field?.default_value === "\u007b\u007bdate\u007d\u007d"
);

function buildAreaOptionValue(area: Area): string {
  if (!area.option_id || !props.field?.options) return "";
  const option = props.field.options.find((o: { id?: string; value?: string }) => o.id === area.option_id);
  return (option && "value" in option ? option.value : "") ?? "";
}

function formatNumber(value: number | string, format?: string): string {
  const num = typeof value === "string" ? parseFloat(value) : value;
  if (Number.isNaN(num)) return String(value);
  if (!format || format === "none") return String(num);
  if (format === "comma") return new Intl.NumberFormat("en-US").format(num);
  if (format === "dot") return new Intl.NumberFormat("de-DE").format(num);
  if (format === "space") return new Intl.NumberFormat("fr-FR").format(num);
  if (format === "usd")
    return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(num);
  if (format === "eur")
    return new Intl.NumberFormat("de-DE", { style: "currency", currency: "EUR" }).format(num);
  if (format === "gbp")
    return new Intl.NumberFormat("en-GB", { style: "currency", currency: "GBP" }).format(num);
  return String(num);
}

const fontStyle = computed(() => {
  const field = props.field;
  if (!field?.preferences) return {};
  const style: Record<string, string> = {};
  if (field.preferences.font) style.fontFamily = field.preferences.font;
  if (field.preferences.font_size) style.fontSize = String(field.preferences.font_size) + "px";
  if (field.preferences.font_type) {
    if (field.preferences.font_type === "bold") style.fontWeight = "bold";
    else if (field.preferences.font_type === "italic") style.fontStyle = "italic";
    else if (field.preferences.font_type === "bold_italic") {
      style.fontWeight = "bold";
      style.fontStyle = "italic";
    }
  }
  if (field.preferences.align) style.textAlign = field.preferences.align;
  if (field.preferences.valign) {
    style.alignItems =
      field.preferences.valign === "top"
        ? "flex-start"
        : field.preferences.valign === "bottom"
          ? "flex-end"
          : "center";
  }
  return style;
});
const isNameFocus = ref(false);
const textOverflowChars = ref(0);
const dragFrom = ref({ x: 0, y: 0 });
const pointerMode = ref<"drag" | "resize" | "resize-cell" | null>(null);
const lastClickTime = ref(0);

const defaultName = computed(() => {
  if (!props.field) {
    return "Field";
  }

  if (props.field.type === "payment" && props.field.preferences?.price) {
    const { price, currency } = props.field.preferences;
    const formattedPrice = new Intl.NumberFormat([], {
      style: "currency",
      currency
    }).format(price ?? 0);
    return `${fieldNames[props.field.type]} ${formattedPrice}`;
  }

  const partyName = subNames[submitterIndex.value]?.replace(" Party", "") || "First";
  const typeName = fieldNames[props.field.type] || "Field";
  if (!props.field) {
    return `${partyName} ${typeName} 1`;
  }
  const sameTypeAndPartyFields = template.value.fields.filter(
    (f: any) => f.type === props.field?.type && f.submitter_id === props.field?.submitter_id && f.id !== props.field?.id
  );
  const fieldNumber = sameTypeAndPartyFields.length + 1;

  return `${partyName} ${typeName} ${fieldNumber}`;
});

const optionIndexText = computed(() => {
  const opts = props.field?.options as Array<{ id?: string }> | undefined;
  if (!props.area.option_id || !opts?.length) return "";
  const idx = opts.findIndex((o) => o && typeof o === "object" && "id" in o && o.id === props.area.option_id);
  return idx >= 0 ? `${idx + 1}.` : "";
});

function getEffectiveCellW(area: Area): number {
  if (area.cell_w != null && area.cell_w > 0) return area.cell_w;
  if (area.w <= 0) return 0;
  if (area.h > 0) {
    const denom = Math.floor(area.w / area.h);
    return denom > 0 ? (area.w * 2) / denom : area.w / 5;
  }
  return area.w / 5;
}

/** Returns number of cells for a cells-type area (same formula as cells computed). */
function getCellCountFromArea(area: Area): number {
  const cellWidth = getEffectiveCellW(area);
  if (!cellWidth || cellWidth <= 0 || area.w <= 0) return 0;
  let currentWidth = 0;
  let count = 0;
  while (currentWidth + (cellWidth + cellWidth / 4) < area.w) {
    currentWidth += cellWidth;
    count++;
  }
  return Math.max(count, 1);
}

const effectiveCellW = computed(() => getEffectiveCellW(props.area));

const cells = computed(() => {
  const cellsList: number[] = [];
  const cellWidth = getEffectiveCellW(props.area);
  if (!cellWidth || cellWidth <= 0) {
    return cellsList;
  }

  let currentWidth = 0;
  while (currentWidth + (cellWidth + cellWidth / 4) < props.area.w) {
    currentWidth += cellWidth;
    cellsList.push(currentWidth);
  }
  return cellsList;
});

const submitter = computed(() => {
  return template.value.submitters.find((s: any) => s.id === props.field?.submitter_id);
});

const submitterIndex = computed(() => {
  if (!submitter.value) {
    return 0;
  }
  return template.value.submitters.indexOf(submitter.value);
});

const isSelected = computed(() => {
  return selectedAreaRef?.value === props.area;
});

const positionStyle = computed(() => {
  const { x, y, w, h } = props.area;
  return {
    top: y * 100 + "%",
    left: x * 100 + "%",
    width: w * 100 + "%",
    height: h * 100 + "%"
  };
});

watch(
  () => props.field?.default_value,
  () => {
    if (
      props.field?.type === "text" &&
      props.field.default_value &&
      textContainer.value &&
      (textOverflowChars.value === 0 || textOverflowChars.value - 4 > props.field.default_value.length)
    ) {
      nextTick(() => {
        const el = document.querySelector(".group.absolute.overflow-visible") as HTMLElement;
        if (el && textContainer.value && props.field?.default_value) {
          textOverflowChars.value =
            el.clientHeight < textContainer.value.clientHeight ? props.field.default_value.length : 0;
        } else {
          textOverflowChars.value = 0;
        }
      });
    }
  }
);

watch(
  () => props.field?.submitter_id,
  (newSubmitterId, oldSubmitterId) => {
    if (newSubmitterId !== oldSubmitterId && isDefaultName(props.field?.name || "")) {
      if (props.field) {
        props.field.name = "";
        save();
      }
    }
  }
);

function isDefaultName(name: string): boolean {
  if (!name) {
    return true;
  }
  const pattern = /^(First|Second|Third|Fourth|Fifth|Sixth|Seventh|Eighth|Ninth|Tenth)\s+\w+\s+\d+$/;
  return pattern.test(name);
}

function onNameFocus(): void {
  if (selectedAreaRef) selectedAreaRef.value = props.area;
  isNameFocus.value = true;
  if (name.value) {
    name.value.style.minWidth = name.value.clientWidth + "px";
  }

  if (!props.field?.name) {
    setTimeout(() => {
      if (name.value) {
        name.value.innerText = " ";
      }
    }, 1);
  }
}

function onNameBlur(): void {
  const text = name.value?.innerText.trim() || "";
  isNameFocus.value = false;
  if (name.value) {
    name.value.style.minWidth = "";
  }

  if (props.field) {
    props.field.name = text || "";
    if (!text && name.value) {
      name.value.innerText = defaultName.value;
    }
    save?.();
  }
}

function onNameEnter(): void {
  name.value?.blur();
}

function onDefaultValueBlur(): void {
  const el = defaultValue.value;
  if (!el || !props.field) return;
  const text = el.innerText.trim();
  if (props.field.default_value !== text) {
    (props.field as { default_value?: string }).default_value = text;
    save?.();
  }
}

function onDefaultValuePaste(e: ClipboardEvent): void {
  e.preventDefault();
  const text = e.clipboardData?.getData("text") ?? "";
  document.execCommand("insertText", false, text);
}

function onDefaultValueEnter(): void {
  defaultValue.value?.blur();
}

function focusArea(a: Area): void {
  if (selectedAreaRef) selectedAreaRef.value = a;
}

function maybeUpdateOptions(): void {
  if (!props.field) {
    return;
  }

  if (props.field.type !== "cells") {
    delete props.field.default_value;
  }

  if (!["radio", "multiple", "select"].includes(props.field.type)) {
    delete props.field.options;
  }

  if (["select", "multiple", "radio"].includes(props.field.type)) {
    type OptionItem = { id: string; value: string };
    const opts = (props.field as { options?: OptionItem[] }).options;
    (props.field as { options: OptionItem[] }).options = opts?.length ? opts : [{ value: "", id: crypto.randomUUID() }];
  }

  (props.field.areas || []).forEach((area: Area) => {
    if (props.field?.type === "cells") {
      const denom = area.h > 0 ? Math.floor(area.w / area.h) : 0;
      if (denom > 0) {
        area.cell_w = (area.w * 2) / denom;
      } else if (area.w > 0 && (!area.cell_w || area.cell_w <= 0)) {
        area.cell_w = area.w / 5;
      }
      area.cell_count = getCellCountFromArea(area);
    } else {
      delete area.cell_w;
      delete area.cell_count;
    }
  });
}

// Handle double click to select submitter
function handleDoubleClick(): void {
  if (props.field) {
    emit("select-submitter", props.field.submitter_id);
  }
}

// Unified pointer event handlers
function handleDragStart(e: PointerEvent): void {
  if (selectedAreaRef) selectedAreaRef.value = props.area;

  if (!props.editable) {
    return;
  }

  // Check for double click to prevent drag
  const now = Date.now();
  if (now - lastClickTime.value < 300) {
    lastClickTime.value = 0;
    return;
  }
  lastClickTime.value = now;

  const target = e.target as HTMLElement;
  if (target !== touchTarget.value && e.pointerType === "touch") {
    return;
  }

  if (e.pointerType === "touch") {
    name.value?.blur();
    e.preventDefault();
  }

  const el = rootRef.value || (e.target as HTMLElement);
  const rect = el.getBoundingClientRect();
  dragFrom.value = { x: e.clientX - rect.left, y: e.clientY - rect.top };
  pointerMode.value = "drag";

  document.addEventListener("pointermove", handlePointerMove);
  document.addEventListener("pointerup", handlePointerUp);
  emit("start-drag");
}

function handleResizeStart(e: PointerEvent): void {
  if (selectedAreaRef) selectedAreaRef.value = props.area;
  if (e.pointerType === "touch") {
    name.value?.blur();
    e.preventDefault();
  }

  pointerMode.value = "resize";
  document.addEventListener("pointermove", handlePointerMove);
  document.addEventListener("pointerup", handlePointerUp);
  emit("start-resize", "nwse");
}

function handleResizeCellStart(): void {
  pointerMode.value = "resize-cell";
  document.addEventListener("pointermove", handlePointerMove);
  document.addEventListener("pointerup", handlePointerUp);
  emit("start-resize", "ew");
}

function handlePointerMove(e: PointerEvent): void {
  if (pointerMode.value === "drag") {
    handleDrag(e);
  } else if (pointerMode.value === "resize") {
    handleResize(e);
  } else if (pointerMode.value === "resize-cell") {
    handleResizeCell(e);
  }
}

function handlePointerUp(): void {
  document.removeEventListener("pointermove", handlePointerMove);
  document.removeEventListener("pointerup", handlePointerUp);

  if (pointerMode.value === "drag") {
    if (isDragged.value) {
      save?.();
    }
    isDragged.value = false;
    emit("stop-drag");
  } else if (pointerMode.value === "resize" || pointerMode.value === "resize-cell") {
    emit("stop-resize");
    save?.();
  }

  pointerMode.value = null;
}

function getMaskForArea(): HTMLElement | null {
  let current: HTMLElement | null = touchTarget.value;
  if (!current) return null;

  while (current && current.parentElement) {
    current = current.parentElement;
    if (current.classList.contains("relative") && current.classList.contains("cursor-crosshair")) {
      const mask = current.querySelector("#mask") as HTMLElement | null;
      if (mask && typeof mask.clientWidth === "number" && mask.clientWidth > 0) {
        return mask;
      }
      if (!mask && process.env.NODE_ENV === "development") {
        console.warn("[Area] getMaskForArea: #mask not found in page container");
      }
      return mask;
    }
  }

  return null;
}

function handleDrag(e: PointerEvent): void {
  const mask: HTMLElement | null =
    getMaskForArea() || ((e.target as HTMLElement).id === "mask" ? (e.target as HTMLElement) : null);
  if (!mask) return;

  isDragged.value = true;

  const rect = mask.getBoundingClientRect();
  const width = rect.width || mask.clientWidth || 1;
  const height = rect.height || mask.clientHeight || 1;
  const newX = (e.clientX - rect.left - dragFrom.value.x) / width;
  const newY = (e.clientY - rect.top - dragFrom.value.y) / height;
  props.area.x = Math.min(Math.max(newX, 0), 1 - props.area.w);
  props.area.y = Math.min(Math.max(newY, 0), 1 - props.area.h);
}

function handleResize(e: PointerEvent): void {
  let mask: HTMLElement | null = getMaskForArea();
  if (!mask && (e.target as HTMLElement).id === "mask") {
    mask = e.target as HTMLElement;
  }
  if (!mask) return;

  if (e.pointerType === "touch") {
    const rect = mask.getBoundingClientRect();
    props.area.w = (e.clientX - rect.left) / rect.width - props.area.x;
    props.area.h = (e.clientY - rect.top) / rect.height - props.area.y;
  } else {
    if ((e.target as HTMLElement).id === "mask") {
      props.area.w = e.offsetX / mask.clientWidth - props.area.x;
      props.area.h = e.offsetY / mask.clientHeight - props.area.y;
    } else {
      const rect = mask.getBoundingClientRect();
      props.area.w = (e.clientX - rect.left) / mask.clientWidth - props.area.x;
      props.area.h = (e.clientY - rect.top) / mask.clientHeight - props.area.y;
    }
  }
}

function handleResizeCell(e: PointerEvent): void {
  let mask: HTMLElement | null = getMaskForArea();
  if (!mask && (e.target as HTMLElement).id === "mask") {
    mask = e.target as HTMLElement;
  }
  if (!mask) return;

  let positionX: number;
  if ((e.target as HTMLElement).id === "mask") {
    positionX = e.offsetX / mask.clientWidth;
  } else {
    const rect = mask.getBoundingClientRect();
    positionX = (e.clientX - rect.left) / mask.clientWidth;
  }

  if (positionX > props.area.x) {
    props.area.cell_w = positionX - props.area.x;
    props.area.cell_count = getCellCountFromArea(props.area);
  }
}

defineExpose({
  get area() {
    return props.area;
  },
  get field() {
    return props.field;
  }
});
</script>
