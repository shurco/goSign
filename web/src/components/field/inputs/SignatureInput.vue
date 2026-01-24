<template>
  <div class="signature-input">
    <div
      class="rounded-lg border border-[var(--color-base-300)] bg-white p-3"
      :class="{ 'opacity-60 pointer-events-none': disabled }"
    >
      <div class="mb-2 flex items-center justify-between gap-3">
        <div class="text-sm font-medium text-[--color-base-content]">
          {{ mode === "initials" ? t('fields.initials') : t('fields.signature') }}
        </div>
        <button
          type="button"
          class="btn btn-ghost btn-xs"
          :disabled="disabled"
          @click="clear"
        >
          {{ t('common.clear') }}
        </button>
      </div>

      <div class="relative">
        <canvas
          ref="canvasEl"
          class="w-full rounded-md border border-dashed border-[var(--color-base-300)] bg-[--color-base-100]"
          style="height: 160px"
          @pointerdown="onPointerDown"
          @pointermove="onPointerMove"
          @pointerup="onPointerUp"
          @pointercancel="onPointerUp"
          @pointerleave="onPointerUp"
        />
        <div
          v-if="!hasValue"
          class="pointer-events-none absolute inset-0 flex items-center justify-center text-sm text-[--color-base-content]/60"
        >
          {{
            placeholder ||
            (mode === "initials" ? t("signing.drawInitials") : t("signing.drawSignature"))
          }}
        </div>
      </div>

      <div v-if="error" class="mt-2 text-sm text-error">
        {{ error }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

interface Props {
  modelValue?: string;
  mode?: "signature" | "initials";
  placeholder?: string;
  required?: boolean;
  disabled?: boolean;
  error?: string;
}

interface Emits {
  (e: "update:modelValue", value: string): void;
  (e: "blur"): void;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: "",
  mode: "signature",
  placeholder: "",
  required: false,
  disabled: false,
  error: ""
});

const emit = defineEmits<Emits>();
const { t } = useI18n();

const canvasEl = ref<HTMLCanvasElement | null>(null);
const isDrawing = ref(false);
const last = ref({ x: 0, y: 0 });

const hasValue = computed(() => !!props.modelValue && props.modelValue.trim() !== "");

let resizeObserver: ResizeObserver | null = null;

function getCtx(): CanvasRenderingContext2D | null {
  const c = canvasEl.value;
  if (!c) return null;
  return c.getContext("2d");
}

function setupCanvasSize(): void {
  const c = canvasEl.value;
  if (!c) return;
  const ctx = getCtx();
  if (!ctx) return;

  // Make canvas crisp on high-DPI displays.
  const dpr = window.devicePixelRatio || 1;
  const cssWidth = c.getBoundingClientRect().width;
  const cssHeight = c.getBoundingClientRect().height;
  const width = Math.max(1, Math.floor(cssWidth * dpr));
  const height = Math.max(1, Math.floor(cssHeight * dpr));

  if (c.width !== width || c.height !== height) {
    c.width = width;
    c.height = height;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  }

  ctx.lineCap = "round";
  ctx.lineJoin = "round";
  ctx.strokeStyle = "#111827"; // gray-900
  ctx.lineWidth = props.mode === "initials" ? 3 : 2;
}

function clearCanvas(): void {
  const c = canvasEl.value;
  const ctx = getCtx();
  if (!c || !ctx) return;
  ctx.clearRect(0, 0, c.width, c.height);
}

function drawFromModelValue(dataUrl: string): void {
  if (!dataUrl) {
    clearCanvas();
    return;
  }
  const ctx = getCtx();
  const c = canvasEl.value;
  if (!ctx || !c) return;

  const img = new Image();
  img.onload = () => {
    clearCanvas();
    // Draw image to fit canvas area.
    const cssW = c.getBoundingClientRect().width;
    const cssH = c.getBoundingClientRect().height;
    ctx.drawImage(img, 0, 0, cssW, cssH);
  };
  img.src = dataUrl;
}

function getPoint(e: PointerEvent): { x: number; y: number } {
  const c = canvasEl.value!;
  const rect = c.getBoundingClientRect();
  return {
    x: e.clientX - rect.left,
    y: e.clientY - rect.top
  };
}

function onPointerDown(e: PointerEvent): void {
  if (props.disabled) return;
  if (!canvasEl.value) return;
  setupCanvasSize();

  const ctx = getCtx();
  if (!ctx) return;

  isDrawing.value = true;
  const p = getPoint(e);
  last.value = p;
  ctx.beginPath();
  ctx.moveTo(p.x, p.y);

  try {
    canvasEl.value.setPointerCapture(e.pointerId);
  } catch {
    // ignore
  }
}

function onPointerMove(e: PointerEvent): void {
  if (!isDrawing.value || props.disabled) return;
  const ctx = getCtx();
  if (!ctx) return;

  const p = getPoint(e);
  ctx.lineTo(p.x, p.y);
  ctx.stroke();
  last.value = p;
}

async function onPointerUp(): Promise<void> {
  if (!isDrawing.value) return;
  isDrawing.value = false;

  const c = canvasEl.value;
  if (!c) return;
  // Export on stroke end to avoid doing it every move.
  const dataUrl = c.toDataURL("image/png");
  emit("update:modelValue", dataUrl);
  // Important: wait a tick so parent v-model updates before validation runs.
  await nextTick();
  emit("blur");
}

function clear(): void {
  if (props.disabled) return;
  clearCanvas();
  emit("update:modelValue", "");
  // Same reason as onPointerUp: let v-model update before validating.
  nextTick().then(() => emit("blur"));
}

onMounted(() => {
  setupCanvasSize();
  drawFromModelValue(props.modelValue || "");

  if (canvasEl.value && typeof ResizeObserver !== "undefined") {
    resizeObserver = new ResizeObserver(() => {
      setupCanvasSize();
      // Re-draw current value after resize to keep it visible.
      drawFromModelValue(props.modelValue || "");
    });
    resizeObserver.observe(canvasEl.value);
  }
});

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
  resizeObserver = null;
});

watch(
  () => props.modelValue,
  (v) => {
    // If value changes externally (e.g., form reset), keep canvas in sync.
    setupCanvasSize();
    drawFromModelValue(v || "");
  }
);
</script>

<style scoped>
.signature-input canvas {
  touch-action: none;
}
</style>

