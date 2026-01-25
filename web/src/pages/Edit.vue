<template>
  <div class="edit-page flex h-full min-h-0 flex-col">
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-3xl font-bold">
        <span v-if="template">
          {{ $t("templates.title") }}
          <span class="mx-2 text-gray-500">→</span>
          <span class="text-gray-900">{{ template.name }}</span>
        </span>
        <span v-else>
          {{ $t("templates.title") }}
          <span class="mx-2 text-gray-500">→</span>
          <span class="text-gray-900">{{ $t("templates.editor") }}</span>
        </span>
      </h1>

      <div v-if="!loadingTemplate && template" class="flex items-center gap-2">
        <button
          type="button"
          class="inline-flex items-center rounded-lg border bg-white px-3 py-2 text-sm font-medium hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60"
          :class="[
            lastSaveError
              ? 'border-red-300 text-red-700'
              : !isDirty && lastSavedAt
                ? 'border-green-300 text-green-700'
                : 'border-gray-200 text-gray-700'
          ]"
          :disabled="uploading || isSaving || (!isDirty && !lastSaveError)"
          :title="lastSaveError || ''"
          @click="save({ force: true })"
        >
          <LoadingSpinner v-if="isSaving" class="mr-2 h-4 w-4" />
          <SvgIcon
            v-else-if="!isDirty && lastSavedAt && !lastSaveError"
            name="check-circle"
            class="mr-2 h-4 w-4"
          />
          <SvgIcon v-else-if="lastSaveError" name="error-circle" class="mr-2 h-4 w-4" />
          <span>
            {{
              isSaving
                ? $t("common.loading")
                : lastSaveError
                  ? $t("common.save")
                  : !isDirty && lastSavedAt
                    ? $t("success.saved")
                    : $t("common.save")
            }}
          </span>
        </button>
      </div>
    </div>

    <div v-if="loadingTemplate" class="flex flex-1 items-center justify-center">
      <div class="flex items-center gap-3 rounded-xl bg-white px-5 py-4 text-gray-700">
        <LoadingSpinner class="h-5 w-5" />
        <div class="text-sm font-medium">{{ $t("templates.loadingTemplate") }}</div>
      </div>
    </div>

    <div v-else-if="template && template.schema && template.schema.length > 0" class="flex flex-1 min-h-0 overflow-hidden">
      <!-- Left previews: fixed column; scrolls only when hovered and overflowing -->
      <div ref="previewsRef" class="hidden h-full w-28 flex-none overflow-x-hidden overflow-y-auto pr-3 lg:block">
        <DocumentPreview
          v-for="(item, index) in template && template.schema ? template.schema : []"
          :key="index"
          :with-arrows="(template && template.schema && template.schema.length) > 1"
          :item="item"
          :document="sortedDocuments[index]"
          :editable="editable"
          :template="template"
          @scroll-to="scrollIntoDocument(item)"
          @remove="onDocumentRemove"
          @replace="onDocumentReplace"
          @up="moveDocument(item, -1)"
          @down="moveDocument(item, 1)"
          @change="save"
        />
        <button
          type="button"
          class="mt-2 flex w-full cursor-pointer flex-col items-center justify-center rounded border-2 border-dashed border-gray-300 bg-gray-50/80 py-4 transition-colors hover:border-gray-400 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-gray-400 focus:ring-offset-1 disabled:cursor-not-allowed disabled:opacity-60"
          :class="[{ 'aspect-[210/297]': template?.schema?.length }, 'min-h-[5rem]']"
          :disabled="uploading"
          :aria-label="$t('templates.addPages')"
          @click="showAddPagesModal = true"
        >
          <SvgIcon name="plus" class="h-6 w-6 text-gray-400" />
          <span class="mt-1.5 text-center text-xs font-medium text-gray-500">{{ $t("templates.addPages") }}</span>
        </button>
      </div>

      <!-- Center documents: the ONLY column that should scroll by default -->
      <div class="flex-1 min-h-0 overflow-hidden">
        <div ref="documents" class="h-full overflow-x-hidden overflow-y-auto pr-3.5 pl-0.5">
          <template v-for="document in sortedDocuments" :key="document.id">
            <Document
              :ref="setDocumentRefs"
              :areas-index="fieldAreasIndex[document.id]"
              :selected-submitter="selectedSubmitter"
              :document="document"
              :is-drag="!!dragField"
              :default-fields="[]"
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

      <!-- Right fields panel: fixed column; scrolls only when hovered and overflowing -->
      <div class="relative hidden h-full w-80 flex-none overflow-x-hidden overflow-y-auto pl-0.5 md:block">
        <div v-if="drawField" class="sticky inset-0 z-20 h-full">
          <div class="space-y-4 rounded-lg bg-[--color-base-300] p-5 text-center">
            <p>Draw {{ drawField.name }} field on the document</p>
            <div>
              <button class="base-button" @click="clearDrawField()">Cancel</button>
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
          :selected-field="selectedField"
          :default-submitters="defaultSubmitters"
          :default-fields="defaultFields"
          :only-defined-fields="onlyDefinedFields"
          :signing-mode="signingMode"
          :editable="editable"
          @set-draw="[(drawField = $event.field), (drawOption = $event.option)]"
          @set-drag="dragField = $event"
          @change-submitter="selectedSubmitter = $event"
          @update-signing-mode="signingMode = $event"
          @update-submitter-order="template.submitters = $event"
          @drag-end="dragField = null"
          @scroll-to-area="scrollToArea"
        />
      </div>
    </div>

    <div v-else class="flex h-full w-full items-center justify-center">
      <div class="w-full max-w-2xl rounded-2xl bg-white p-8">
        <div class="mb-6">
          <div class="text-xl font-semibold text-gray-900">{{ $t("templates.uploadPdfToStartEditing") }}</div>
          <div class="mt-1 text-sm text-gray-600">
            {{ $t("templates.uploadPdfToStartEditingHint") }}
          </div>
        </div>

        <label
          for="templateEditFileInput"
          class="relative block h-40 w-full cursor-pointer rounded-xl border-2 border-dashed border-gray-300 hover:bg-gray-50"
          :class="{ 'bg-gray-50 border-blue-400': selectedFile, 'opacity-60 cursor-not-allowed': uploading }"
          @dragover.prevent
          @drop.prevent="handleDrop"
        >
          <div class="absolute inset-0 flex items-center justify-center">
            <div class="flex flex-col items-center text-center">
              <span v-if="!selectedFile" class="flex flex-col items-center">
                <SvgIcon name="cloud-upload" class="h-10 w-10 text-gray-400" />
                <div class="mt-2 text-sm font-medium text-gray-700">{{ $t("templates.clickToUpload") }}</div>
                <div class="text-xs text-gray-500">{{ $t("templates.dragAndDrop") }}</div>
              </span>
              <span v-else class="flex flex-col items-center">
                <SvgIcon name="document" class="h-10 w-10 text-blue-500" />
                <div class="mt-2 text-sm font-medium text-gray-900">{{ selectedFile.name }}</div>
                <div v-if="uploading" class="mt-1 text-xs text-gray-600">{{ $t("templates.uploading") }}</div>
                <button
                  v-else
                  type="button"
                  class="mt-1 text-xs text-red-600 hover:text-red-800"
                  @click.stop="removeSelectedFile"
                >
                  {{ $t("templates.removeFile") }}
                </button>
              </span>
            </div>
          </div>

          <input
            id="templateEditFileInput"
            ref="templateFileInput"
            type="file"
            accept=".pdf"
            class="hidden"
            :disabled="uploading"
            @change="handleFileSelect"
          />
        </label>

        <div v-if="uploadError" class="mt-3 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
          {{ uploadError }}
        </div>
      </div>
    </div>

    <!-- Add pages modal -->
    <FormModal
      ref="addPagesModalRef"
      v-model="showAddPagesModal"
      :title="$t('templates.addPagesTitle')"
      :submit-text="$t('templates.addPagesButton')"
      :on-submit="handleAddPagesSubmit"
      @cancel="resetAddPagesSelection"
    >
      <template #default>
        <div class="space-y-3">
          <div class="text-sm text-gray-600">{{ $t("templates.addPagesHint") }}</div>

          <label
            for="addPagesFileInput"
            class="relative block h-32 w-full cursor-pointer rounded-xl border-2 border-dashed border-gray-300 hover:bg-gray-50"
            :class="{ 'bg-gray-50 border-blue-400': addPagesSelectedFile }"
            @dragover.prevent
            @drop.prevent="handleAddPagesDrop"
          >
            <div class="absolute inset-0 flex items-center justify-center">
              <div class="flex flex-col items-center text-center">
                <span v-if="!addPagesSelectedFile" class="flex flex-col items-center">
                  <SvgIcon name="cloud-upload" class="h-8 w-8 text-gray-400" />
                  <div class="mt-2 text-sm font-medium text-gray-700">{{ $t("templates.clickToUpload") }}</div>
                  <div class="text-xs text-gray-500">{{ $t("templates.dragAndDrop") }}</div>
                </span>
                <span v-else class="flex flex-col items-center">
                  <SvgIcon name="document" class="h-8 w-8 text-blue-500" />
                  <div class="mt-2 text-sm font-medium text-gray-900">{{ addPagesSelectedFile.name }}</div>
                  <button
                    type="button"
                    class="mt-1 text-xs text-red-600 hover:text-red-800"
                    @click.stop="resetAddPagesSelection"
                  >
                    {{ $t("templates.removeFile") }}
                  </button>
                </span>
              </div>
            </div>

            <input
              id="addPagesFileInput"
              ref="addPagesFileInput"
              type="file"
              accept=".pdf"
              class="hidden"
              :disabled="uploading"
              @change="handleAddPagesFileSelect"
            />
          </label>

          <div v-if="addPagesError" class="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {{ addPagesError }}
          </div>
        </div>
      </template>
    </FormModal>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, provide, ref } from "vue";
import { onBeforeRouteLeave, useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import Document from "@/components/template/Document.vue";
import DocumentPreview from "@/components/template/Preview.vue";
import Fields from "@/components/field/List.vue";
import SvgIcon from "@/components/SvgIcon.vue";
import LoadingSpinner from "@/components/ui/LoadingSpinner.vue";
import FormModal from "@/components/common/FormModal.vue";
import type { Template } from "@/models";
import { apiGet, apiPost } from "@/services/api";
import { fetchWithAuth } from "@/utils/auth";
import { fileToBase64Payload } from "@/utils/file";
import { v4 } from "uuid";

const template: any = ref(null);
const signingMode: any = ref("sequential");
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

const onSave = ref();

const defaultSubmitters = ref<any[]>([]);
const defaultFields = ref<string[]>(["text", "signature"]);
const onlyDefinedFields = ref(false);

const fetchOptions = { headers: {} };
const route = useRoute();
const { t } = useI18n();

const templateFileInput = ref<HTMLInputElement | null>(null);
const selectedFile = ref<File | null>(null);
const uploading = ref(false);
const uploadError = ref<string | null>(null);
const loadingTemplate = ref(true);

const showAddPagesModal = ref(false);
const addPagesModalRef = ref<any>(null);
const addPagesFileInput = ref<HTMLInputElement | null>(null);
const addPagesSelectedFile = ref<File | null>(null);
const addPagesError = ref<string | null>(null);

provide("template", template);
provide("save", save);
provide("baseFetch", baseFetch);
provide("selectedAreaRef", selectedAreaRef); // computed(() => selectedAreaRef.value),

// Manual save state (document + form edits)
const isSaving = ref(false);
const saveQueued = ref(false);
const isDirty = ref(false);
const lastSaveError = ref<string | null>(null);
const lastSavedAt = ref<number | null>(null);

function applyLoadedTemplate(tpl: any): void {
  template.value = tpl;

  // Ensure at least one submitter exists (editor requires one)
  if (!template.value.submitters || template.value.submitters.length === 0) {
    template.value.submitters = [
      {
        id: v4(),
        name: "Signer 1",
        colorIndex: 0
      }
    ];
  }

  if (template.value.submitters && template.value.submitters.length > 0) {
    selectedSubmitter.value = template.value.submitters[0];
  }
  signingMode.value = template.value.signing_mode || "sequential";

  // Reset save state when template is loaded
  isDirty.value = false;
  lastSavedAt.value = Date.now();
  lastSaveError.value = null;
}

async function uploadAndApplyPdf(file: File, append = false): Promise<void> {
  const templateId = route.params.id as string;
  if (!templateId) return;

  uploading.value = true;
  uploadError.value = null;

  try {
    const base64 = await fileToBase64Payload(file);
    if (!base64) {
      throw new Error("Failed to read file");
    }

    const res = await apiPost(`/api/v1/templates/${templateId}/from-file`, {
      type: "pdf",
      file_base64: base64,
      append
    });

    if (!res || !res.data) {
      throw new Error("Unexpected response");
    }

    applyLoadedTemplate(res.data);
    // Reset save state after applying PDF
    isDirty.value = false;
    lastSavedAt.value = Date.now();
    lastSaveError.value = null;
    undoStack.value = [JSON.stringify(template.value)];
    redoStack.value = [];
  } catch (err: any) {
    console.error("Failed to upload/apply PDF:", err);
    throw err;
  } finally {
    uploading.value = false;
  }
}

onMounted(async () => {
  loadingTemplate.value = true;
  try {
    // Get template ID from route params
    const templateId = route.params.id as string;

    if (!templateId) {
      console.error("Template ID is required");
      return;
    }

    // Load specific template by ID
    const res = await apiGet<Template>(`/api/v1/templates/${templateId}`);
    // API v1 returns: { message: "template", data: Template }
    if (res && res.data) {
      applyLoadedTemplate(res.data);
    } else {
      console.error("Template not found");
    }
  } catch (error) {
    console.error("Failed to load template:", error);
    // If auth failed, redirect will happen automatically
    if (route.name === "template-edit") {
      // Only show error if still on edit page (not redirected)
      console.error("Could not load template data");
    }
  }

  try {
    if (template.value) {
      undoStack.value = [JSON.stringify(template.value)];
    }
    redoStack.value = [];

    await nextTick();
    document.addEventListener("keyup", onKeyUp);
    window.addEventListener("keydown", onKeyDown);
  } finally {
    loadingTemplate.value = false;
  }
});

onUnmounted(() => {
  document.removeEventListener("keyup", onKeyUp);
  window.removeEventListener("keydown", onKeyDown);
  window.removeEventListener("beforeunload", onBeforeUnload);
  document.removeEventListener("visibilitychange", onVisibilityChange);
});

function buildUpdatePayload(): Record<string, any> {
  return {
    name: template.value?.name,
    schema: template.value?.schema ?? [],
    submitters: template.value?.submitters ?? [],
    fields: template.value?.fields ?? []
  };
}

async function putTemplateUpdate(payload: any, { keepalive }: { keepalive: boolean }): Promise<void> {
  const id = template.value?.id;
  if (!id) return;

  const res = await fetchWithAuth(`/api/v1/templates/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
    credentials: "include",
    keepalive
  });

  if (!res.ok) {
    const data = await res.json().catch(() => null);
    const msg = data?.message || data?.error || `HTTP ${res.status}`;
    throw new Error(msg);
  }
}

async function flushSave({ keepalive } = { keepalive: false }): Promise<void> {
  if (!template.value) return;

  if (isSaving.value) {
    saveQueued.value = true;
    return;
  }

  // Nothing to do
  if (!isDirty.value && !saveQueued.value) {
    return;
  }

  isSaving.value = true;
  lastSaveError.value = null;
  try {
    await putTemplateUpdate(buildUpdatePayload(), { keepalive });
    lastSavedAt.value = Date.now();
    isDirty.value = false;
    if (onSave.value) {
      onSave.value(template.value);
    }
  } catch (err: any) {
    lastSaveError.value = err?.message || "Failed to save";
    isDirty.value = true;
    throw err;
  } finally {
    isSaving.value = false;
    if (saveQueued.value) {
      saveQueued.value = false;
      await flushSave({ keepalive: false });
    }
  }
}

function onBeforeUnload(): void {
  // Fire-and-forget; keepalive lets it continue after navigation/reload.
  if (isDirty.value || saveQueued.value) {
    void flushSave({ keepalive: true });
  }
}

function onVisibilityChange(): void {
  if (document.visibilityState === "hidden") {
    if (isDirty.value || saveQueued.value) {
      void flushSave({ keepalive: true });
    }
  }
}

onBeforeRouteLeave(async () => {
  try {
    await flushSave({ keepalive: false });
  } catch {
    // error is shown in UI; don't block navigation
  }
  return true;
});

onMounted(() => {
  window.addEventListener("beforeunload", onBeforeUnload);
  document.addEventListener("visibilitychange", onVisibilityChange);
});

const selectedField = computed(() => {
  if (!template.value || !template.value.fields) {
    return null;
  }
  return template.value.fields.find((f: any) => f.areas?.includes(selectedAreaRef.value));
});

const sortedDocuments = computed(() => {
  if (!template.value || !template.value.schema || !template.value.documents) {
    return [];
  }
  return template.value.schema.map((item: any) => {
    return template.value.documents.find((doc: any) => doc.id === item.attachment_id);
  });
});

const fieldAreasIndex = computed(() => {
  if (!template.value || !template.value.fields) {
    return {};
  }
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

function undo(): void {
  if (undoStack.value.length > 1) {
    undoStack.value.pop();
    const stringData = undoStack.value[undoStack.value.length - 1];
    const currentStringData = JSON.stringify(template.value);

    if (stringData && stringData !== currentStringData) {
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

function scrollIntoDocument(item: any): void {
  const refElement: any = documentRefs.value.find((e: any) => {
    return e && e.document && e.document.id === item.attachment_id;
  });
  if (refElement?.$el) {
    refElement.$el.scrollIntoView({ behavior: "smooth", block: "start" });
  }
}

function clearDrawField(): void {
  if (drawField.value && !drawOption.value && drawField.value.areas.length === 0) {
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
  const field = template.value.fields.find((f: any) => f.areas?.includes(area));
  field.areas.splice(field.areas.indexOf(area), 1);

  if (!field.areas.length) {
    template.value.fields.splice(template.value.fields.indexOf(field), 1);
  }
  save();
}

function handleSelectSubmitter(submitterId: string): void {
  const submitter = template.value.submitters.find((s: any) => s.id === submitterId);
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
  const documentRefForArea = documentRefs.value.find((e: any) => e?.document?.id === area.attachment_id);
  const maxPage = documentRefForArea?.pageRefs?.length ?? 1;
  const clampedPage = Math.max(0, Math.min(typeof area.page === "number" ? area.page : 0, maxPage - 1));
  area.page = clampedPage;

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
        if (!documentRef) return;
        const maxPage = documentRef.pageRefs.length;
        area.page = Math.max(0, Math.min(area.page, maxPage - 1));
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

    const areaToAdd = {
      ...area,
      page: Math.max(0, Math.min(typeof area.page === "number" ? area.page : 0, maxPage - 1)),
      attachment_id: area.attachment_id || ""
    };

    const insertBeforeAreaIndex = drawField.value.areas.findIndex((a: any) => {
      return a.attachment_id === areaToAdd.attachment_id && a.page > areaToAdd.page;
    });

    if (insertBeforeAreaIndex !== -1) {
      drawField.value.areas.splice(insertBeforeAreaIndex, 0, areaToAdd);
    } else {
      drawField.value.areas.push(areaToAdd);
    }

    if (template.value.fields.indexOf(drawField.value) === -1) {
      template.value.fields.push(drawField.value);
    }

    drawField.value = null;
    drawOption.value = null;
    selectedAreaRef.value = areaToAdd;
    save();
  } else {
    const documentRef = documentRefs.value.find((e: any) => e?.document?.id === area.attachment_id);
    if (!documentRef) return;
    const maxPage = documentRef.pageRefs.length;
    area.page = Math.max(0, Math.min(area.page, maxPage - 1));
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
      // Ensure page number and attachment_id are preserved
      const areaToAdd = {
        ...area,
        page: typeof area.page === 'number' ? area.page : 0,
        attachment_id: area.attachment_id || ''
      };

      const field = {
        name: "",
        id: v4(),
        required: type !== "checkbox",
        type,
        submitter_id: selectedSubmitter.value.id,
        areas: [areaToAdd]
      };

      template.value.fields.push(field);
      selectedAreaRef.value = areaToAdd;
      save();
    }
  }
}

function onDropfield(area: any): void {
  const field = {
    name: "",
    id: v4(),
    submitter_id: selectedSubmitter.value.id,
    required: true,
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
  let baseArea;
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

  // Fix: if baseArea.h is undefined, calculate default height based on field type
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
      // Default for text and other fields
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

function onDocumentRemove(item: any): void {
  if (!template.value) return;
  if (window.confirm("Are you sure?")) {
    if (template.value.schema) {
      template.value.schema.splice(template.value.schema.indexOf(item), 1);
    }
  }

  if (template.value.fields) {
    template.value.fields.forEach((field: any) => {
      [...(field.areas || [])].forEach((area) => {
        if (area.attachment_id === item.attachment_id) {
          field.areas.splice(field.areas.indexOf(area), 1);
        }
      });
    });
  }
  save();
}

function onDocumentReplace({
  replaceSchemaItem,
  schema,
  documents
}: {
  replaceSchemaItem: any;
  schema: any[];
  documents: any[];
}): void {
  if (!template.value) return;
  if (template.value.schema) {
    template.value.schema.splice(template.value.schema.indexOf(replaceSchemaItem), 1, schema[0]);
  }
  if (template.value.documents) {
    template.value.documents.push(...documents);
  }
  if (template.value.fields) {
    template.value.fields.forEach((field: any) => {
      (field.areas || []).forEach((area: any) => {
        if (area.attachment_id === replaceSchemaItem.attachment_id) {
          area.attachment_id = schema[0].attachment_id;
        }
      });
    });
  }
  save();
}

function moveDocument(item: any, direction: any): void {
  if (!template.value || !template.value.schema) return;
  const currentIndex = template.value.schema.indexOf(item);
  if (currentIndex === -1) return;
  template.value.schema.splice(currentIndex, 1);

  if (currentIndex + direction > template.value.schema.length) {
    template.value.schema.unshift(item);
  } else if (currentIndex + direction < 0) {
    template.value.schema.push(item);
  } else {
    template.value.schema.splice(currentIndex + direction, 0, item);
  }
  save();
}

function scrollToArea(area: any): void {
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

const handleFileSelect = async (event: Event) => {
  const input = event.target as HTMLInputElement;
  const file = input?.files?.[0];
  if (!file) {
    selectedFile.value = null;
    return;
  }

  if (file.type !== "application/pdf" && !file.name.toLowerCase().endsWith(".pdf")) {
    uploadError.value = t("templates.invalidFileType");
    selectedFile.value = null;
    if (input) input.value = "";
    return;
  }

  selectedFile.value = file;
  uploadError.value = null;
  try {
    await uploadAndApplyPdf(file, false);
    removeSelectedFile();
  } catch (err: any) {
    uploadError.value = err?.message || t("templates.uploadError");
  }
};

const handleDrop = async (event: DragEvent) => {
  if (uploading.value) return;
  event.preventDefault();
  const file = event.dataTransfer?.files?.[0];
  if (!file) return;

  if (file.type !== "application/pdf" && !file.name.toLowerCase().endsWith(".pdf")) {
    uploadError.value = t("templates.invalidFileType");
    return;
  }

  selectedFile.value = file;
  if (templateFileInput.value) {
    const dt = new DataTransfer();
    dt.items.add(file);
    templateFileInput.value.files = dt.files;
  }
  uploadError.value = null;
  try {
    await uploadAndApplyPdf(file, false);
    removeSelectedFile();
  } catch (err: any) {
    uploadError.value = err?.message || t("templates.uploadError");
  }
};

const removeSelectedFile = () => {
  selectedFile.value = null;
  uploadError.value = null;
  if (templateFileInput.value) templateFileInput.value.value = "";
};

const resetAddPagesSelection = () => {
  addPagesSelectedFile.value = null;
  addPagesError.value = null;
  if (addPagesFileInput.value) addPagesFileInput.value.value = "";
};

const handleAddPagesFileSelect = (event: Event) => {
  const input = event.target as HTMLInputElement;
  const file = input?.files?.[0];
  if (!file) {
    addPagesSelectedFile.value = null;
    return;
  }
  if (file.type !== "application/pdf" && !file.name.toLowerCase().endsWith(".pdf")) {
    addPagesError.value = t("templates.invalidFileType");
    addPagesSelectedFile.value = null;
    input.value = "";
    return;
  }
  addPagesError.value = null;
  addPagesSelectedFile.value = file;
};

const handleAddPagesDrop = (event: DragEvent) => {
  if (uploading.value) return;
  event.preventDefault();
  const file = event.dataTransfer?.files?.[0];
  if (!file) return;

  if (file.type !== "application/pdf" && !file.name.toLowerCase().endsWith(".pdf")) {
    addPagesError.value = t("templates.invalidFileType");
    return;
  }

  addPagesError.value = null;
  addPagesSelectedFile.value = file;
  if (addPagesFileInput.value) {
    const dt = new DataTransfer();
    dt.items.add(file);
    addPagesFileInput.value.files = dt.files;
  }
};

const handleAddPagesSubmit = async (): Promise<void> => {
  if (!addPagesSelectedFile.value) {
    addPagesError.value = t("templates.selectPdfFile");
    // Keep modal open; stop loading state on button
    if (addPagesModalRef.value?.resetSubmitting) addPagesModalRef.value.resetSubmitting();
    return;
  }

  addPagesError.value = null;
  try {
    await uploadAndApplyPdf(addPagesSelectedFile.value, true);
    showAddPagesModal.value = false;
    resetAddPagesSelection();
  } catch (err: any) {
    addPagesError.value = err?.message || t("templates.uploadError");
    if (addPagesModalRef.value?.resetSubmitting) addPagesModalRef.value.resetSubmitting();
  }
};

function updateTemplateBuilderDataset(): void {
  nextTick(() => {
    const templateBuilder = document.querySelector("template-builder") as HTMLElement | null;
    if (templateBuilder && template.value) {
      templateBuilder.dataset.template = JSON.stringify(template.value);
    }
  });
}

async function save({ force } = { force: false }): Promise<object> {
  if (!template.value) return {};

  // Mark as dirty when any change occurs (even without auto-save)
  isDirty.value = true;
  updateTemplateBuilderDataset();
  pushUndo();

  if (!force) {
    return Promise.resolve({});
  }

  try {
    await flushSave({ keepalive: false });
  } catch {
    // error is stored in lastSaveError
  }
  return {};
}
</script>
