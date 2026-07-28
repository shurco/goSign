<script lang="ts">
  import { onMount, tick, setContext } from "svelte";
  import { page } from "$app/state";
  import Document from "@/components/template/Document.svelte";
  import Fields from "@/components/field/List.svelte";
  import type { Template } from "@/models";
  import { apiGet, apiUrl } from "@/services/api";
  import { fetchWithAuth } from "@/utils/auth";
  import { v4 } from "uuid";

  // This page reuses legacy builder logic which is not strictly typed yet.
  // Keep `template` loosely typed to avoid TS friction in the editor UI code.
  let template = $state<any>();
  let loading = $state(false);
  let error = $state<string | null>(null);
  let undoStack = $state<any[]>([]);
  let redoStack = $state<any[]>([]);
  let lastRedoData = $state<any>();

  let drawField = $state<any>();
  let drawOption = $state<any>();
  let selectedSubmitter = $state<any>();
  let selectedAreaRef = $state<any>();
  let editable = $state(false); // or whatever the initial value should be
  let dragField = $state<any>();
  let autosave = $state(false); // or whatever the initial value should be

  let onSave = $state<((template: any) => void) | undefined>();

  let defaultSubmitters = $state<any[]>([]);
  let defaultFields = $state<any[]>(["text", "signature", "initials", "date"]);
  let onlyDefinedFields = $state(false);

  const fetchOptions = { headers: {} };

  setContext("template", {
    get value() {
      return template;
    },
    set value(v: any) {
      template = v;
    }
  });
  setContext("save", save);
  setContext("baseFetch", baseFetch);
  setContext("selectedAreaRef", {
    get value() {
      return selectedAreaRef;
    },
    set value(v: any) {
      selectedAreaRef = v;
    }
  });

  const templateId = $derived.by(() => {
    // Support both:
    // - /templates/:id/view
    // - /templates/:id/view?template_id=... (or ?id=...)
    const fromParams = (page.params?.id as string | undefined) || "";
    const fromQuery = page.url.searchParams.get("template_id") || page.url.searchParams.get("id") || "";
    return (fromParams || fromQuery || "").trim();
  });

  const missingTemplateId = $derived(templateId === "");

  const selectedField = $derived.by(() => {
    if (!template || !template.fields) {
      return null;
    }
    return template.fields.find((f: any) => f.areas?.includes(selectedAreaRef));
  });

  const sortedDocuments = $derived.by(() => {
    if (!template || !template.schema || !template.documents) {
      return [];
    }
    return template.schema
      .map((item: any) => {
        return template.documents.find((doc: any) => doc.id === item.attachment_id);
      })
      .filter(Boolean);
  });

  const fieldAreasIndex = $derived.by(() => {
    if (!template || !template.fields) {
      return {};
    }
    const areas: any = {};
    template.fields.forEach((f: any) => {
      (f.areas || []).forEach((a: any) => {
        areas[a.attachment_id] ||= {};
        const acc = (areas[a.attachment_id][a.page] ||= []);
        acc.push({ area: a, field: f });
      });
    });
    return areas;
  });

  // Writable derived: reset when the document count changes; bind:this fills the slots in between
  let documentRefEls: Array<ReturnType<typeof Document> | undefined> = $derived(new Array(sortedDocuments.length));

  function getPageMask(documentRef: ReturnType<typeof Document> | undefined, pageIndex: number): HTMLDivElement | null {
    if (!documentRef) {
      return null;
    }
    // Vue equivalent: documentRef.pageRefs[area.page].$refs.mask
    return (documentRef.pageRefs[pageIndex] as any)?.mask ?? null;
  }

  function undo(): void {
    if (undoStack.length > 1) {
      undoStack.pop();
      const stringData = undoStack[undoStack.length - 1];
      const currentStringData = JSON.stringify(template);

      if (stringData && stringData !== currentStringData) {
        if (!template) {
          return;
        }
        redoStack.push(currentStringData);
        Object.assign(template, JSON.parse(stringData));
        save();
      }
    }
  }

  function redo(): void {
    const stringData = redoStack.pop();
    lastRedoData = stringData;
    const currentStringData: any = JSON.stringify(template);

    if (stringData && stringData !== currentStringData) {
      if (!template) {
        return;
      }
      if (undoStack[undoStack.length - 1] !== currentStringData) {
        undoStack.push(currentStringData);
      }
      Object.assign(template, JSON.parse(stringData));
      save();
    }
  }

  function clearDrawField(): void {
    if (drawField && !drawOption && drawField.areas.length === 0) {
      if (!template) {
        return;
      }
      const fieldIndex = template.fields.indexOf(drawField);

      if (fieldIndex !== -1) {
        template.fields.splice(fieldIndex, 1);
      }
    }
    drawField = null;
    drawOption = null;
  }

  function onKeyUp(e: KeyboardEvent): void {
    if (e.code === "Escape") {
      clearDrawField();
      selectedAreaRef = null;
    }
    if (
      editable &&
      ["Backspace", "Delete"].includes(e.key) &&
      selectedAreaRef &&
      document.activeElement === document.body
    ) {
      removeArea(selectedAreaRef);
      selectedAreaRef = null;
    }
  }

  function onKeyDown(event: KeyboardEvent): void {
    if ((event.metaKey && event.shiftKey && event.key === "z") || (event.ctrlKey && event.key === "Z")) {
      event.stopImmediatePropagation();
      event.preventDefault();
      redo();
    } else if ((event.ctrlKey || event.metaKey) && event.key === "z") {
      event.stopImmediatePropagation();
      event.preventDefault();
      undo();
    }
  }

  function removeArea(area: any): void {
    const field = template.fields.find((f: any) => f.areas?.includes(area));
    field.areas.splice(field.areas.indexOf(area), 1);

    if (!field.areas.length) {
      template.fields.splice(template.fields.indexOf(field), 1);
    }
    save();
  }

  function handleSelectSubmitter(submitterId: string): void {
    const submitter = template.submitters.find((s: any) => s.id === submitterId);
    if (submitter) {
      selectedSubmitter = submitter;
    }
  }

  function pushUndo(): void {
    const stringData: any = JSON.stringify(template);

    if (undoStack[undoStack.length - 1] !== stringData) {
      undoStack.push(stringData);
      if (lastRedoData !== stringData) {
        redoStack = [];
      }
    }
  }

  function onDraw(area: any): void {
    if (drawField) {
      if (drawOption) {
        const areaWithoutOption = drawField.areas?.find((a: any) => !a.option_id);

        if (areaWithoutOption && !drawField.areas.find((a: any) => a.option_id === drawField.options[0].id)) {
          areaWithoutOption.option_id = drawField.options[0].id;
        }

        area.option_id = drawOption.id;
      }

      if (area.w === 0 || area.h === 0) {
        const previousArea = drawField.areas?.[drawField.areas.length - 1];

        if (selectedField?.type === drawField.type) {
          area.w = selectedAreaRef.w;
          area.h = selectedAreaRef.h;
        } else if (previousArea) {
          area.w = previousArea.w;
          area.h = previousArea.h;
        } else {
          const documentRef = documentRefEls.find((e) => e && e.document && e.document.id === area.attachment_id);
          if (!documentRef) {
            return;
          }
          const pageMask = getPageMask(documentRef, area.page);

          if (!pageMask) {
            return;
          }

          if (drawField.type === "checkbox" || drawOption) {
            area.w = pageMask.clientWidth / 30 / pageMask.clientWidth;
            area.h =
              (pageMask.clientWidth / 30 / pageMask.clientWidth) * (pageMask.clientWidth / pageMask.clientHeight);
          } else if (drawField.type === "image") {
            area.w = pageMask.clientWidth / 5 / pageMask.clientWidth;
            area.h = (pageMask.clientWidth / 5 / pageMask.clientWidth) * (pageMask.clientWidth / pageMask.clientHeight);
          } else if (drawField.type === "signature" || drawField.type === "stamp") {
            area.w = pageMask.clientWidth / 5 / pageMask.clientWidth;
            area.h =
              ((pageMask.clientWidth / 5 / pageMask.clientWidth) * (pageMask.clientWidth / pageMask.clientHeight)) / 2;
          } else if (drawField.type === "initials") {
            area.w = pageMask.clientWidth / 10 / pageMask.clientWidth;
            area.h = pageMask.clientWidth / 35 / pageMask.clientWidth;
          } else {
            area.w = pageMask.clientWidth / 5 / pageMask.clientWidth;
            area.h = pageMask.clientWidth / 35 / pageMask.clientWidth;
          }
        }

        area.x -= area.w / 2;
        area.y -= area.h / 2;
      }

      drawField.areas ||= [];

      const insertBeforeAreaIndex = drawField.areas.findIndex((a: any) => {
        return a.attachment_id === area.attachment_id && a.page > area.page;
      });

      if (insertBeforeAreaIndex !== -1) {
        drawField.areas.splice(insertBeforeAreaIndex, 0, area);
      } else {
        drawField.areas.push(area);
      }

      if (template.fields.indexOf(drawField) === -1) {
        template.fields.push(drawField);
      }

      drawField = null;
      drawOption = null;
      selectedAreaRef = area;
      save();
    } else {
      const documentRef = documentRefEls.find((e) => e && e.document && e.document.id === area.attachment_id);
      if (!documentRef) {
        return;
      }
      const pageMask = getPageMask(documentRef, area.page);

      if (!pageMask) {
        return;
      }

      let type = pageMask.clientWidth * area.w < 35 ? "checkbox" : "text";
      if (type === "checkbox") {
        const previousField = [...template.fields].reverse().find((f: any) => f.type === type);
        const previousArea = previousField?.areas?.[previousField.areas.length - 1];

        if (previousArea || area.w) {
          const areaW = previousArea?.w || 30 / pageMask.clientWidth;
          const areaH = previousArea?.h || 30 / pageMask.clientHeight;

          if (pageMask.clientWidth * area.w < 5) {
            area.x = area.x - areaW / 2;
            area.y = area.y - areaH / 2;
          }

          area.w = areaW;
          area.h = areaH;
        }
      }

      if (area.w) {
        const field = {
          name: "",
          id: v4(),
          required: type !== "checkbox",
          type,
          submitter_id: selectedSubmitter.id,
          areas: [area]
        };

        template.fields.push(field);
        selectedAreaRef = area;
        save();
      }
    }
  }

  function onDropfield(area: any): void {
    const field = {
      name: "",
      id: v4(),
      submitter_id: selectedSubmitter.id,
      required: dragField.type !== "checkbox",
      ...dragField
    };

    if (["select", "multiple", "radio"].includes(field.type)) {
      field.options = [{ value: "", id: v4() }];
    }

    if (field.type === "stamp") {
      field.readonly = false;
    }

    if (field.type === "date") {
      field.preferences = {
        format: Intl.DateTimeFormat().resolvedOptions().locale.endsWith("-US") ? "MM/DD/YYYY" : "DD/MM/YYYY"
      };
    }

    const fieldArea: any = {
      x: (area.x - 6) / area.maskW,
      y: area.y / area.maskH,
      page: area.page,
      attachment_id: area.attachment_id
    };

    const previousField = [...template.fields].reverse().find((f: any) => f.type === field.type);
    let baseArea: any;
    if (selectedField?.type === field.type) {
      baseArea = selectedAreaRef;
    } else if (previousField?.areas?.length) {
      baseArea = previousField.areas[previousField.areas.length - 1];
    } else {
      if (["checkbox"].includes(field.type)) {
        baseArea = {
          w: area.maskW / 30 / area.maskW,
          h: (area.maskW / 30 / area.maskW) * (area.maskW / area.maskH)
        };
      } else if (field.type === "image") {
        baseArea = {
          w: area.maskW / 5 / area.maskW,
          h: (area.maskW / 5 / area.maskW) * (area.maskW / area.maskH)
        };
      } else if (field.type === "signature" || field.type === "stamp") {
        baseArea = {
          w: area.maskW / 5 / area.maskW,
          h: ((area.maskW / 5 / area.maskW) * (area.maskW / area.maskH)) / 2
        };
      } else if (field.type === "initials") {
        baseArea = {
          w: area.maskW / 10 / area.maskW,
          h: area.maskW / 35 / area.maskW
        };
      } else {
        baseArea = {
          w: area.maskW / 5 / area.maskW,
          h: area.maskW / 35 / area.maskW
        };
      }
    }

    fieldArea.w = baseArea.w;
    fieldArea.h = baseArea.h;
    // If baseArea.h is missing, calculate a reasonable default based on field type
    if (!fieldArea.h) {
      if (["checkbox"].includes(field.type)) {
        fieldArea.h = (area.maskW / 30 / area.maskW) * (area.maskW / area.maskH);
      } else if (field.type === "image") {
        fieldArea.h = (area.maskW / 5 / area.maskW) * (area.maskW / area.maskH);
      } else if (field.type === "signature" || field.type === "stamp") {
        fieldArea.h = ((area.maskW / 5 / area.maskW) * (area.maskW / area.maskH)) / 2;
      } else if (field.type === "initials") {
        fieldArea.h = area.maskW / 35 / area.maskW;
      } else {
        fieldArea.h = area.maskW / 35 / area.maskW;
      }
    }
    fieldArea.y = fieldArea.y - fieldArea.h / 2;

    if (field.type === "cells") {
      fieldArea.cell_w = baseArea.cell_w || baseArea.w / 5;
    }

    field.areas = [fieldArea];
    selectedAreaRef = fieldArea;
    template.fields.push(field);
    save();
  }

  function scrollToArea(area: any): void {
    const documentRef = documentRefEls.find((a) => a && a.document && a.document.id === area.attachment_id);
    if (documentRef) {
      documentRef.scrollToArea(area);
      selectedAreaRef = area;
    }
  }

  function baseFetch(path: string, options: RequestInit = {}): Promise<Response> {
    // Use fetchWithAuth to ensure token is included in headers
    return fetchWithAuth(apiUrl(path), {
      ...options,
      headers: { ...fetchOptions.headers, ...options.headers }
    });
  }

  async function save({ force } = { force: false }): Promise<object> {
    if (!autosave && !force) {
      return Promise.resolve({});
    }
    if (!template) {
      return Promise.resolve({});
    }

    tick().then(() => {
      const templateBuilder = document.querySelector("template-builder") as HTMLElement;
      if (templateBuilder) {
        templateBuilder.dataset.template = JSON.stringify(template);
      }
    });

    pushUndo();

    await baseFetch(`/templates/${template.id}`, {
      method: "PUT",
      body: JSON.stringify({
        template: {
          name: template.name,
          schema: template.schema,
          submitters: template.submitters,
          fields: template.fields
        }
      }),
      headers: { "Content-Type": "application/json" }
    });
    if (onSave) {
      onSave(template);
    }

    return {};
  }

  async function init(): Promise<void> {
    if (!templateId) {
      return;
    }

    loading = true;
    error = null;
    try {
      // Load specific template by ID (same approach as Edit page)
      const res: any = await apiGet<Template>(`/templates/${templateId}`);
      if (res && res.data) {
        template = res.data as Template;

        // Ensure at least one submitter exists (some flows create empty submitters array)
        if (!template.submitters || template.submitters.length === 0) {
          template.submitters = [
            {
              id: v4(),
              name: "Signer 1",
              colorIndex: 0
            }
          ];
        }
        selectedSubmitter = template?.submitters?.[0];

        undoStack = [JSON.stringify(template)];
        redoStack = [];
      } else {
        error = "Template not found.";
      }
    } catch (e) {
      error = e instanceof Error ? e.message : "Failed to load template data.";
    } finally {
      loading = false;
    }

    await tick();
    document.addEventListener("keyup", onKeyUp);
    window.addEventListener("keydown", onKeyDown);
  }

  onMount(() => {
    init();

    return () => {
      document.removeEventListener("keyup", onKeyUp);
      window.removeEventListener("keydown", onKeyDown);
    };
  });
</script>

{#if loading}
  <div class="mx-auto max-w-2xl p-6">
    <div class="rounded-lg border border-gray-200 bg-white p-4 text-gray-900">
      <p class="font-medium">Loading…</p>
      <p class="mt-1 text-sm text-gray-600">Fetching template data.</p>
    </div>
  </div>
{:else if missingTemplateId}
  <div class="mx-auto max-w-2xl p-6">
    <div class="rounded-lg border border-amber-200 bg-amber-50 p-4 text-amber-900">
      <p class="font-medium">Nothing to show</p>
      <p class="mt-1 text-sm">
        This page requires a template id. Open
        <code class="rounded bg-amber-100 px-1 py-0.5">/templates/&lt;id&gt;/view</code>
        .
      </p>
    </div>
  </div>
{:else if error}
  <div class="mx-auto max-w-2xl p-6">
    <div class="rounded-lg border border-red-200 bg-red-50 p-4 text-red-900">
      <p class="font-medium">Unable to load template view</p>
      <p class="mt-1 text-sm">{error}</p>
    </div>
  </div>
{:else if template}
  <div class="-my-5 flex h-screen pt-5">
    <div class="w-full overflow-x-hidden overflow-y-hidden md:overflow-y-auto">
      <div class="pr-3.5 pl-0.5">
        {#each sortedDocuments as document, i (document.id)}
          <Document
            bind:this={documentRefEls[i]}
            areasIndex={fieldAreasIndex[document.id]}
            {document}
            isDrag={!!dragField}
            {defaultFields}
            allowDraw={!onlyDefinedFields}
            {drawField}
            {editable}
            {onDraw}
            onDropField={onDropfield}
            onRemoveArea={removeArea}
            onSelectSubmitter={handleSelectSubmitter}
          />
        {/each}
      </div>
    </div>

    <div class="relative hidden w-80 flex-none overflow-x-hidden overflow-y-auto pl-0.5 md:block">
      {#if drawField}
        <div class="sticky inset-0 z-20 h-full">
          <div class="space-y-4 rounded-lg bg-[--color-base-300] p-5 text-center">
            <p>Draw {drawField.name} field on the document</p>
            <div>
              <button class="base-button" onclick={clearDrawField}>Cancel</button>
              {#if !drawOption && !drawField.areas.length && !["stamp", "signature", "initials"].includes(drawField.type)}
                <button
                  type="button"
                  class="link mt-3 block w-full text-sm"
                  onclick={() => {
                    drawField = null;
                    drawOption = null;
                  }}
                >
                  Or add field without drawing
                </button>
              {/if}
            </div>
          </div>
        </div>
      {/if}

      <Fields
        fields={template.fields}
        submitters={template.submitters}
        {selectedSubmitter}
        {defaultSubmitters}
        {defaultFields}
        {onlyDefinedFields}
        {editable}
        onSetDraw={(payload) => {
          drawField = payload.field;
          drawOption = payload.option;
        }}
        onSetDrag={(field) => {
          dragField = field;
        }}
        onChangeSubmitter={(submitter) => {
          selectedSubmitter = submitter;
        }}
        onDragEnd={() => {
          dragField = null;
        }}
        onScrollToArea={scrollToArea}
      />
    </div>
  </div>
{/if}
