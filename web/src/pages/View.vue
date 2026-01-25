<template>
  <div v-if="loading" class="mx-auto max-w-2xl p-6">
    <div class="rounded-lg border border-gray-200 bg-white p-4 text-gray-900">
      <p class="font-medium">Loading…</p>
      <p class="mt-1 text-sm text-gray-600">Fetching template data.</p>
    </div>
  </div>

  <div v-else-if="missingTemplateId" class="mx-auto max-w-2xl p-6">
    <div class="rounded-lg border border-amber-200 bg-amber-50 p-4 text-amber-900">
      <p class="font-medium">Nothing to show</p>
      <p class="mt-1 text-sm">
        This page requires a template id. Open
        <code class="rounded bg-amber-100 px-1 py-0.5">/templates/&lt;id&gt;/view</code>
        .
      </p>
    </div>
  </div>

  <div v-else-if="error" class="mx-auto max-w-2xl p-6">
    <div class="rounded-lg border border-red-200 bg-red-50 p-4 text-red-900">
      <p class="font-medium">Unable to load template view</p>
      <p class="mt-1 text-sm">{{ error }}</p>
    </div>
  </div>

  <div v-else-if="template" class="-my-5 flex h-screen pt-5">
    <div class="w-full overflow-x-hidden overflow-y-hidden md:overflow-y-auto">
      <div ref="documents" class="pr-3.5 pl-0.5">
        <template v-for="document in sortedDocuments" :key="document.id">
          <Document
            :ref="setDocumentRefs"
            :areas-index="fieldAreasIndex[document.id]"
            :selected-submitter="selectedSubmitter"
            :document="document"
            :is-drag="!!dragField"
            :default-fields="defaultFields"
            :allow-draw="!onlyDefinedFields"
            :draw-field="drawField"
            :editable="editable"
            @draw="onDraw"
            @drop-field="onDropfield"
            @remove-area="removeArea"
            @select-submitter="handleSelectSubmitter"
          />
        </template>
      </div>
    </div>

    <div class="relative hidden w-80 flex-none overflow-x-hidden overflow-y-auto pl-0.5 md:block">
      <div v-if="drawField" class="sticky inset-0 z-20 h-full">
        <div class="space-y-4 rounded-lg bg-[--color-base-300] p-5 text-center">
          <p>Draw {{ drawField.name }} field on the document</p>
          <div>
            <button class="base-button" @click="clearDrawField">Cancel</button>
            <a
              v-if="
                !drawOption && !drawField.areas.length && !['stamp', 'signature', 'initials'].includes(drawField.type)
              "
              href="#"
              class="link mt-3 block text-sm"
              @click.prevent="[(drawField = null), (drawOption = null)]"
            >
              Or add field without drawing
            </a>
          </div>
        </div>
      </div>

      <Fields
        ref="fields"
        :fields="template.fields"
        :submitters="template.submitters"
        :selected-submitter="selectedSubmitter"
        :default-submitters="defaultSubmitters"
        :default-fields="defaultFields"
        :only-defined-fields="onlyDefinedFields"
        :editable="editable"
        @set-draw="[(drawField = $event.field), (drawOption = $event.option)]"
        @set-drag="dragField = $event"
        @change-submitter="selectedSubmitter = $event"
        @drag-end="dragField = null"
        @scroll-to-area="scrollToArea"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, provide, ref } from "vue";
import { useRoute } from "vue-router";
import Document from "@/components/template/Document.vue";
import Fields from "@/components/field/List.vue";
import type { Template } from "@/models";
import { apiGet } from "@/services/api";
import { fetchWithAuth } from "@/utils/auth";
import { v4 } from "uuid";

const route = useRoute();

// This page reuses legacy builder logic which is not strictly typed yet.
// Keep `template` loosely typed to avoid TS friction in the editor UI code.
const template: any = ref();
const loading = ref(false);
const error = ref<string | null>(null);
const undoStack: any = ref([]);
const redoStack: any = ref([]);
const lastRedoData: any = ref();

const documentRefs: any = ref([]);

const drawField: any = ref();
const drawOption: any = ref();
const selectedSubmitter: any = ref();
const selectedAreaRef: any = ref();
const editable = ref(false); // or whatever the initial value should be
const dragField = ref();
const autosave = ref(false); // or whatever the initial value should be

const onSave = ref();

const defaultSubmitters = ref<any[]>([]);
const defaultFields = ref<any[]>(["text", "signature", "initials", "date"]);
const onlyDefinedFields = ref(false);

const fetchOptions = { headers: {} };

provide("template", template);
provide("save", save);
provide("baseFetch", baseFetch);
provide("selectedAreaRef", selectedAreaRef); // computed(() => selectedAreaRef.value),

const templateId = computed(() => {
  // Support both:
  // - /templates/:id/view
  // - /templates/:id/view?template_id=... (or ?id=...)
  const fromParams = (route.params?.id as string | undefined) || "";
  const fromQuery =
    (route.query?.template_id as string | undefined) || (route.query?.id as string | undefined) || "";
  return (fromParams || fromQuery || "").trim();
});

const missingTemplateId = computed(() => templateId.value === "");

onMounted(async () => {
  if (!templateId.value) {
    return;
  }

  loading.value = true;
  error.value = null;
  try {
    // Load specific template by ID (same approach as Edit page)
    const res: any = await apiGet<Template>(`/api/v1/templates/${templateId.value}`);
    if (res && res.data) {
      template.value = res.data as Template;

      // Ensure at least one submitter exists (some flows create empty submitters array)
      if (!template.value.submitters || template.value.submitters.length === 0) {
        (template.value as any).submitters = [
          {
            id: v4(),
            name: "Signer 1",
            colorIndex: 0
          }
        ];
      }
      selectedSubmitter.value = (template.value as any)?.submitters?.[0];

      undoStack.value = [JSON.stringify(template.value)];
      redoStack.value = [];
    } else {
      error.value = "Template not found.";
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : "Failed to load template data.";
  } finally {
    loading.value = false;
  }

  await nextTick();
  document.addEventListener("keyup", onKeyUp);
  window.addEventListener("keydown", onKeyDown);
});

onUnmounted(() => {
  document.removeEventListener("keyup", onKeyUp);
  window.removeEventListener("keydown", onKeyDown);
});

const selectedField = computed(() => {
  if (!template.value || !(template.value as any).fields) {
    return null;
  }
  return (template.value as any).fields.find((f: any) => f.areas?.includes(selectedAreaRef.value));
});

const sortedDocuments = computed(() => {
  if (!template.value || !(template.value as any).schema || !(template.value as any).documents) {
    return [];
  }
  return (template.value as any).schema
    .map((item: any) => {
      return (template.value as any).documents.find((doc: any) => doc.id === item.attachment_id);
    })
    .filter(Boolean);
});

const fieldAreasIndex = computed(() => {
  if (!template.value || !(template.value as any).fields) {
    return {};
  }
  const areas: any = {};
  (template.value as any).fields.forEach((f: any) => {
    (f.areas || []).forEach((a: any) => {
      areas[a.attachment_id] ||= {};
      const acc = (areas[a.attachment_id][a.page] ||= []);
      acc.push({ area: a, field: f });
    });
  });
  return areas;
});

function undo(): void {
  if (undoStack.value.length > 1) {
    undoStack.value.pop();
    const stringData = undoStack.value[undoStack.value.length - 1];
    const currentStringData = JSON.stringify(template.value);

    if (stringData && stringData !== currentStringData) {
      if (!template.value) {
        return;
      }
      redoStack.value.push(currentStringData);
      Object.assign(template.value, JSON.parse(stringData));
      save();
    }
  }
}

function redo(): void {
  const stringData = redoStack.value.pop();
  lastRedoData.value = stringData;
  const currentStringData: any = JSON.stringify(template.value);

  if (stringData && stringData !== currentStringData) {
    if (!template.value) {
      return;
    }
    if (undoStack.value[undoStack.value.length - 1] !== currentStringData) {
      undoStack.value.push(currentStringData);
    }
    Object.assign(template.value, JSON.parse(stringData));
    save();
  }
}

function setDocumentRefs(el: any): void {
  if (el) {
    documentRefs.value.push(el);
  }
}

function clearDrawField(): void {
  if (drawField.value && !drawOption.value && drawField.value.areas.length === 0) {
    if (!template.value) {
      return;
    }
    const fieldIndex = template.value.fields.indexOf(drawField.value);

    if (fieldIndex !== -1) {
      template.value.fields.splice(fieldIndex, 1);
    }
  }
  drawField.value = null;
  drawOption.value = null;
}

function onKeyUp(e: KeyboardEvent): void {
  if (e.code === "Escape") {
    clearDrawField();
    selectedAreaRef.value = null;
  }
  if (
    editable.value &&
    ["Backspace", "Delete"].includes(e.key) &&
    selectedAreaRef.value &&
    document.activeElement === document.body
  ) {
    removeArea(selectedAreaRef.value);
    selectedAreaRef.value = null;
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
  const field = (template.value as any).fields.find((f: any) => f.areas?.includes(area));
  field.areas.splice(field.areas.indexOf(area), 1);

  if (!field.areas.length) {
    (template.value as any).fields.splice((template.value as any).fields.indexOf(field), 1);
  }
  save();
}

function handleSelectSubmitter(submitterId: string): void {
  const submitter = (template.value as any).submitters.find((s: any) => s.id === submitterId);
  if (submitter) {
    selectedSubmitter.value = submitter;
  }
}

function pushUndo(): void {
  const stringData: any = JSON.stringify(template.value);

  if (undoStack.value[undoStack.value.length - 1] !== stringData) {
    undoStack.value.push(stringData);
    if (lastRedoData.value !== stringData) {
      redoStack.value = [];
    }
  }
}

function onDraw(area: any): void {
  if (drawField.value) {
    if (drawOption.value) {
      const areaWithoutOption = drawField.value.areas?.find((a: any) => !a.option_id);

      if (areaWithoutOption && !drawField.value.areas.find((a: any) => a.option_id === drawField.value.options[0].id)) {
        areaWithoutOption.option_id = drawField.value.options[0].id;
      }

      area.option_id = drawOption.value.id;
    }

    if (area.w === 0 || area.h === 0) {
      const previousArea = drawField.value.areas?.[drawField.value.areas.length - 1];

      if (selectedField.value?.type === drawField.value.type) {
        area.w = selectedAreaRef.value.w;
        area.h = selectedAreaRef.value.h;
      } else if (previousArea) {
        area.w = previousArea.w;
        area.h = previousArea.h;
      } else {
        const documentRef = documentRefs.value.find(
          (e: any) => e && e.document && e.document.id === area.attachment_id
        );
        if (!documentRef) {
          return;
        }
        const pageMask = documentRef.pageRefs[area.page].$refs.mask;

        if (drawField.value.type === "checkbox" || drawOption.value) {
          area.w = pageMask.clientWidth / 30 / pageMask.clientWidth;
          area.h = (pageMask.clientWidth / 30 / pageMask.clientWidth) * (pageMask.clientWidth / pageMask.clientHeight);
        } else if (drawField.value.type === "image") {
          area.w = pageMask.clientWidth / 5 / pageMask.clientWidth;
          area.h = (pageMask.clientWidth / 5 / pageMask.clientWidth) * (pageMask.clientWidth / pageMask.clientHeight);
        } else if (drawField.value.type === "signature" || drawField.value.type === "stamp") {
          area.w = pageMask.clientWidth / 5 / pageMask.clientWidth;
          area.h =
            ((pageMask.clientWidth / 5 / pageMask.clientWidth) * (pageMask.clientWidth / pageMask.clientHeight)) / 2;
        } else if (drawField.value.type === "initials") {
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

    drawField.value.areas ||= [];

    const insertBeforeAreaIndex = drawField.value.areas.findIndex((a: any) => {
      return a.attachment_id === area.attachment_id && a.page > area.page;
    });

    if (insertBeforeAreaIndex !== -1) {
      drawField.value.areas.splice(insertBeforeAreaIndex, 0, area);
    } else {
      drawField.value.areas.push(area);
    }

    if (template.value.fields.indexOf(drawField.value) === -1) {
      template.value.fields.push(drawField.value);
    }

    drawField.value = null;
    drawOption.value = null;
    selectedAreaRef.value = area;
    save();
  } else {
    const documentRef = documentRefs.value.find((e: any) => e.document.id === area.attachment_id);
    const pageMask = documentRef.pageRefs[area.page].$refs.mask;

    let type = pageMask.clientWidth * area.w < 35 ? "checkbox" : "text";
    if (type === "checkbox") {
      const previousField = [...template.value.fields].reverse().find((f: any) => f.type === type);
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
        submitter_id: selectedSubmitter.value.id,
        areas: [area]
      };

      template.value.fields.push(field);
      selectedAreaRef.value = area;
      save();
    }
  }
}

function onDropfield(area: any): void {
  const field = {
    name: "",
    id: v4(),
    submitter_id: selectedSubmitter.value.id,
    required: dragField.value.type !== "checkbox",
    ...dragField.value
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

  const previousField = [...template.value.fields].reverse().find((f: any) => f.type === field.type);
  let baseArea: any;
  if (selectedField.value?.type === field.type) {
    baseArea = selectedAreaRef.value;
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
  selectedAreaRef.value = fieldArea;
  template.value.fields.push(field);
  save();
}

function scrollToArea(area: any): void {
  //console.log(documentRefs.value)
  const documentRef = documentRefs.value.find((a: any) => a && a.document && a.document.id === area.attachment_id);
  if (documentRef) {
    documentRef.scrollToArea(area);
    selectedAreaRef.value = area;
  }
}

function baseFetch(path: string, options: RequestInit = {}): Promise<Response> {
  // Normalize path: replace /api/ with /api/v1/ if needed
  let normalizedPath = path;
  if (path.startsWith("/api/") && !path.startsWith("/api/v1/")) {
    normalizedPath = path.replace("/api/", "/api/v1/");
  }
  // Use fetchWithAuth to ensure token is included in headers
  return fetchWithAuth(normalizedPath, {
    ...options,
    headers: { ...fetchOptions.headers, ...options.headers }
  });
}

async function save({ force } = { force: false }): Promise<object> {

  if (!autosave.value && !force) {
    return Promise.resolve({});
  }
  if (!template.value) {
    return Promise.resolve({});
  }

  nextTick(() => {
    const templateBuilder = document.querySelector("template-builder") as HTMLElement;
    if (templateBuilder) {
      templateBuilder.dataset.template = JSON.stringify(template.value);
    }
  });

  pushUndo();

  await baseFetch(`/api/templates/${(template.value as any).id}`, {
    method: "PUT",
    body: JSON.stringify({
      template: {
        name: (template.value as any).name,
        schema: (template.value as any).schema,
        submitters: (template.value as any).submitters,
        fields: (template.value as any).fields
      }
    }),
    headers: { "Content-Type": "application/json" }
  });
  if (onSave.value) {
    onSave.value(template.value);
  }

  return {};
}
</script>
