<script lang="ts">
  import { getContext } from "svelte";
  import Field from "@/components/field/Field.svelte";
  import Submitter from "@/components/field/Submitter.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { fieldIcons as fieldIconsConst, fieldNames as fieldNamesConst } from "@/components/field/constants";
  import { v4 } from "uuid";

  interface Props {
    fields: Record<string, unknown>[];
    editable?: boolean;
    defaultFields?: string[];
    onlyDefinedFields?: boolean;
    defaultSubmitters?: Record<string, unknown>[];
    submitters: Record<string, unknown>[];
    selectedSubmitter?: Record<string, unknown> | null;
    selectedField?: Record<string, unknown> | null;
    onSetDraw?: (payload: Record<string, unknown>) => void;
    onSetDrag?: (field: Record<string, unknown>) => void;
    onDragEnd?: () => void;
    onScrollToArea?: (area: unknown) => void;
    onChangeSubmitter?: (submitter: Record<string, unknown> | undefined) => void;
  }

  let {
    fields,
    editable = true,
    defaultFields = [],
    onlyDefinedFields = true,
    defaultSubmitters = [],
    submitters,
    selectedSubmitter = null,
    selectedField = null,
    onSetDraw,
    onSetDrag,
    onDragEnd,
    onScrollToArea,
    onChangeSubmitter
  }: Props = $props();

  const save = getContext<() => void>("save") ?? (() => {});

  let dragField = $state<Record<string, unknown> | undefined>(undefined);
  let fieldsRef = $state<HTMLElement | null>(null);

  const fieldNames = fieldNamesConst;
  const fieldIcons = fieldIconsConst;

  const submitterFields = $derived(
    selectedSubmitter ? fields.filter((f) => f.submitter_id === selectedSubmitter.id) : []
  );

  function onDragstart(field: Record<string, unknown>): void {
    onSetDrag?.(field);
  }

  function onFieldDragover(e: DragEvent): void {
    const target = e.target as HTMLElement;
    const targetField = target.closest("[data-uuid]");
    const dragFieldElement = dragField as unknown as Element | undefined;

    if (dragFieldElement && targetField && targetField !== dragFieldElement) {
      const container = e.currentTarget as HTMLElement;
      const fieldElements = Array.from(container.children);
      const currentIndex = fieldElements.indexOf(dragFieldElement);
      const targetIndex = fieldElements.indexOf(targetField);

      if (currentIndex < targetIndex) {
        targetField.after(dragFieldElement);
      } else {
        targetField.before(dragFieldElement);
      }
    }
  }

  function reorderFields(): void {
    if (!fieldsRef) {
      return;
    }
    Array.from(fieldsRef.children).forEach((el, index) => {
      const htmlEl = el as HTMLElement;
      if (htmlEl.dataset.id !== fields[index].id) {
        const field = fields.find((f) => f.id === htmlEl.dataset.id);
        if (field) {
          fields.splice(fields.indexOf(field), 1);
          fields.splice(index, 0, field);
        }
      }
    });
    save();
  }

  function removeSubmitter(submitter: Record<string, unknown>): void {
    [...fields].forEach((field) => {
      if (field.submitter_id === submitter.id) {
        removeField(field);
      }
    });
    submitters.splice(submitters.indexOf(submitter), 1);

    if (selectedSubmitter === submitter) {
      onChangeSubmitter?.(submitters[0] as Record<string, unknown> | undefined);
    }
    save();
  }

  function removeField(field: Record<string, unknown>): void {
    fields.splice(fields.indexOf(field), 1);
    save();
  }

  function addField(type: string): void {
    const field: Record<string, unknown> = {
      name: "",
      id: v4(),
      required: type !== "checkbox",
      areas: [],
      submitter_id: selectedSubmitter?.id || (submitters.length > 0 ? submitters[0].id : ""),
      type
    };

    if (["select", "multiple", "radio"].includes(type)) {
      field.options = [{ value: "", id: v4() }];
    }

    if (type === "stamp") {
      field.readonly = false;
    }

    if (type === "date") {
      field.preferences = {
        format: Intl.DateTimeFormat().resolvedOptions().locale.endsWith("-US") ? "MM/DD/YYYY" : "DD/MM/YYYY"
      };
    }

    fields.push(field);

    if (!["payment", "file"].includes(type)) {
      onSetDraw?.({ field });
    }
    save();
  }
</script>

<div class="sticky top-0 z-10">
  {#if selectedSubmitter}
    <Submitter
      value={selectedSubmitter.id as string}
      class="w-full rounded-lg bg-[#faf7f5]"
      {submitters}
      editable={editable && !defaultSubmitters.length}
      onNewSubmitter={save}
      onRemove={removeSubmitter}
      onNameChange={save}
      onValueChange={(id) => {
        onChangeSubmitter?.(submitters.find((s) => s.id === id));
      }}
    />
  {/if}
</div>

<div
  bind:this={fieldsRef}
  class="mt-2 mb-1"
  role="presentation"
  ondragover={(e) => {
    e.preventDefault();
    onFieldDragover(e);
  }}
  ondrop={reorderFields}
>
  {#each submitterFields as field (field.id)}
    <div
      data-uuid={field.id}
      role="presentation"
      draggable={editable}
      ondragstart={() => {
        dragField = field;
      }}
      ondragend={() => {
        dragField = undefined;
      }}
    >
      <Field
        {field}
        editable={editable && (!dragField || dragField !== field)}
        defaultField={((defaultFields as unknown[]).find((f) =>
          typeof f === "string" ? f === field.name : (f as Record<string, unknown>).name === field.name
        ) as Record<string, unknown> | undefined) ?? null}
        isSelected={!!(selectedField && selectedField.id === field.id)}
        onRemove={removeField}
        onScrollTo={(area) => onScrollToArea?.(area)}
        onSetDraw={(payload) => onSetDraw?.(payload)}
      />
    </div>
  {/each}
</div>

{#if editable && !onlyDefinedFields}
  <div class="grid grid-cols-3 gap-1 pb-2">
    {#each Object.entries(fieldIcons) as [type, icon] (type)}
      <button
        draggable={true}
        class="group relative flex w-full items-center justify-center rounded border border-dashed border-[#e7e2df] hover:border-[#291334]/20"
        ondragstart={() => onDragstart({ type })}
        ondragend={() => onDragEnd?.()}
        onclick={() => addField(type)}
      >
        <div class="items-console absolute left-0 flex h-full cursor-grab transition-all group-hover:bg-[#efeae6]/50">
          <SvgIcon name="drag" width="18" height="18" class="my-auto" />
        </div>
        <div class="flex flex-col items-center px-2 py-2">
          <SvgIcon name={icon} width="20" height="20" />
          <span class="mt-1 text-xs">{fieldNames[type]}</span>
        </div>
      </button>
    {/each}
  </div>
{/if}

{#if fields.length < 4 && editable}
  <div class="rounded border border-[#efeae6] p-2 text-xs">
    <ul class="ml-2 list-outside list-disc pl-2">
      <li>Draw a text field on the page with a mouse</li>
      <li>Drag &amp; drop any other field type on the page</li>
      <li>Click on the field type above to start drawing it</li>
    </ul>
  </div>
{/if}
