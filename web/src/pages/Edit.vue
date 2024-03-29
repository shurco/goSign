<template>
  <div class="h-full flex pt-4" v-if="template && template.schema.length > 0">
    <div ref="previewsRef" class="overflow-y-auto overflow-x-hidden w-28 flex-none pr-3 hidden lg:block">
      <DocumentPreview v-for="(item, index) in template.schema" :key="index" :with-arrows="template.schema.length > 1" :item="item" :document="sortedDocuments[index]"
        :editable="editable" :template="template" @scroll-to="scrollIntoDocument(item)" @remove="onDocumentRemove" @replace="onDocumentReplace" @up="moveDocument(item, -1)"
        @down="moveDocument(item, 1)" @change="save" />
    </div>

    <div class="w-full overflow-y-hidden md:overflow-y-auto overflow-x-hidden">
      <div ref="documents" class="pr-3.5 pl-0.5">
        <template v-for="document in sortedDocuments" :key="document.id">
          <Document :ref="setDocumentRefs" :areas-index="fieldAreasIndex[document.id]" :selected-submitter="selectedSubmitter" :document="document" :is-drag="!!dragField"
            :default-fields="defaultFields" :allow-draw="!onlyDefinedFields" :draw-field="drawField" :editable="editable" @draw="onDraw" @drop-field="onDropfield"
            @remove-area="removeArea" />
        </template>
      </div>
    </div>

    <div class="relative w-80 flex-none pl-0.5 hidden md:block overflow-y-auto overflow-x-hidden">
      <div v-if="drawField" class="sticky inset-0 h-full z-20">
        <div class="bg-base-300 rounded-lg p-5 text-center space-y-4">
          <p>Draw {{ drawField.name }} field on the document</p>
          <div>
            <button class="base-button" @click="clearDrawField">Cancel</button>
            <a v-if="!drawOption && !drawField.areas.length && !['stamp', 'signature', 'initials'].includes(drawField.type)
              " href="#" class="link block mt-3 text-sm" @click.prevent="[(drawField = null), (drawOption = null)]">
              Or add field without drawing
            </a>
          </div>
        </div>
      </div>

      <Fields ref="fields" :fields="template.fields" :submitters="template.submitters" :selected-submitter="selectedSubmitter" :default-submitters="defaultSubmitters"
        :default-fields="defaultFields" :only-defined-fields="onlyDefinedFields" :editable="editable" @set-draw="[(drawField = $event.field), (drawOption = $event.option)]"
        @set-drag="dragField = $event" @change-submitter="selectedSubmitter = $event" @drag-end="dragField = null" @scroll-to-area="scrollToArea" />
    </div>
  </div>

  <div v-else class="h-full flex pt-4" >
    add new files
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref, computed, provide, onMounted, onUnmounted } from "vue";
import Document from "@/components/template/Document.vue";
import DocumentPreview from "@/components/template/Preview.vue";
import Fields from "@/components/field/List.vue";
import type { Template } from "@/models";
import { apiGet } from "@/utils/api";
import { v4 } from "uuid";

const template: any = ref();
const undoStack: any = ref([]);
const redoStack: any = ref([]);
const lastRedoData: any = ref();

const documentRefs: any = ref([]);
const previewsRef: any = ref([]);

const drawField: any = ref();
const drawOption: any = ref();

const selectedSubmitter: any = ref();
const selectedAreaRef: any = ref();
const editable = ref(true); // or whatever the initial value should be
const dragField = ref();

const isSaving = ref(false); // использовать для отображения статуса сохранения

const onSave = ref();

const defaultSubmitters = ref([]);
const defaultFields = ref(["text", "signature"]);
const onlyDefinedFields = ref(false);

const fetchOptions = { headers: {} };

provide("template", template);
provide("save", save);
provide("baseFetch", baseFetch);
provide("selectedAreaRef", selectedAreaRef); // computed(() => selectedAreaRef.value),

onMounted(async () => {
  apiGet(`/api/templates`).then(res => {
    if (res.success) {
      template.value = res.data as Template;
      selectedSubmitter.value = template.value.submitters[0];
    }
  });

  undoStack.value = [JSON.stringify(template.value)];
  redoStack.value = [];

  await nextTick();
  document.addEventListener("keyup", onKeyUp);
  window.addEventListener("keydown", onKeyDown);
});

onUnmounted(() => {
  document.removeEventListener("keyup", onKeyUp);
  window.removeEventListener("keydown", onKeyDown);
});

const selectedField = computed(() => {
  return template.value.fields.find((f: any) => f.areas?.includes(selectedAreaRef.value));
});

const sortedDocuments = computed(() => {
  return template.value.schema.map((item: any) => {
    return template.value.documents.find((doc: any) => doc.id === item.attachment_id);
  });
});

const fieldAreasIndex = computed(() => {
  const areas: any = {};
  template.value.fields.forEach((f: any) => {
    (f.areas || []).forEach((a: any) => {
      areas[a.attachment_id] ||= {};
      const acc = (areas[a.attachment_id][a.page] ||= []);
      acc.push({ area: a, field: f });
    });
  });
  return areas;
});

function undo() {
  if (undoStack.value.length > 1) {
    undoStack.value.pop();
    const stringData = undoStack.value[undoStack.value.length - 1];
    const currentStringData = JSON.stringify(template.value);

    if (stringData && stringData !== currentStringData) {
      redoStack.value.push(currentStringData);
      Object.assign(template.value, JSON.parse(stringData));
      save;
    }
  }
}

function redo() {
  const stringData = redoStack.value.pop();
  lastRedoData.value = stringData;
  const currentStringData: any = JSON.stringify(template.value);

  if (stringData && stringData !== currentStringData) {
    if (undoStack.value[undoStack.value.length - 1] !== currentStringData) {
      undoStack.value.push(currentStringData);
    }
    Object.assign(template.value, JSON.parse(stringData));
    save;
  }
}

function setDocumentRefs(el: any) {
  if (el) {
    documentRefs.value.push(el);
  }
}

function scrollIntoDocument(item: any) {
  const refElement: any = documentRefs.value.find((e: any) => e.document.id === item.attachment_id);
  if (refElement && refElement.$el) {
    refElement.$el.scrollIntoView({ behavior: "smooth", block: "start" });
  }
}

function clearDrawField() {
  if (drawField.value && !drawOption.value && drawField.value.areas.length === 0) {
    const fieldIndex = template.value.fields.indexOf(drawField.value);

    if (fieldIndex !== -1) {
      template.value.fields.splice(fieldIndex, 1);
    }
  }
  drawField.value = null;
  drawOption.value = null;
}

function onKeyUp(e: KeyboardEvent) {
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

function onKeyDown(event: KeyboardEvent) {
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

function removeArea(area: any) {
  const field = template.value.fields.find((f: any) => f.areas?.includes(area));
  field.areas.splice(field.areas.indexOf(area), 1);

  if (!field.areas.length) {
    template.value.fields.splice(template.value.fields.indexOf(field), 1);
  }
  save;
}

function pushUndo() {
  const stringData: any = JSON.stringify(template.value);

  if (undoStack.value[undoStack.value.length - 1] !== stringData) {
    undoStack.value.push(stringData);
    if (lastRedoData.value !== stringData) {
      redoStack.value = [];
    }
  }
}

function onDraw(area: any) {
  if (drawField.value) {
    if (drawOption.value) {
      const areaWithoutOption = drawField.value.areas?.find((a: any) => !a.option_id);

      if (
        areaWithoutOption &&
        !drawField.value.areas.find((a: any) => a.option_id === drawField.value.options[0].id)
      ) {
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
        const documentRef = documentRefs.value.find((e: any) => e.document.id === area.attachment_id);
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
    save;
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
        areas: [area],
      };

      template.value.fields.push(field);
      selectedAreaRef.value = area;
      save;
    }
  }
}

function onDropfield(area: any) {
  const field = {
    name: "",
    id: v4(),
    submitter_id: selectedSubmitter.value.id,
    required: dragField.value.type !== "checkbox",
    ...dragField.value,
  };

  if (["select", "multiple", "radio"].includes(field.type)) {
    field.options = [{ value: "", id: v4() }];
  }

  if (field.type === "stamp") {
    field.readonly = true;
  }

  if (field.type === "date") {
    field.preferences = {
      format: Intl.DateTimeFormat().resolvedOptions().locale.endsWith("-US") ? "MM/DD/YYYY" : "DD/MM/YYYY",
    };
  }

  const fieldArea: any = {
    x: (area.x - 6) / area.maskW,
    y: area.y / area.maskH,
    page: area.page,
    attachment_id: area.attachment_id,
  };

  const previousField = [...template.value.fields].reverse().find((f: any) => f.type === field.type);
  let baseArea;
  if (selectedField.value?.type === field.type) {
    baseArea = selectedAreaRef.value;
  } else if (previousField?.areas?.length) {
    baseArea = previousField.areas[previousField.areas.length - 1];
  } else {
    if (["checkbox"].includes(field.type)) {
      baseArea = {
        w: area.maskW / 30 / area.maskW,
        h: (area.maskW / 30 / area.maskW) * (area.maskW / area.maskH),
      };
    } else if (field.type === "image") {
      baseArea = {
        w: area.maskW / 5 / area.maskW,
        h: (area.maskW / 5 / area.maskW) * (area.maskW / area.maskH),
      };
    } else if (field.type === "signature" || field.type === "stamp") {
      baseArea = {
        w: area.maskW / 5 / area.maskW,
        h: ((area.maskW / 5 / area.maskW) * (area.maskW / area.maskH)) / 2,
      };
    } else if (field.type === "initials") {
      baseArea = {
        w: area.maskW / 10 / area.maskW,
        h: area.maskW / 35 / area.maskW,
      };
    } else {
      baseArea = {
        w: area.maskW / 5 / area.maskW,
        h: area.maskW / 35 / area.maskW,
      };
    }
  }

  fieldArea.w = baseArea.w;
  fieldArea.h = baseArea.h;
  fieldArea.y = fieldArea.y - baseArea.h / 2;

  if (field.type === "cells") {
    fieldArea.cell_w = baseArea.cell_w || baseArea.w / 5;
  }

  field.areas = [fieldArea];
  selectedAreaRef.value = fieldArea;
  template.value.fields.push(field);
  save;
}

function updateFromUpload({ schema, documents }) {
  template.value.schema.push(...schema);
  template.value.documents.push(...documents);
  nextTick(() => {
    if (previewsRef.value) {
      previewsRef.value.scrollTop = previewsRef.value.scrollHeight;
      scrollIntoDocument(schema[0]);
    }
  });
  if (template.value.name === "New Document") {
    template.value.name = template.value.schema[0].name;
  }
  save;
}

function updateName(value: string) {
  template.value.name = value;
  save;
}

function onDocumentRemove(item: any) {
  if (window.confirm("Are you sure?")) {
    template.value.schema.splice(template.value.schema.indexOf(item), 1);
  }

  template.value.fields.forEach((field: any) => {
    [...(field.areas || [])].forEach((area) => {
      if (area.attachment_id === item.attachment_id) {
        field.areas.splice(field.areas.indexOf(area), 1);
      }
    });
  });
  save;
}

function onDocumentReplace({ replaceSchemaItem, schema, documents }) {
  template.value.schema.splice(template.value.schema.indexOf(replaceSchemaItem), 1, schema[0]);
  template.value.documents.push(...documents);
  template.value.fields.forEach((field: any) => {
    (field.areas || []).forEach((area: any) => {
      if (area.attachment_id === replaceSchemaItem.attachment_id) {
        area.attachment_id = schema[0].attachment_id;
      }
    });
  });
  save;
}

function moveDocument(item: any, direction: any) {
  const currentIndex = template.value.schema.indexOf(item);
  template.value.schema.splice(currentIndex, 1);

  if (currentIndex + direction > template.value.schema.length) {
    template.value.schema.unshift(item);
  } else if (currentIndex + direction < 0) {
    template.value.schema.push(item);
  } else {
    template.value.schema.splice(currentIndex + direction, 0, item);
  }
  save;
}

function maybeShowEmptyTemplateAlert(e: Event) {
  if (!template.value.fields.length) {
    e.preventDefault();
    alert("Please draw fields to prepare the document.");
  }
}

function onSaveClick() {
  if (template.value.fields.length) {
    isSaving.value = true;
    try {
      save;
      console.log("onSaveClick");
      //window.Turbo.visit(`/templates/${template.value.id}`)
    } finally {
      isSaving.value = false;
    }
  } else {
    alert("Please draw fields to prepare the document.");
  }
}

function scrollToArea(area: any) {
  const documentRef = documentRefs.value.find((a: any) => a.document.id === area.attachment_id);
  documentRef.scrollToArea(area);
  selectedAreaRef.value = area;
}

function baseFetch(path: string, options: RequestInit = {}) {
  return fetch(path, {
    ...options,
    headers: { ...fetchOptions.headers, ...options.headers },
  });
}

async function save({ force } = { force: false }) {
  console.log("save");

  if (!force) {
    return Promise.resolve({});
  }

  nextTick(() => {
    const templateBuilder = document.querySelector("template-builder") as HTMLElement;
    if (templateBuilder) {
      templateBuilder.dataset.template = JSON.stringify(template.value);
    }
  });

  pushUndo();

  await baseFetch(`/api/templates/${template.value.id}`, {
    method: "PUT",
    body: JSON.stringify({
      template: {
        name: template.value.name,
        schema: template.value.schema,
        submitters: template.value.submitters,
        fields: template.value.fields,
      },
    }),
    headers: { "Content-Type": "application/json" },
  });
  if (onSave.value) {
    onSave.value(template.value);
  }
}
</script>
