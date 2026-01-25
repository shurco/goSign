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

      <!-- Tabs when multiple formats allowed (DocuSeal-style) -->
      <div
        v-if="tabs.length > 1"
        class="tabs tabs-boxed mb-3 flex gap-1 rounded-lg bg-[var(--color-base-200)] p-1"
        role="tablist"
      >
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          role="tab"
          :aria-selected="activeTab === tab.id"
          class="tab tab-sm flex-1"
          :class="{ 'tab-active': activeTab === tab.id }"
          :disabled="disabled"
          @click="activeTab = tab.id"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Drawn panel -->
      <div v-if="showDrawn" v-show="activeTab === 'drawn'" class="relative">
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
          v-if="!hasValue && activeTab === 'drawn'"
          class="pointer-events-none absolute inset-0 flex items-center justify-center text-sm text-[--color-base-content]/60"
        >
          {{
            placeholder ||
            (mode === "initials" ? t("signing.drawInitials") : t("signing.drawSignature"))
          }}
        </div>
      </div>

      <!-- Typed panel -->
      <div v-if="showTyped" v-show="activeTab === 'typed'" class="space-y-2">
        <input
          v-model="typedText"
          type="text"
          class="input input-bordered w-full"
          :placeholder="mode === 'initials' ? t('signing.typeInitials') : t('signing.typeSignature')"
          :disabled="disabled"
          :style="{ fontFamily: typedFontFamily }"
          @input="emitTypedAsImage"
          @blur="$emit('blur')"
        />
        <div
          v-if="typedText && typedPreviewUrl"
          class="rounded-md border border-[var(--color-base-300)] bg-[--color-base-100] p-2"
          style="min-height: 60px"
        >
          <img :src="typedPreviewUrl" alt="" class="max-h-20 w-full object-contain" style="font-family: cursive" />
        </div>
      </div>

      <!-- Upload panel -->
      <div v-if="showUpload" v-show="activeTab === 'upload'" class="space-y-2">
        <FileDropZone
          accept="image/*"
          :disabled="disabled"
          :selected-label="uploadFileName || (hasValue ? 'Image' : '')"
          @change="onUploadChange"
          @clear="clearUpload"
        />
        <div
          v-if="hasValue && activeTab === 'upload'"
          class="rounded-md border border-[var(--color-base-300)] bg-[--color-base-100] p-2"
        >
          <img :src="modelValue" alt="" class="max-h-32 w-full object-contain" />
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
import FileDropZone from "@/components/ui/FileDropZone.vue";

export type SignatureFormat =
  | ""
  | "drawn"
  | "typed"
  | "drawn_or_typed"
  | "drawn_or_upload"
  | "upload";

interface Props {
  modelValue?: string;
  mode?: "signature" | "initials";
  /** Format from template: '', drawn, typed, drawn_or_typed, drawn_or_upload, upload */
  format?: SignatureFormat | string;
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
  format: "",
  placeholder: "",
  required: false,
  disabled: false,
  error: ""
});

const emit = defineEmits<Emits>();
const { t } = useI18n();

const canvasEl = ref<HTMLCanvasElement | null>(null);
const uploadFileName = ref("");
const isDrawing = ref(false);
const last = ref({ x: 0, y: 0 });
const typedText = ref("");
const typedPreviewUrl = ref("");

// Which panels to show (DocuSeal-style: enable/disable drawn, typed, upload)
const showDrawn = computed(() => {
  const f = (props.format || "").toLowerCase();
  if (!f || f === "any") return true;
  return f === "drawn" || f === "drawn_or_typed" || f === "drawn_or_upload";
});
const showTyped = computed(() => {
  const f = (props.format || "").toLowerCase();
  if (!f || f === "any") return true;
  return f === "typed" || f === "drawn_or_typed";
});
const showUpload = computed(() => {
  const f = (props.format || "").toLowerCase();
  if (!f || f === "any") return true;
  return f === "upload" || f === "drawn_or_upload";
});

const tabs = computed(() => {
  const list: { id: "drawn" | "typed" | "upload"; label: string }[] = [];
  if (showDrawn.value) list.push({ id: "drawn", label: t("signing.signatureDraw") });
  if (showTyped.value) list.push({ id: "typed", label: t("signing.signatureType") });
  if (showUpload.value) list.push({ id: "upload", label: t("signing.signatureUpload") });
  return list;
});

const activeTab = ref<"drawn" | "typed" | "upload">("drawn");

// Set initial activeTab to first available when format changes
watch(
  () => [showDrawn.value, showTyped.value, showUpload.value],
  () => {
    if (showDrawn.value && (activeTab.value === "drawn" || !tabs.value.find((x) => x.id === activeTab.value)))
      activeTab.value = "drawn";
    else if (showTyped.value && activeTab.value !== "drawn" && activeTab.value !== "upload") activeTab.value = "typed";
    else if (showUpload.value) activeTab.value = "upload";
    if (tabs.value.length && !tabs.value.some((x) => x.id === activeTab.value))
      activeTab.value = tabs.value[0].id;
  },
  { immediate: true }
);

const hasValue = computed(() => !!props.modelValue && props.modelValue.trim() !== "");

/** Cursive font for typed signature (system fallback). */
const typedFontFamily = "cursive";

const TYPED_CANVAS_WIDTH = 400;
const TYPED_CANVAS_HEIGHT = 120;

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
  ctx.strokeStyle = "#111827";
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
  const dataUrl = c.toDataURL("image/png");
  emit("update:modelValue", dataUrl);
  await nextTick();
  emit("blur");
}

/** Render typed text to image (data URL) using offscreen canvas. */
function typedTextToDataUrl(text: string): string {
  if (!text || !text.trim()) return "";
  const canvas = document.createElement("canvas");
  canvas.width = TYPED_CANVAS_WIDTH;
  canvas.height = TYPED_CANVAS_HEIGHT;
  const ctx = canvas.getContext("2d");
  if (!ctx) return "";
  ctx.fillStyle = "transparent";
  ctx.fillRect(0, 0, canvas.width, canvas.height);
  ctx.fillStyle = "#111827";
  ctx.font = `italic 48px ${typedFontFamily}`;
  ctx.textAlign = "left";
  ctx.textBaseline = "middle";
  ctx.fillText(text.trim(), 20, TYPED_CANVAS_HEIGHT / 2);
  return canvas.toDataURL("image/png");
}

function emitTypedAsImage(): void {
  const text = typedText.value;
  const dataUrl = typedTextToDataUrl(text);
  typedPreviewUrl.value = dataUrl;
  emit("update:modelValue", dataUrl);
  nextTick().then(() => emit("blur"));
}

function onUploadChange(file: File): void {
  uploadFileName.value = file.name;
  const reader = new FileReader();
  reader.onload = (ev) => {
    const result = ev.target?.result as string;
    if (result) {
      emit("update:modelValue", result);
      nextTick().then(() => emit("blur"));
    }
  };
  reader.readAsDataURL(file);
}

function clearUpload(): void {
  uploadFileName.value = "";
  emit("update:modelValue", "");
  nextTick().then(() => emit("blur"));
}

function clear(): void {
  if (props.disabled) return;
  clearCanvas();
  typedText.value = "";
  typedPreviewUrl.value = "";
  uploadFileName.value = "";
  emit("update:modelValue", "");
  nextTick().then(() => emit("blur"));
}

onMounted(() => {
  setupCanvasSize();
  drawFromModelValue(props.modelValue || "");

  if (canvasEl.value && typeof ResizeObserver !== "undefined") {
    resizeObserver = new ResizeObserver(() => {
      setupCanvasSize();
      if (activeTab.value === "drawn") drawFromModelValue(props.modelValue || "");
    });
    resizeObserver.observe(canvasEl.value);
  }
});

let resizeObserver: ResizeObserver | null = null;

onBeforeUnmount(() => {
  resizeObserver?.disconnect();
  resizeObserver = null;
});

watch(
  () => props.modelValue,
  (v) => {
    setupCanvasSize();
    if (activeTab.value === "drawn") drawFromModelValue(v || "");
    if (activeTab.value === "upload") typedPreviewUrl.value = "";
    if (!v) {
      typedText.value = "";
      typedPreviewUrl.value = "";
      uploadFileName.value = "";
    }
  }
);
</script>

<style scoped>
.signature-input canvas {
  touch-action: none;
}
</style>
