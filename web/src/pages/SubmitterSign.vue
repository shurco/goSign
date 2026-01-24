<template>
  <div class="submitter-sign-page min-h-screen bg-[--color-base-200]">
    <!-- Loading State -->
    <div v-if="isLoading" class="flex h-screen items-center justify-center">
      <span class="loading loading-spinner loading-lg"></span>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="container mx-auto px-4 py-8">
      <div class="alert alert-error">
        <SvgIcon name="error-circle" class="h-6 w-6 shrink-0" />
        <span>{{ error }}</span>
      </div>
    </div>

    <!-- Completed State -->
    <div v-else-if="submitter?.status === 'completed'" class="container mx-auto px-4 py-8">
      <div class="mx-auto max-w-2xl rounded-lg border border-[var(--color-base-300)] bg-white">
        <div class="px-6 py-5 text-center">
          <div class="text-success mb-4 text-6xl">✓</div>
          <h2 class="card-title justify-center text-2xl">{{ t('signing.completedTitle') }}</h2>
          <p>{{ t('signing.completedThanks') }}</p>
          <p class="text-sm text-[--color-base-content]/60">
            {{ t('signing.completedOn') }}: {{ formatDate(submitter.completed_at) }}
          </p>

          <div class="mt-5 flex flex-col items-center gap-2">
            <a
              v-if="submissionStatus === 'completed' && completedDocumentUrl"
              class="btn btn-primary btn-sm"
              :href="completedDocumentUrl"
              target="_blank"
              rel="noopener"
            >
              {{ t('common.download') }}
            </a>
            <p v-else class="text-sm text-[--color-base-content]/60">
              {{ t('signing.waitingForOthers') }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Declined State -->
    <div v-else-if="submitter?.status === 'declined'" class="container mx-auto px-4 py-8">
      <div class="mx-auto max-w-2xl rounded-lg border border-[var(--color-base-300)] bg-white">
        <div class="px-6 py-5 text-center">
          <div class="text-error mb-4 text-6xl">✕</div>
          <h2 class="card-title justify-center text-2xl">{{ t('signing.declinedTitle') }}</h2>
          <p>{{ t('signing.declinedText') }}</p>
          <p class="text-sm text-[--color-base-content]/60">
            {{ t('signing.declinedOn') }}: {{ formatDate(submitter.declined_at) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Email/Name Form (if missing) -->
    <div v-else-if="needsEmailOrName" class="container mx-auto px-4 py-8">
      <div class="mx-auto max-w-2xl rounded-lg border border-[var(--color-base-300)] bg-white">
        <div class="px-6 py-5">
          <h2 class="card-title mb-4 text-2xl">{{ t('signing.enterYourInfo') }}</h2>
          <p class="mb-6 text-[--color-base-content]/60">{{ t('signing.enterYourInfoDescription') }}</p>

          <form @submit.prevent="handleUpdateSubmitter" novalidate>
            <div class="space-y-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text font-semibold">
                  {{ t('auth.firstName') }}
                  <span class="text-error">*</span>
                </span>
              </label>
              <input
                v-model="submitterInfo.name"
                type="text"
                class="input input-bordered"
                :class="{ 'input-error': submitterInfoErrors.name }"
                :placeholder="t('auth.firstName')"
                @blur="validateSubmitterInfo"
                @input="submitterInfoErrors.name = ''"
              />
              <label v-if="submitterInfoErrors.name" class="label">
                <span class="label-text-alt text-error">{{ submitterInfoErrors.name }}</span>
              </label>
            </div>

            <div class="form-control">
              <label class="label">
                <span class="label-text font-semibold">
                  {{ t('auth.email') }}
                  <span class="text-error">*</span>
                </span>
              </label>
              <input
                v-model="submitterInfo.email"
                type="text"
                class="input input-bordered"
                :class="{ 'input-error': submitterInfoErrors.email }"
                :placeholder="t('auth.email')"
                @blur="validateSubmitterInfo"
                @input="submitterInfoErrors.email = ''"
              />
              <label v-if="submitterInfoErrors.email" class="label">
                <span class="label-text-alt text-error">{{ submitterInfoErrors.email }}</span>
              </label>
            </div>

            <div class="card-actions mt-6">
              <Button
                type="submit"
                variant="primary"
                :loading="isUpdatingSubmitter"
                :disabled="!isSubmitterInfoValid || isUpdatingSubmitter"
              >
                {{ t('common.continue') }}
              </Button>
            </div>
          </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Signing Form -->
    <div v-else class="container mx-auto px-4 py-8">
      <!-- Header -->
      <div class="mb-6 bg-white">
        <div class="px-6 py-5">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <h1 class="card-title text-2xl">{{ template?.name }}</h1>
              <p v-if="template?.description" class="text-[--color-base-content]/60">{{ template.description }}</p>
              <div class="mt-2 flex gap-2">
                <div v-if="submitter?.name" class="badge badge-outline">{{ submitter?.name }}</div>
                <div v-if="submitter?.email" class="badge badge-outline">{{ submitter?.email }}</div>
              </div>
            </div>

            <!-- Signing language selector: default to browser language, user can override -->
            <div v-if="showLanguageSelector" class="w-full sm:w-48">
              <label class="mb-1 block text-xs font-medium text-gray-600">{{ t('settings.language') }}</label>
              <select class="select select-bordered select-sm w-full" :value="signingLocale" @change="onSigningLocaleChange">
                <option v-for="(name, code) in SUPPORTED_LOCALES" :key="code" :value="code">
                  {{ name }}
                </option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <!-- Document Preview -->
        <div class="lg:col-span-2">
          <div class="bg-white">
            <div class="px-6 py-5">
              <div class="overflow-hidden">
                <div v-for="(doc, docIndex) in sortedDocuments" :key="doc.id">
                  <div v-for="(page, pageIndex) in getSortedPreviewImages(doc)" :key="page.id" class="relative mb-4">
                    <div class="relative">
                      <img
                        :src="`${page.url}/${page.filename}`"
                        :alt="`Page ${pageIndex + 1}`"
                        :width="page.metadata?.width"
                        :height="page.metadata?.height"
                        class="mb-4 rounded border border-[#e7e2df]"
                        loading="lazy"
                        @load="onImageLoad"
                      />
                      <!-- Field Overlays -->
                      <div
                        v-for="field in getFieldsForPage(doc.id, pageNumberFromPreview(page, pageIndex))"
                        :key="field.id"
                        class="bg-primary/10 hover:bg-primary/20 border-primary absolute cursor-pointer rounded border-2 transition"
                        :style="getFieldStyle(field, doc.id, pageNumberFromPreview(page, pageIndex))"
                        @click="scrollToField(field.id)"
                      >
                        <span class="bg-primary text-primary-content px-1 text-xs">
                          {{ getFieldLabel(field) }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Form Fields -->
        <div class="lg:col-span-1">
          <div class="sticky top-4 rounded-lg border border-[var(--color-base-300)] bg-white">
            <div class="px-6 py-5">
              <div class="space-y-4">
                <div v-for="field in visibleFields" :id="`field-${field.id}`" :key="field.id" class="form-control">
                  <label class="label">
                    <span class="label-text font-semibold">
                      {{ getFieldLabel(field) }}
                      <span v-if="fieldStates[field.id]?.required || field.required" class="text-error">*</span>
                    </span>
                  </label>

                  <FieldInput
                    v-model="formData[field.id]"
                    :type="field.type as any"
                    :required="fieldStates[field.id]?.required || field.required"
                    :disabled="fieldStates[field.id]?.disabled"
                    :options="field.options"
                    :placeholder="getFieldLabel(field)"
                    :error="fieldErrors[field.id]"
                    :formula="field.formula"
                    :calculationType="field.calculationType as 'number' | 'currency' | undefined"
                    :calculatedValue="calculatedValues[field.id]"
                    :cell-count="getCellCount(field)"
                    @blur="validateField(field)"
                  />
                </div>
              </div>

              <!-- Actions -->
              <div class="card-actions mt-6 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                <Button type="button" variant="ghost" :disabled="isSubmitting" @click="handleReset">
                  {{ t('common.reset') }}
                </Button>
                <Button
                  type="button"
                  variant="primary"
                  :loading="isSubmitting"
                  :disabled="!isFormValid || isSubmitting"
                  @click="handleSubmit"
                >
                  {{ t('signing.sign') }}
                </Button>
              </div>

              <!-- Progress -->
              <div class="mt-4">
                <div class="text-center text-sm text-[--color-base-content]/60">
                  {{ t('signing.fieldsProgress', { completed: completedFieldsCount, total: myFields.length }) }}
                </div>
                <progress
                  class="progress progress-primary mt-2 w-full"
                  :value="completedFieldsCount"
                  :max="visibleFields.length"
                ></progress>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Danger zone -->
      <div class="mt-8 rounded-lg border border-red-200 bg-red-50">
        <div class="px-6 py-5">
          <h3 class="text-base font-semibold text-red-900">{{ t('signing.dangerTitle') }}</h3>
          <p class="mt-1 text-sm text-red-800">
            {{ t('signing.declineWarning') }}
          </p>

          <div class="mt-4">
            <Button type="button" variant="error" :disabled="isSubmitting" @click="handleDecline">
              {{ t('signing.decline') }}
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import FieldInput from "@/components/common/FieldInput.vue";
import SvgIcon from "@/components/SvgIcon.vue";
import Button from "@/components/ui/Button.vue";
// Public signer-facing endpoints do not require authentication.
import { useConditions } from "@/composables/useConditions";
import { useFormulas } from "@/composables/useFormulas";
import type { Field } from "@/models/template";
import { SUPPORTED_LOCALES, type Locale } from "@/i18n";
import { fieldNames, subNames } from "@/components/field/constants";

// Field type is imported from @/models/template

interface Submitter {
  id: string;
  name: string;
  email: string;
  slug: string;
  status: string;
  completed_at?: string;
  declined_at?: string;
}

interface Template {
  id: string;
  name: string;
  description?: string;
  fields: Field[];
  schema?: Array<{
    attachment_id: string;
    [key: string]: any;
  }>;
  documents: {
    id: string;
    url?: string;
    preview_images: {
      id: string;
      url?: string; // populated client-side as `${document.url}/${document.id}`
      filename: string;
      metadata: {
        width: number;
        height: number;
      };
    }[];
    metadata?: {
      pdf?: {
        number_of_pages?: number;
      };
    };
  }[];
}

const route = useRoute();
const { t, locale } = useI18n();

const slug = ref(route.params.slug as string);
const isLoading = ref(true);
const isSubmitting = ref(false);
const error = ref("");
const SIGNING_LOCALE_STORAGE_KEY = "signing_locale";
const previousLocale = ref<string | null>(null);
const signingLocale = ref(locale.value as string);
const showLanguageSelector = ref(true);

const template = ref<Template | null>(null);
const submitter = ref<Submitter | null>(null);
const submissionStatus = ref<string>("");
const completedDocumentUrl = ref<string>("");
const formData = ref<Record<string, any>>({});
const fieldErrors = ref<Record<string, string>>({});
const isUpdatingSubmitter = ref(false);
const submitterInfo = ref({ name: "", email: "" });
const submitterInfoErrors = ref<Record<string, string>>({});

const myFields = computed(() => {
  if (!template.value || !submitter.value) {
    return [];
  }
  return template.value.fields.filter((f) => f.submitter_id === submitter.value?.id);
});

// Use conditions composable
const { fieldStates } = useConditions(
  computed(() => template.value?.fields || []),
  formData
);

// Use formulas composable
const { calculatedValues } = useFormulas(
  computed(() => template.value?.fields || []),
  formData
);

// Filter visible fields based on conditions
const visibleFields = computed(() => {
  return myFields.value.filter(field => {
    const state = fieldStates.value[field.id]
    return state ? state.visible : true
  })
});

const completedFieldsCount = computed(() => {
  return visibleFields.value.filter((field) => {
    const value = formData.value[field.id];
    const required = fieldStates.value[field.id]?.required || field.required;
    if (!required) {
      return true;
    }
    
    // Special handling for image/file types
    if (field.type === "image" || field.type === "file") {
      // For image, value is base64 string (starts with "data:")
      // For file, value is filename string
      return typeof value === "string" && value.trim() !== "";
    }
    
    // Special handling for cells type - check if all cells are filled
    if (field.type === "cells") {
      const cellCount = getCellCount(field);
      return typeof value === "string" && value.length === cellCount;
    }
    
    if (typeof value === "string") {
      return value.trim() !== "";
    }
    if (Array.isArray(value)) {
      return value.length > 0;
    }
    if (typeof value === "boolean") {
      return value === true;
    }
    return value !== undefined && value !== null;
  }).length;
});

const isFormValid = computed(() => {
  return completedFieldsCount.value === visibleFields.value.length && Object.keys(fieldErrors.value).length === 0;
});

// Sort documents by schema order (same as editor)
const sortedDocuments = computed(() => {
  if (!template.value || !template.value.documents) {
    return [];
  }
  if (template.value.schema && template.value.schema.length > 0) {
    return template.value.schema.map((item: any) => {
      return template.value?.documents.find((doc: any) => doc.id === item.attachment_id);
    }).filter(Boolean);
  }
  // Fallback to original order if no schema
  return template.value.documents;
});

// Sort preview images within each document (same as editor)
function getSortedPreviewImages(doc: any): any[] {
  if (!doc || !doc.preview_images || doc.preview_images.length === 0) {
    return [];
  }
  
  const numberOfPages = doc.metadata?.pdf?.number_of_pages || doc.preview_images.length;
  const previewImagesIndex = doc.preview_images.reduce(
    (acc: any, e: any) => {
      acc[parseInt(e.filename, 10)] = e;
      return acc;
    },
    {} as Record<number, any>
  );
  
  const lazyloadMetadata = doc.preview_images[doc.preview_images.length - 1].metadata;
  return [...Array(numberOfPages).keys()].map((i) => {
    return (
      previewImagesIndex[i] || {
        metadata: lazyloadMetadata,
        id: Math.random().toString(),
        url: doc.url ? `${doc.url}/${doc.id}` : undefined,
        filename: doc.preview_images[i]?.filename || `${i}`
      }
    );
  });
}

const needsEmailOrName = computed(() => {
  if (!submitter.value) return false;
  return !submitter.value.email || !submitter.value.name;
});

const isSubmitterInfoValid = computed(() => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return (
    submitterInfo.value.name.trim() !== "" &&
    submitterInfo.value.email.trim() !== "" &&
    emailRegex.test(submitterInfo.value.email) &&
    Object.keys(submitterInfoErrors.value).length === 0
  );
});

function detectBrowserSigningLocale(): Locale {
  const browser = (navigator.language || "en").split("-")[0];
  if (browser in SUPPORTED_LOCALES) {
    return browser as Locale;
  }
  return "en";
}

function initialSigningLocale(): Locale {
  const stored = localStorage.getItem(SIGNING_LOCALE_STORAGE_KEY);
  if (stored && stored in SUPPORTED_LOCALES) {
    return stored as Locale;
  }
  return detectBrowserSigningLocale();
}

function applySigningLocale(next: Locale): void {
  signingLocale.value = next;
  locale.value = next;
  document.documentElement.setAttribute("lang", next);
  localStorage.setItem(SIGNING_LOCALE_STORAGE_KEY, next);
}

function onSigningLocaleChange(event: Event): void {
  const value = (event.target as HTMLSelectElement | null)?.value as Locale | undefined;
  if (!value || !(value in SUPPORTED_LOCALES)) {
    return;
  }
  applySigningLocale(value);
}

onMounted(async () => {
  // For public signing we intentionally default to browser language (not app/user locale).
  // We also keep this preference isolated from the admin UI locale.
  previousLocale.value = locale.value as string;
  applySigningLocale(initialSigningLocale());
  await loadSubmission();
});

onUnmounted(() => {
  // Restore original app locale when leaving the signing page.
  if (previousLocale.value) {
    locale.value = previousLocale.value;
    document.documentElement.setAttribute("lang", previousLocale.value);
  }
});

function initializeFormData(): void {
  // Reset field values for the current submitter fields.
  // This also clears signature/initials because SignatureInput watches v-model.
  const next: Record<string, any> = {};

  myFields.value.forEach((field) => {
    if (field.type === "checkbox") {
      next[field.id] = false;
      return;
    }
    if (field.type === "multiple" || (field as any).type === "multi_select") {
      next[field.id] = [];
      return;
    }
    // Default: empty string (text/number/date/signature/initials/etc.)
    next[field.id] = "";
  });

  formData.value = next;
  fieldErrors.value = {};
}

function getFieldLabel(field: Field): string {
  // Prefer i18n label in this order:
  // 1) Field-level translations (field.translations[locale])
  // 2) Field.label
  // 3) Field.name
  // 4) Generated default name based on field type and submitter
  // 5) Field.id (last resort)
  const loc = (signingLocale.value || locale.value || "en").toString();
  const anyField: any = field as any;

  const fieldTranslations = anyField.translations as Record<string, string> | undefined;
  const translated = fieldTranslations?.[loc];
  if (translated && translated.trim() !== "") {
    return translated;
  }

  if (anyField.label && String(anyField.label).trim() !== "") {
    return String(anyField.label);
  }
  if (field.name && field.name.trim() !== "") {
    return field.name;
  }

  // Generate default name if field.name is empty
  const defaultName = generateDefaultFieldName(field);
  if (defaultName) {
    return defaultName;
  }

  return field.id;
}

function generateDefaultFieldName(field: Field): string {
  if (!template.value || !submitter.value) {
    return "";
  }

  // Get submitter index for party name
  const submitterIndex = template.value.submitters.findIndex((s: any) => s.id === submitter.value?.id);
  const partyName = subNames[submitterIndex]?.replace(" Party", "") || "First";

  // Get type name from constants
  const typeName = fieldNames[field.type] || "Field";

  // Count how many fields of this type and party already exist
  const sameTypeAndPartyFields = template.value.fields.filter(
    (f: any) => f.type === field.type && f.submitter_id === submitter.value?.id && f.id !== field.id
  );

  const fieldNumber = sameTypeAndPartyFields.length + 1;

  return `${partyName} ${typeName} ${fieldNumber}`;
}

function getCellCount(field: Field): number {
  if (field.type !== "cells" || !field.areas || field.areas.length === 0) {
    return 6; // Default cell count
  }

  const area = field.areas[0] as any;
  const cellWidth = area.cell_w;
  const areaWidth = area.w;

  if (!cellWidth || !areaWidth) {
    return 6; // Default cell count
  }

  // Calculate number of cells based on cell_w and area width
  // Formula: while (currentWidth + (cellWidth + cellWidth / 4) < areaWidth)
  let currentWidth = 0;
  let count = 0;
  while (currentWidth + (cellWidth + cellWidth / 4) < areaWidth) {
    currentWidth += cellWidth;
    count++;
  }

  return Math.max(count, 1); // At least 1 cell
}

function normalizeTemplateForSigning(tpl: Template | null): void {
  if (!tpl) {
    return;
  }

  // Ensure preview images have a usable base URL (same convention as builder `Document.vue`).
  for (const doc of tpl.documents || []) {
    const base = (doc.url || "/drive/pages").replace(/\/$/, "");
    const docBase = `${base}/${doc.id}`;
    for (const img of doc.preview_images || []) {
      // In API, preview_images do not include url. We add it here.
      (img as any).url ||= docBase;
    }
  }

  // Ensure field areas have a height.
  // Backend historically used `z` for height. Some parts of the UI expect `h`.
  for (const f of tpl.fields || []) {
    for (const a of (f.areas as any[]) || []) {
      if (!a) continue;
      if (a.h === undefined && a.z !== undefined) {
        a.h = a.z;
      }
    }
  }
}

function pageNumberFromPreview(preview: any, fallbackIndex: number): number {
  const n = Number.parseInt(String(preview?.filename ?? ""), 10);
  return Number.isFinite(n) ? n : fallbackIndex;
}

async function loadSubmission(): Promise<void> {
  try {
    isLoading.value = true;
    const response = await fetch(`/public/sign/${slug.value}`);

    if (!response.ok) {
      throw new Error(t("signing.submissionNotFound"));
    }

    const data = await response.json();
    // API returns: { success, message, data: { template, submitter } }
    const payload = data.data || data;
    template.value = payload.template;
    submitter.value = payload.submitter;
    submissionStatus.value = String(payload.submission_status || "");
    completedDocumentUrl.value = String(payload.completed_document_url || "");
    normalizeTemplateForSigning(template.value);

    // Mark as opened
    if (submitter.value?.status === "pending") {
      await fetch(`/public/sign/${slug.value}/open`, {
        method: "POST"
      });
    }

    // Initialize form data
    initializeFormData();

    // Pre-fill submitter info if available
    if (submitter.value) {
      submitterInfo.value.name = submitter.value.name || "";
      submitterInfo.value.email = submitter.value.email || "";
    }
  } catch (err) {
    error.value = (err as Error).message;
  } finally {
    isLoading.value = false;
  }
}

function getFieldsForPage(docId: string, pageNumber: number): Field[] {
  if (!template.value || !submitter.value) {
    return [];
  }
  return myFields.value.filter((field) =>
    field.areas?.some((area: any) => area?.attachment_id === docId && area?.page === pageNumber)
  );
}

function getFieldStyle(field: Field, docId: string, pageNumber: number): Record<string, string> {
  const area: any = field.areas?.find((a: any) => a?.attachment_id === docId && a?.page === pageNumber);
  if (!area) {
    return {};
  }

  const x = Number(area.x) || 0;
  const y = Number(area.y) || 0;
  const w = Number(area.w) || 0;
  const h = Number(area.h ?? area.z) || 0;

  return {
    left: `${x * 100}%`,
    top: `${y * 100}%`,
    width: `${w * 100}%`,
    height: `${h * 100}%`
  };
}

function onImageLoad(e: Event): void {
  const target = e.target as HTMLImageElement;
  target.setAttribute("width", target.naturalWidth.toString());
  target.setAttribute("height", target.naturalHeight.toString());
}

function scrollToField(fieldId: string): void {
  const element = document.getElementById(`field-${fieldId}`);
  if (element) {
    element.scrollIntoView({ behavior: "smooth", block: "center" });
    element.classList.add("ring-2", "ring-primary", "rounded");
    setTimeout(() => {
      element.classList.remove("ring-2", "ring-primary");
    }, 2000);
  }
}

function validateField(field: Field): void {
  const value = formData.value[field.id];

  const required = fieldStates.value[field.id]?.required || field.required;
  if (required) {
    if (!value || (typeof value === "string" && value.trim() === "")) {
      fieldErrors.value[field.id] = t("signing.required");
      return;
    }
    if (Array.isArray(value) && value.length === 0) {
      fieldErrors.value[field.id] = t("signing.selectAtLeastOne");
      return;
    }
  }

  delete fieldErrors.value[field.id];
}

async function handleSubmit(): Promise<void> {
  if (!submitter.value || isSubmitting.value) {
    return;
  }

  // Validate all fields
  myFields.value.forEach((field) => validateField(field));

  if (!isFormValid.value) {
    return;
  }

  isSubmitting.value = true;

  try {
    const response = await fetch(`/public/sign/${slug.value}/complete`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        fields: formData.value
      })
    });

    if (!response.ok) {
      throw new Error(t("signing.submitFailed"));
    }

    // Reload to show completed state
    await loadSubmission();
  } catch (err) {
    error.value = (err as Error).message;
  } finally {
    isSubmitting.value = false;
  }
}

async function handleDecline(): Promise<void> {
  if (!submitter.value || isSubmitting.value) {
    return;
  }

  const confirmed = confirm(t("signing.declineConfirm"));
  if (!confirmed) {
    return;
  }

  isSubmitting.value = true;

  try {
    const response = await fetch(`/public/sign/${slug.value}/decline`, {
      method: "POST"
    });

    if (!response.ok) {
      throw new Error(t("signing.declineFailed"));
    }

    // Reload to show declined state
    await loadSubmission();
  } catch (err) {
    error.value = (err as Error).message;
  } finally {
    isSubmitting.value = false;
  }
}

function handleReset(event?: Event): void {
  event?.preventDefault?.();
  if (isSubmitting.value) {
    return;
  }

  const confirmed = confirm(t("signing.resetConfirm"));
  if (!confirmed) {
    return;
  }

  initializeFormData();
}

function formatDate(dateString?: string | null): string {
  if (!dateString) return "—";

  const d = new Date(dateString);
  if (Number.isNaN(d.getTime())) return "—";

  const loc = (signingLocale.value || locale.value || "en").toString();
  return d.toLocaleString(loc, {
    year: "numeric",
    month: "long",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit"
  });
}

function validateSubmitterInfo(): void {
  submitterInfoErrors.value = {};

  if (!submitterInfo.value.name || submitterInfo.value.name.trim() === "") {
    submitterInfoErrors.value.name = t("signing.required");
  }

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!submitterInfo.value.email || submitterInfo.value.email.trim() === "") {
    submitterInfoErrors.value.email = t("signing.required");
  } else if (!emailRegex.test(submitterInfo.value.email)) {
    submitterInfoErrors.value.email = t("signing.invalidEmail");
  }
}

async function handleUpdateSubmitter(event?: Event): Promise<void> {
  event?.preventDefault?.();
  
  if (isUpdatingSubmitter.value) {
    return;
  }

  validateSubmitterInfo();

  if (!isSubmitterInfoValid.value) {
    return;
  }

  isUpdatingSubmitter.value = true;
  error.value = "";

  try {
    const response = await fetch(`/public/sign/${slug.value}/update`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        name: submitterInfo.value.name.trim(),
        email: submitterInfo.value.email.trim()
      })
    });

    // Check content type before parsing
    const contentType = response.headers.get("content-type");
    let data: any = {};
    
    if (contentType && contentType.includes("application/json")) {
      try {
        data = await response.json();
      } catch (parseErr) {
        // If JSON parsing fails, read as text
        const text = await response.text();
        error.value = text || t("signing.updateFailed");
        return;
      }
    } else {
      // If not JSON, read as text
      const text = await response.text();
      error.value = text || t("signing.updateFailed");
      return;
    }

    if (!response.ok) {
      // Try to extract validation errors
      const errorMsg = data.message || data.error || t("signing.updateFailed");
      
      // Check if it's an email validation error
      if (errorMsg.toLowerCase().includes("email") && errorMsg.toLowerCase().includes("valid")) {
        submitterInfoErrors.value.email = t("signing.invalidEmail");
      } else if (errorMsg.toLowerCase().includes("name") || errorMsg.toLowerCase().includes("required")) {
        submitterInfoErrors.value.name = t("signing.required");
      } else {
        error.value = errorMsg;
      }
      return;
    }

    // Reload to show signing form
    await loadSubmission();
  } catch (err) {
    error.value = (err as Error).message || t("signing.updateFailed");
  } finally {
    isUpdatingSubmitter.value = false;
  }
}
</script>

<style scoped>
.submitter-sign-page {
  @apply min-h-screen;
}
</style>
