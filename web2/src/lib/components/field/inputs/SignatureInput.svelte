<script lang="ts">
  import { onMount, tick } from "svelte";
  import { t } from "@/i18n/index.svelte";
  import FileDropZone from "@/components/ui/FileDropZone.svelte";

  export type SignatureFormat = "" | "drawn" | "typed" | "drawn_or_typed" | "drawn_or_upload" | "upload";

  interface Props {
    value?: string;
    mode?: "signature" | "initials" | "stamp";
    /** Format from template: '', drawn, typed, drawn_or_typed, drawn_or_upload, upload. Stamp uses upload only. */
    format?: SignatureFormat | string;
    placeholder?: string;
    disabled?: boolean;
    error?: string;
    onBlur?: () => void;
  }

  let {
    value = $bindable(""),
    mode = "signature",
    format = "",
    placeholder = "",
    disabled = false,
    error = "",
    onBlur
  }: Props = $props();

  let canvasEl = $state<HTMLCanvasElement | null>(null);
  let uploadFileName = $state("");
  let isDrawing = $state(false);
  let typedText = $state("");
  let typedPreviewUrl = $state("");
  let activeTab = $state<"drawn" | "typed" | "upload">("drawn");

  // Which panels to show (DocuSeal-style: enable/disable drawn, typed, upload)
  const showDrawn = $derived.by(() => {
    const f = (format || "").toLowerCase();
    if (!f || f === "any") {
      return true;
    }
    return f === "drawn" || f === "drawn_or_typed" || f === "drawn_or_upload";
  });

  const showTyped = $derived.by(() => {
    const f = (format || "").toLowerCase();
    if (!f || f === "any") {
      return true;
    }
    return f === "typed" || f === "drawn_or_typed";
  });

  const showUpload = $derived.by(() => {
    const f = (format || "").toLowerCase();
    if (!f || f === "any") {
      return true;
    }
    return f === "upload" || f === "drawn_or_upload";
  });

  const tabs = $derived.by(() => {
    const list: { id: "drawn" | "typed" | "upload"; label: string }[] = [];
    if (showDrawn) {
      list.push({ id: "drawn", label: t("signing.signatureDraw") });
    }
    if (showTyped) {
      list.push({ id: "typed", label: t("signing.signatureType") });
    }
    if (showUpload) {
      list.push({ id: "upload", label: t("signing.signatureUpload") });
    }
    return list;
  });

  // Set initial activeTab to first available when format changes
  $effect(() => {
    // Read all three up front so the effect tracks them despite short-circuiting below
    const drawn = showDrawn;
    const typed = showTyped;
    const upload = showUpload;
    if (drawn && (activeTab === "drawn" || !tabs.find((x) => x.id === activeTab))) {
      activeTab = "drawn";
    } else if (typed && activeTab !== "drawn" && activeTab !== "upload") {
      activeTab = "typed";
    } else if (upload) {
      activeTab = "upload";
    }
    if (tabs.length && !tabs.some((x) => x.id === activeTab)) {
      activeTab = tabs[0].id;
    }
  });

  const hasValue = $derived(!!value && value.trim() !== "");

  /** Cursive font for typed signature (system fallback). */
  const typedFontFamily = "cursive";

  const TYPED_CANVAS_WIDTH = 400;
  const TYPED_CANVAS_HEIGHT = 120;

  let resizeObserver: ResizeObserver | null = null;

  function getCtx(): CanvasRenderingContext2D | null {
    const c = canvasEl;
    if (!c) {
      return null;
    }
    return c.getContext("2d");
  }

  function setupCanvasSize(): void {
    const c = canvasEl;
    if (!c) {
      return;
    }
    const ctx = getCtx();
    if (!ctx) {
      return;
    }

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
    ctx.lineWidth = mode === "initials" ? 3 : 2;
  }

  function clearCanvas(): void {
    const c = canvasEl;
    const ctx = getCtx();
    if (!c || !ctx) {
      return;
    }
    ctx.clearRect(0, 0, c.width, c.height);
  }

  function drawFromModelValue(dataUrl: string): void {
    if (!dataUrl) {
      clearCanvas();
      return;
    }
    const ctx = getCtx();
    const c = canvasEl;
    if (!ctx || !c) {
      return;
    }

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
    const rect = canvasEl?.getBoundingClientRect();
    if (!rect) {
      return { x: 0, y: 0 };
    }
    return {
      x: e.clientX - rect.left,
      y: e.clientY - rect.top
    };
  }

  function onPointerDown(e: PointerEvent): void {
    if (disabled) {
      return;
    }
    if (!canvasEl) {
      return;
    }
    setupCanvasSize();

    const ctx = getCtx();
    if (!ctx) {
      return;
    }

    isDrawing = true;
    const p = getPoint(e);
    ctx.beginPath();
    ctx.moveTo(p.x, p.y);

    try {
      canvasEl.setPointerCapture(e.pointerId);
    } catch {
      // ignore
    }
  }

  function onPointerMove(e: PointerEvent): void {
    if (!isDrawing || disabled) {
      return;
    }
    const ctx = getCtx();
    if (!ctx) {
      return;
    }

    const p = getPoint(e);
    ctx.lineTo(p.x, p.y);
    ctx.stroke();
  }

  async function onPointerUp(): Promise<void> {
    if (!isDrawing) {
      return;
    }
    isDrawing = false;

    const c = canvasEl;
    if (!c) {
      return;
    }
    const dataUrl = c.toDataURL("image/png");
    value = dataUrl;
    await tick();
    onBlur?.();
  }

  /** Render typed text to image (data URL) using offscreen canvas. */
  function typedTextToDataUrl(text: string): string {
    if (!text || !text.trim()) {
      return "";
    }
    const canvas = document.createElement("canvas");
    canvas.width = TYPED_CANVAS_WIDTH;
    canvas.height = TYPED_CANVAS_HEIGHT;
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      return "";
    }
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
    const text = typedText;
    const dataUrl = typedTextToDataUrl(text);
    typedPreviewUrl = dataUrl;
    value = dataUrl;
    tick().then(() => onBlur?.());
  }

  function onUploadChange(file: File): void {
    uploadFileName = file.name;
    const reader = new FileReader();
    reader.onload = (ev) => {
      const result = ev.target?.result as string;
      if (result) {
        value = result;
        tick().then(() => onBlur?.());
      }
    };
    reader.readAsDataURL(file);
  }

  function clearUpload(): void {
    uploadFileName = "";
    value = "";
    tick().then(() => onBlur?.());
  }

  function clear(): void {
    if (disabled) {
      return;
    }
    clearCanvas();
    typedText = "";
    typedPreviewUrl = "";
    uploadFileName = "";
    value = "";
    tick().then(() => onBlur?.());
  }

  onMount(() => {
    setupCanvasSize();
    drawFromModelValue(value || "");

    if (canvasEl && typeof ResizeObserver !== "undefined") {
      resizeObserver = new ResizeObserver(() => {
        setupCanvasSize();
        if (activeTab === "drawn") {
          drawFromModelValue(value || "");
        }
      });
      resizeObserver.observe(canvasEl);
    }

    return () => {
      resizeObserver?.disconnect();
      resizeObserver = null;
    };
  });

  $effect(() => {
    const v = value;
    setupCanvasSize();
    if (activeTab === "drawn") {
      drawFromModelValue(v || "");
    }
    if (activeTab === "upload") {
      typedPreviewUrl = "";
    }
    if (!v) {
      typedText = "";
      typedPreviewUrl = "";
      uploadFileName = "";
    }
  });
</script>

<div class="signature-input">
  <div
    class="rounded-lg border border-[var(--color-base-300)] bg-white p-3"
    class:pointer-events-none={disabled}
    class:opacity-60={disabled}
  >
    <div class="mb-2 flex items-center justify-between gap-3">
      <div class="text-sm font-medium text-[--color-base-content]">
        {mode === "initials" ? t("fields.initials") : mode === "stamp" ? t("fields.stamp") : t("fields.signature")}
      </div>
      <button type="button" class="btn btn-ghost btn-xs" {disabled} onclick={clear}>
        {t("common.clear")}
      </button>
    </div>

    <!-- Tabs when multiple formats allowed (DocuSeal-style) -->
    {#if tabs.length > 1}
      <div class="tabs tabs-boxed mb-3 flex gap-1 rounded-lg bg-[var(--color-base-200)] p-1" role="tablist">
        {#each tabs as tab (tab.id)}
          <button
            type="button"
            role="tab"
            aria-selected={activeTab === tab.id}
            class="tab tab-sm flex-1"
            class:tab-active={activeTab === tab.id}
            {disabled}
            onclick={() => (activeTab = tab.id)}
          >
            {tab.label}
          </button>
        {/each}
      </div>
    {/if}

    <!-- Drawn panel -->
    {#if showDrawn}
      <div class="relative" style:display={activeTab === "drawn" ? null : "none"}>
        <canvas
          bind:this={canvasEl}
          class="w-full rounded-md border border-dashed border-[var(--color-base-300)] bg-[--color-base-100]"
          style="height: 160px"
          onpointerdown={onPointerDown}
          onpointermove={onPointerMove}
          onpointerup={onPointerUp}
          onpointercancel={onPointerUp}
          onpointerleave={onPointerUp}
        ></canvas>
        {#if !hasValue && activeTab === "drawn"}
          <div
            class="pointer-events-none absolute inset-0 flex items-center justify-center text-sm text-[--color-base-content]/60"
          >
            {placeholder || (mode === "initials" ? t("signing.drawInitials") : t("signing.drawSignature"))}
          </div>
        {/if}
      </div>
    {/if}

    <!-- Typed panel -->
    {#if showTyped}
      <div class="space-y-2" style:display={activeTab === "typed" ? null : "none"}>
        <input
          bind:value={typedText}
          type="text"
          class="input input-bordered w-full"
          placeholder={mode === "initials" ? t("signing.typeInitials") : t("signing.typeSignature")}
          {disabled}
          style:font-family={typedFontFamily}
          oninput={emitTypedAsImage}
          onblur={() => onBlur?.()}
        />
        {#if typedText && typedPreviewUrl}
          <div
            class="rounded-md border border-[var(--color-base-300)] bg-[--color-base-100] p-2"
            style="min-height: 60px"
          >
            <img src={typedPreviewUrl} alt="" class="max-h-20 w-full object-contain" style="font-family: cursive" />
          </div>
        {/if}
      </div>
    {/if}

    <!-- Upload panel -->
    {#if showUpload}
      <div class="space-y-2" style:display={activeTab === "upload" ? null : "none"}>
        <FileDropZone
          accept="image/*"
          {disabled}
          selectedLabel={uploadFileName || (hasValue ? "Image" : "")}
          onChange={onUploadChange}
          onClear={clearUpload}
        />
        {#if hasValue && activeTab === "upload"}
          <div class="rounded-md border border-[var(--color-base-300)] bg-[--color-base-100] p-2">
            <img src={value} alt="" class="max-h-32 w-full object-contain" />
          </div>
        {/if}
      </div>
    {/if}

    {#if error}
      <div class="mt-2 text-sm text-error">
        {error}
      </div>
    {/if}
  </div>
</div>

<style>
  .signature-input canvas {
    touch-action: none;
  }
</style>
