<script lang="ts">
  import { tick } from "svelte";
  import FieldArea from "@/components/template/Area.svelte";
  import type { PreviewImages } from "@/models/index";
  import type { Area, Field } from "@/models/template";

  type PageImage = PreviewImages & { url?: string };

  interface AreaItem {
    area: Area;
    field: Field;
  }

  interface Props {
    image: PageImage;
    areas?: AreaItem[];
    defaultFields?: Field[];
    allowDraw?: boolean;
    drawField?: Field | null;
    editable?: boolean;
    isDrag?: boolean;
    number: number;
    onDraw?: (area: Area) => void;
    onDropField?: (event: Record<string, unknown>) => void;
    onRemoveArea?: (area: Area) => void;
    onSelectSubmitter?: (submitterId: string) => void;
  }

  let {
    image,
    areas = [],
    defaultFields = [],
    allowDraw = true,
    drawField = null,
    editable = true,
    isDrag = false,
    number,
    onDraw,
    onDropField,
    onRemoveArea,
    onSelectSubmitter
  }: Props = $props();

  // Writable derived: reset when the area count changes; bind:this fills the slots in between
  let areaRefEls: Array<ReturnType<typeof FieldArea> | undefined> = $derived(new Array(areas.length));
  let isMove = $state(false);
  let resizeDirection = $state<string | null>(null);
  let newArea = $state<Area | null>(null);
  let mask = $state<HTMLDivElement | null>(null);
  let imageEl = $state<HTMLImageElement | null>(null);

  const width = $derived(image.metadata.width);
  const height = $derived(image.metadata.height);

  export { areaRefEls as areaRefs, mask };

  function onImageLoad(e: Event): void {
    const target = e.target as HTMLImageElement;
    target.setAttribute("width", target.naturalWidth.toString());
    target.setAttribute("height", target.naturalHeight.toString());
  }

  function onDrop(e: DragEvent): void {
    e.preventDefault();
    if (!mask) {
      return;
    }

    const rect = mask.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    onDropField?.({
      x,
      y,
      maskW: mask.clientWidth,
      maskH: mask.clientHeight,
      page: number
    });
  }

  function onStartDraw(e: PointerEvent): void {
    if (!allowDraw || !drawField || !editable || !mask) {
      return;
    }

    tick().then(() => {
      if (!mask) {
        return;
      }

      const initialX = e.layerX / mask.clientWidth;
      const initialY = e.layerY / mask.clientHeight;

      newArea = {
        initialX,
        initialY,
        x: initialX,
        y: initialY,
        w: 0,
        h: 0,
        page: number,
        attachment_id: ""
      };
    });
  }

  function onPointermove(e: PointerEvent): void {
    if (!newArea || !mask) {
      return;
    }

    const dx = e.layerX / mask.clientWidth - (newArea.initialX || 0);
    const dy = e.layerY / mask.clientHeight - (newArea.initialY || 0);

    newArea.x = dx > 0 ? newArea.initialX || 0 : e.layerX / mask.clientWidth;
    newArea.y = dy > 0 ? newArea.initialY || 0 : e.layerY / mask.clientHeight;

    if (drawField?.type === "cells") {
      newArea.cell_w = newArea.h * (mask.clientHeight / mask.clientWidth);
    }

    newArea.w = Math.abs(dx);
    newArea.h = Math.abs(dy);
  }

  function onPointerup(): void {
    if (newArea) {
      const area: Area = {
        x: newArea.x,
        y: newArea.y,
        w: newArea.w,
        h: newArea.h,
        page: number, // This is 0-based index (0 = first page, 1 = second page, etc.)
        attachment_id: ""
      };

      if (newArea.cell_w) {
        area.cell_w = newArea.cell_w;
        // Persist cell count for signing page (same formula as Area.vue)
        const cw = newArea.cell_w;
        const aw = newArea.w;
        if (cw > 0 && aw > 0) {
          let currentWidth = 0;
          let count = 0;
          while (currentWidth + (cw + cw / 4) < aw) {
            currentWidth += cw;
            count++;
          }
          area.cell_count = Math.max(count, 1);
        }
      }

      // Ensure page number is explicitly set
      onDraw?.({ ...area, page: number });
    }

    newArea = null;
  }
</script>

<div class="relative cursor-crosshair select-none" style={drawField ? "touch-action: none" : ""}>
  <img
    bind:this={imageEl}
    loading="lazy"
    src={`${image.url}/${image.filename}`}
    alt="Page {number + 1}"
    {width}
    {height}
    class="mb-4 rounded border border-[#e7e2df]"
    onload={onImageLoad}
  />
  <div class="absolute top-0 right-0 bottom-0 left-0" role="presentation" onpointerdown={onStartDraw}>
    {#each areas as item, i (i)}
      <FieldArea
        bind:this={areaRefEls[i]}
        area={item.area}
        field={item.field}
        {editable}
        defaultField={defaultFields.find((f) => f.name === item.field.name)}
        onStartResize={(direction) => {
          resizeDirection = direction;
        }}
        onStopResize={() => {
          resizeDirection = null;
        }}
        onStartDrag={() => {
          isMove = true;
        }}
        onStopDrag={() => {
          isMove = false;
        }}
        onRemove={() => onRemoveArea?.(item.area)}
        {onSelectSubmitter}
      />
    {/each}
    {#if newArea && drawField}
      <FieldArea isDraw={true} field={drawField} area={newArea} />
    {/if}
  </div>
  <div
    style:display={resizeDirection || isMove || isDrag || newArea || drawField ? null : "none"}
    id="mask"
    bind:this={mask}
    role="presentation"
    class="absolute top-0 right-0 bottom-0 left-0 z-10 {isDrag || isMove ? 'cursor-grab' : ''} {drawField ||
    resizeDirection === 'nwse'
      ? 'cursor-nwse-resize'
      : ''} {resizeDirection === 'ew' ? 'cursor-ew-resize' : ''}"
    onpointermove={onPointermove}
    onpointerdown={onStartDraw}
    ondragover={(e) => e.preventDefault()}
    ondrop={onDrop}
    onpointerup={onPointerup}
  ></div>
</div>
