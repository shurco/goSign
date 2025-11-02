<template>
  <div class="relative cursor-crosshair select-none" :style="drawField ? 'touch-action: none' : ''">
    <img
      ref="imageEl"
      loading="lazy"
      :src="`${props.image.url}/${props.image.filename}`"
      :width="width"
      :height="height"
      class="mb-4 rounded border border-[#e7e2df]"
      @load="onImageLoad"
    />
    <div class="absolute top-0 right-0 bottom-0 left-0" @pointerdown="onStartDraw">
      <FieldArea
        v-for="(item, i) in areas"
        :key="i"
        :ref="setAreaRefs"
        :area="item.area"
        :field="item.field"
        :editable="editable"
        :default-field="defaultFields.find((f) => f.name === item.field.name)"
        @start-resize="resizeDirection = $event"
        @stop-resize="resizeDirection = null"
        @start-drag="isMove = true"
        @stop-drag="isMove = false"
        @remove="$emit('remove-area', item.area)"
        @select-submitter="$emit('select-submitter', $event)"
      />
      <FieldArea v-if="newArea && drawField" :is-draw="true" :field="drawField" :area="newArea" />
    </div>
    <div
      v-show="resizeDirection || isMove || isDrag || newArea || drawField"
      id="mask"
      ref="mask"
      class="absolute top-0 right-0 bottom-0 left-0 z-10"
      :class="{
        'cursor-grab': isDrag || isMove,
        'cursor-nwse-resize': drawField || resizeDirection === 'nwse',
        'cursor-ew-resize': resizeDirection === 'ew'
      }"
      @pointermove="onPointermove"
      @pointerdown="onStartDraw"
      @dragover.prevent
      @drop="onDrop"
      @pointerup="onPointerup"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUpdate, ref } from "vue";
import type { PreviewImages, Submitters } from "@/models/index";
import type { Area, Field } from "@/models/template";
import FieldArea from "@/components/template/Area.vue";

interface Props {
  image: PreviewImages;
  areas?: any[];
  defaultFields?: Field[];
  allowDraw?: boolean;
  selectedSubmitter: Submitters;
  drawField?: Field | null;
  editable?: boolean;
  isDrag?: boolean;
  number: number;
}

const props = withDefaults(defineProps<Props>(), {
  areas: () => [],
  defaultFields: () => [],
  allowDraw: true,
  drawField: null,
  editable: true,
  isDrag: false
});

interface Emits {
  draw: [area: Area];
  "drop-field": [event: any];
  "remove-area": [area: Area];
  "select-submitter": [submitterId: string];
}

const emit = defineEmits<Emits>();

const areaRefs = ref<any[]>([]);
const isMove = ref(false);
const resizeDirection = ref<string | null>(null);
const newArea = ref<Area | null>(null);
const mask = ref<HTMLDivElement | null>(null);
const imageEl = ref<HTMLImageElement | null>(null);

const width = computed(() => props.image.metadata.width);
const height = computed(() => props.image.metadata.height);

onBeforeUpdate(() => {
  areaRefs.value = [];
});

function onImageLoad(e: Event): void {
  const target = e.target as HTMLImageElement;
  target.setAttribute("width", target.naturalWidth.toString());
  target.setAttribute("height", target.naturalHeight.toString());
}

function setAreaRefs(el: any): void {
  if (el) {
    areaRefs.value.push(el);
  }
}

function onDrop(e: DragEvent): void {
  e.preventDefault();
  if (!mask.value) {
    return;
  }

  const rect = mask.value.getBoundingClientRect();
  const x = e.clientX - rect.left;
  const y = e.clientY - rect.top;

  emit("drop-field", {
    x,
    y,
    maskW: mask.value.clientWidth,
    maskH: mask.value.clientHeight,
    page: props.number
  });
}

function onStartDraw(e: PointerEvent): void {
  if (!props.allowDraw || !props.drawField || !props.editable || !mask.value) {
    return;
  }

  nextTick(() => {
    if (!mask.value) {
      return;
    }

    const initialX = e.layerX / mask.value.clientWidth;
    const initialY = e.layerY / mask.value.clientHeight;

    newArea.value = {
      initialX,
      initialY,
      x: initialX,
      y: initialY,
      w: 0,
      h: 0,
      page: props.number,
      attachment_id: ""
    };
  });
}

function onPointermove(e: PointerEvent): void {
  if (!newArea.value || !mask.value) {
    return;
  }

  const dx = e.layerX / mask.value.clientWidth - (newArea.value.initialX || 0);
  const dy = e.layerY / mask.value.clientHeight - (newArea.value.initialY || 0);

  newArea.value.x = dx > 0 ? newArea.value.initialX || 0 : e.layerX / mask.value.clientWidth;
  newArea.value.y = dy > 0 ? newArea.value.initialY || 0 : e.layerY / mask.value.clientHeight;

  if (props.drawField?.type === "cells") {
    newArea.value.cell_w = newArea.value.h * (mask.value.clientHeight / mask.value.clientWidth);
  }

  newArea.value.w = Math.abs(dx);
  newArea.value.h = Math.abs(dy);
}

function onPointerup(): void {
  if (newArea.value) {
    const area: Area = {
      x: newArea.value.x,
      y: newArea.value.y,
      w: newArea.value.w,
      h: newArea.value.h,
      page: props.number,
      attachment_id: ""
    };

    if (newArea.value.cell_w) {
      area.cell_w = newArea.value.cell_w;
    }

    emit("draw", area);
  }

  newArea.value = null;
}

defineExpose({
  areaRefs
});
</script>
