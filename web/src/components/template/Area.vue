<template>
  <div
    class="group absolute overflow-visible"
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
        @click="selectedAreaRef = area"
      />
      <FieldType
        v-model="field.type"
        :button-width="27"
        :editable="editable && !defaultField"
        :button-classes="'px-1'"
        :menu-classes="'bg-white rounded-t-none'"
        @update:model-value="[maybeUpdateOptions(), save()]"
        @click="selectedAreaRef = area"
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
        >{{ optionIndexText }} {{ field.name || defaultName }}</span
      >
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
        field?.default_value ? '' : 'justify-center',
        !field?.required ? 'opacity-50' : ''
      ]"
    >
      <span v-if="field" class="flex h-full items-center justify-center space-x-1">
        <div
          v-if="field?.default_value"
          :class="{ 'text-[1.5vw] lg:text-base': !textOverflowChars, 'text-[1.0vw] lg:text-xs': textOverflowChars }"
        >
          <div ref="textContainer" class="flex items-center px-0.5">
            <span class="whitespace-pre-wrap">{{ field.default_value }}</span>
          </div>
        </div>
        <SvgIcon v-else :name="fieldIcons[field.type]" width="100%" height="100%" class="max-h-10 opacity-50" />
      </span>
    </div>

    <div ref="touchTarget" class="absolute top-0 right-0 bottom-0 left-0 cursor-pointer" />
    <span
      v-if="field?.type && editable"
      class="absolute -right-1 -bottom-1 h-4 w-4 cursor-nwse-resize rounded-full border border-gray-400 bg-white shadow-md md:h-2.5 md:w-2.5"
      @pointerdown.stop="handleResizeStart"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, inject, nextTick, type Ref, ref, watch } from "vue";
import FieldSubmitter from "@/components/field/Submitter.vue";
import FieldType from "@/components/field/Type.vue";
import { bgColors, borderColors, fieldIcons, fieldNames, subNames } from "@/components/field/constants.ts";
import type { Template } from "@/models/index";
import type { Area, Field } from "@/models/template";

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

const name = ref<HTMLElement | null>(null);
const textContainer = ref<HTMLElement | null>(null);
const touchTarget = ref<HTMLElement | null>(null);

const isDragged = ref(false);
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
  if (props.area.option_id && props.field?.options) {
    return `${props.field.options.findIndex((o) => o.id === props.area.option_id) + 1}.`;
  }
  return "";
});

const cells = computed(() => {
  const cellsList: number[] = [];
  const cellWidth = props.area.cell_w;
  if (!cellWidth) {
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
  return selectedAreaRef.value === props.area;
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
  selectedAreaRef.value = props.area;
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
    save();
  }
}

function onNameEnter(): void {
  name.value?.blur();
}

function maybeUpdateOptions(): void {
  if (!props.field) {
    return;
  }

  delete props.field.default_value;

  if (!["radio", "multiple", "select"].includes(props.field.type)) {
    delete props.field.options;
  }

  if (["select", "multiple", "radio"].includes(props.field.type)) {
    props.field.options ||= [{ value: "", id: crypto.randomUUID() }];
  }

  (props.field.areas || []).forEach((area: Area) => {
    if (props.field?.type === "cells") {
      area.cell_w = (area.w * 2) / Math.floor(area.w / area.h);
    } else {
      delete area.cell_w;
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
  selectedAreaRef.value = props.area;

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

  const rect = (e.target as HTMLElement).getBoundingClientRect();
  dragFrom.value = { x: e.clientX - rect.left, y: e.clientY - rect.top };
  pointerMode.value = "drag";

  document.addEventListener("pointermove", handlePointerMove);
  document.addEventListener("pointerup", handlePointerUp);
  emit("start-drag");
}

function handleResizeStart(e: PointerEvent): void {
  selectedAreaRef.value = props.area;
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
      save();
    }
    isDragged.value = false;
    emit("stop-drag");
  } else if (pointerMode.value === "resize" || pointerMode.value === "resize-cell") {
    emit("stop-resize");
    save();
  }

  pointerMode.value = null;
}

function handleDrag(e: PointerEvent): void {
  const mask = document.getElementById("mask");
  if (!mask) {
    return;
  }

  isDragged.value = true;

  if (e.pointerType === "touch") {
    const page = mask.previousElementSibling as HTMLElement;
    const rect = page.getBoundingClientRect();
    props.area.x = (dragFrom.value.x + e.clientX - rect.left) / rect.width;
    props.area.y = (dragFrom.value.y + e.clientY - rect.top) / rect.height;
  } else {
    if ((e.target as HTMLElement).id === "mask") {
      props.area.x = (e.layerX - dragFrom.value.x) / mask.clientWidth;
      props.area.y = (e.layerY - dragFrom.value.y) / mask.clientHeight;
    }
  }
}

function handleResize(e: PointerEvent): void {
  const mask = document.getElementById("mask");
  if (!mask) {
    return;
  }

  if (e.pointerType === "touch") {
    const page = mask.previousElementSibling as HTMLElement;
    const rect = page.getBoundingClientRect();
    props.area.w = (e.clientX - rect.left) / rect.width - props.area.x;
    props.area.h = (e.clientY - rect.top) / rect.height - props.area.y;
  } else {
    if ((e.target as HTMLElement).id === "mask") {
      props.area.w = e.layerX / mask.clientWidth - props.area.x;
      props.area.h = e.layerY / mask.clientHeight - props.area.y;
    }
  }
}

function handleResizeCell(e: PointerEvent): void {
  const mask = document.getElementById("mask");
  if (!mask || (e.target as HTMLElement).id !== "mask") {
    return;
  }

  const positionX = e.layerX / (mask.clientWidth - 1);
  if (positionX > props.area.x) {
    props.area.cell_w = positionX - props.area.x;
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
