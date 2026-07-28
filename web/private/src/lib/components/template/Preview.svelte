<script lang="ts">
  import Contenteditable from "@/components/field/Contenteditable.svelte";
  import { fileUrl } from "@/services/api-base";

  interface Props {
    item: Record<string, unknown> & { name: string };
    document: {
      preview_images: Array<{ filename: string; metadata: { width: number; height: number }; url?: string }>;
      url?: string;
      id?: string;
    };
    editable?: boolean;
    withArrows?: boolean;
    onScrollTo?: (item: Props["item"]) => void;
    onChange?: () => void;
    onRemove?: (item: Props["item"]) => void;
    onUp?: (item: Props["item"]) => void;
    onDown?: (item: Props["item"]) => void;
  }

  let {
    item,
    document,
    editable = true,
    withArrows = true,
    onScrollTo,
    onChange,
    onRemove,
    onUp,
    onDown
  }: Props = $props();

  const previewImage = $derived(
    [...document.preview_images].sort((a, b) => parseInt(a.filename, 10) - parseInt(b.filename, 10))[0]
  );

  /** API returns preview_images without url; build base like Document.svelte does. */
  const previewBaseUrl = $derived.by(() => {
    const base = fileUrl((document.url || "/drive/pages").replace(/\/$/, ""));
    return document.id ? `${base}/${document.id}` : base;
  });

  function onUpdateName(value: string): void {
    item.name = value;
    onChange?.();
  }
</script>

<div>
  <div class="relative">
    <img
      src={`${previewBaseUrl}/p/${previewImage.filename}`}
      alt={item.name}
      width={previewImage.metadata.width}
      height={previewImage.metadata.height}
      class="rounded border border-[#e7e2df]"
      loading="lazy"
    />
    <div
      class="group absolute top-0 right-0 bottom-0 left-0 flex cursor-pointer justify-end p-1"
      role="presentation"
      onclick={() => onScrollTo?.(item)}
    >
      {#if editable}
        <div class="flex w-full justify-between">
          <div style="width: 26px"></div>
          <div class="flex flex-col justify-between opacity-0 group-hover:opacity-100">
            <div>
              <button
                class="btn btn-xs w-full rounded border-base-200 bg-white text-[--color-base-content] transition-colors hover:border-base-content hover:bg-base-content hover:text-base-100"
                style="width: 24px; height: 24px"
                onclick={(e) => {
                  e.stopPropagation();
                  onRemove?.(item);
                }}
              >
                &times;
              </button>
            </div>
            {#if withArrows}
              <div class="flex flex-col space-y-1">
                <button
                  class="btn btn-xs w-full rounded border-base-200 bg-white text-[--color-base-content] transition-colors hover:border-base-content hover:bg-base-content hover:text-base-100"
                  style="width: 24px; height: 24px"
                  onclick={(e) => {
                    e.stopPropagation();
                    onUp?.(item);
                  }}
                >
                  &uarr;
                </button>
                <button
                  class="btn btn-xs w-full rounded border-base-200 bg-white text-[--color-base-content] transition-colors hover:border-base-content hover:bg-base-content hover:text-base-100"
                  style="width: 24px; height: 24px"
                  onclick={(e) => {
                    e.stopPropagation();
                    onDown?.(item);
                  }}
                >
                  &darr;
                </button>
              </div>
            {/if}
          </div>
        </div>
      {/if}
    </div>
  </div>
  <div class="flex pt-1.5 pb-2">
    <Contenteditable
      value={item.name}
      iconWidth={16}
      {editable}
      style="max-width: 95%"
      class="mx-auto"
      onValueChange={onUpdateName}
    />
  </div>
</div>
