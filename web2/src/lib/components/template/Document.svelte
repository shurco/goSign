<script lang="ts">
  import Page from "@/components/template/Page.svelte";
  import type { Document as DocumentModel, PreviewImages } from "@/models/index";
  import type { Area, Field } from "@/models/template";

  type PageImage = PreviewImages & { url?: string };

  interface AreaIndexItem {
    area: Area;
    field: Field;
  }

  interface Props {
    document: DocumentModel;
    areasIndex?: Record<number, AreaIndexItem[]>;
    defaultFields?: Field[];
    allowDraw?: boolean;
    editable?: boolean;
    drawField?: Field | null;
    isDrag?: boolean;
    onDropField?: (event: Record<string, unknown>) => void;
    onRemoveArea?: (area: Area) => void;
    onDraw?: (event: Record<string, unknown>) => void;
    onSelectSubmitter?: (submitterId: string) => void;
  }

  let {
    document,
    areasIndex = {},
    defaultFields = [],
    allowDraw = true,
    editable = true,
    drawField = null,
    isDrag = false,
    onDropField,
    onRemoveArea,
    onDraw,
    onSelectSubmitter
  }: Props = $props();

  const numberOfPages = $derived.by(() => {
    if (!document) {
      return 0;
    }
    return document.metadata?.pdf?.number_of_pages || document.preview_images.length;
  });

  const previewImagesIndex = $derived.by(() => {
    if (!document) {
      return {} as Record<number, PageImage>;
    }
    return document.preview_images.reduce<Record<number, PageImage>>((acc, e) => {
      const entry = { ...e, url: `${document.url}/${document.id}` };
      acc[parseInt(e.filename, 10)] = entry;
      return acc;
    }, {});
  });

  const sortedPreviewImages = $derived.by(() => {
    if (!document || !document.preview_images.length) {
      return [] as PageImage[];
    }
    const lazyloadMetadata = document.preview_images[document.preview_images.length - 1].metadata;
    return [...Array(numberOfPages).keys()].map((i) => {
      return (
        previewImagesIndex[i] || {
          metadata: lazyloadMetadata,
          id: Math.random().toString(),
          url: `${document.url}/${document.id}`,
          filename: document.preview_images[i].filename
        }
      );
    });
  });

  // Writable derived: reset when the page count changes; bind:this fills the slots in between
  let pageRefEls: Array<ReturnType<typeof Page> | undefined> = $derived(new Array(sortedPreviewImages.length));

  export function scrollToArea(area: Area): void {
    const pageRef = pageRefEls[area.page];
    if (!pageRef || !pageRef.areaRefs) {
      return;
    }

    const areaRef = pageRef.areaRefs.find((e) => e?.area === area);
    if (!areaRef || !areaRef.rootRef) {
      return;
    }

    areaRef.rootRef.scrollIntoView({ behavior: "smooth", block: "center" });
  }

  export { document, pageRefEls as pageRefs };

  function handleDropField(event: Record<string, unknown>): void {
    if (!document) {
      return;
    }
    onDropField?.({ ...event, attachment_id: document.id });
  }

  function handleRemoveArea(area: Area): void {
    onRemoveArea?.(area);
  }

  function handleDraw(event: Area): void {
    if (!document) {
      return;
    }
    // Preserve page number from the event and add attachment_id
    onDraw?.({
      ...event,
      attachment_id: document.id,
      page: event.page !== undefined ? event.page : 0 // Ensure page is preserved
    });
  }
</script>

<div>
  {#each sortedPreviewImages as image, index (image.id as string)}
    <Page
      bind:this={pageRefEls[index]}
      number={index}
      {editable}
      areas={areasIndex[index]}
      {allowDraw}
      {isDrag}
      {defaultFields}
      {drawField}
      {image}
      onDropField={handleDropField}
      onRemoveArea={handleRemoveArea}
      onDraw={handleDraw}
      {onSelectSubmitter}
    />
  {/each}
</div>
