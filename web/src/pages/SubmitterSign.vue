<template>
  <div class="submitter-sign-page min-h-screen bg-[--color-base-200]">
    <!-- Loading State -->
    <div v-if="isLoading" class="flex h-screen items-center justify-center">
      <span class="loading loading-spinner loading-lg"></span>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="container mx-auto px-4 py-8">
      <div class="alert alert-error">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0 stroke-current" fill="none" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>{{ error }}</span>
      </div>
    </div>

    <!-- Completed State -->
    <div v-else-if="submitter?.status === 'completed'" class="container mx-auto px-4 py-8">
      <div class="mx-auto max-w-2xl rounded-lg border border-[var(--color-base-300)] bg-white">
        <div class="px-6 py-5 text-center">
          <div class="text-success mb-4 text-6xl">✓</div>
          <h2 class="card-title justify-center text-2xl">Document Completed!</h2>
          <p>Thank you for completing this document.</p>
          <p class="text-sm text-[--color-base-content]/60">Completed on: {{ formatDate(submitter.completed_at) }}</p>
        </div>
      </div>
    </div>

    <!-- Declined State -->
    <div v-else-if="submitter?.status === 'declined'" class="container mx-auto px-4 py-8">
      <div class="mx-auto max-w-2xl rounded-lg border border-[var(--color-base-300)] bg-white">
        <div class="px-6 py-5 text-center">
          <div class="text-error mb-4 text-6xl">✕</div>
          <h2 class="card-title justify-center text-2xl">Document Declined</h2>
          <p>This document was declined.</p>
          <p class="text-sm text-[--color-base-content]/60">Declined on: {{ formatDate(submitter.declined_at) }}</p>
        </div>
      </div>
    </div>

    <!-- Signing Form -->
    <div v-else class="container mx-auto px-4 py-8">
      <!-- Header -->
      <div class="mb-6 rounded-lg border border-[var(--color-base-300)] bg-white">
        <div class="px-6 py-5">
          <h1 class="card-title text-2xl">{{ template?.name }}</h1>
          <p v-if="template?.description" class="text-[--color-base-content]/60">{{ template.description }}</p>
          <div class="mt-2 flex gap-2">
            <div class="badge badge-outline">{{ submitter?.name }}</div>
            <div class="badge badge-outline">{{ submitter?.email }}</div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <!-- Document Preview -->
        <div class="lg:col-span-2">
          <div class="rounded-lg border border-[var(--color-base-300)] bg-white">
            <div class="px-6 py-5">
              <h2 class="text-xl font-bold">Document</h2>
              <div class="overflow-hidden rounded-lg border bg-white">
                <div v-for="(doc, docIndex) in template?.documents" :key="doc.id">
                  <div v-for="(page, pageIndex) in doc.preview_images" :key="page.id" class="relative mb-4">
                    <img :src="page.url" :alt="`Page ${pageIndex + 1}`" class="w-full" />
                    <!-- Field Overlays -->
                    <div
                      v-for="field in getFieldsForPage(docIndex, pageIndex)"
                      :key="field.id"
                      class="bg-primary/10 hover:bg-primary/20 border-primary absolute cursor-pointer rounded border-2 transition"
                      :style="getFieldStyle(field, page)"
                      @click="scrollToField(field.id)"
                    >
                      <span class="bg-primary text-primary-content px-1 text-xs">
                        {{ field.name }}
                      </span>
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
              <h2 class="card-title mb-4">Complete Fields</h2>

              <div class="space-y-4">
                <div v-for="field in myFields" :id="`field-${field.id}`" :key="field.id" class="form-control">
                  <label class="label">
                    <span class="label-text font-semibold">
                      {{ field.name }}
                      <span v-if="field.required" class="text-error">*</span>
                    </span>
                  </label>

                  <FieldInput
                    v-model="formData[field.id]"
                    :type="field.type as any"
                    :required="field.required"
                    :options="field.options"
                    :placeholder="field.name"
                    :error="fieldErrors[field.id]"
                    @blur="validateField(field)"
                  />
                </div>
              </div>

              <!-- Actions -->
              <div class="card-actions mt-6 justify-between">
                <button class="btn btn-ghost btn-sm" :disabled="isSubmitting" @click="handleDecline">Decline</button>
                <button
                  class="rounded-[--radius-btn] bg-[--color-primary] px-4 py-2 text-white transition-colors hover:opacity-90 focus:ring-2 focus:ring-[--color-primary] focus:outline-none"
                  :class="{ loading: isSubmitting }"
                  :disabled="!isFormValid || isSubmitting"
                  @click="handleSubmit"
                >
                  {{ isSubmitting ? "Submitting..." : "Complete" }}
                </button>
              </div>

              <!-- Progress -->
              <div class="mt-4">
                <div class="text-center text-sm text-[--color-base-content]/60">
                  {{ completedFieldsCount }} / {{ myFields.length }} fields completed
                </div>
                <progress
                  class="progress progress-primary mt-2 w-full"
                  :value="completedFieldsCount"
                  :max="myFields.length"
                ></progress>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import FieldInput from "@/components/common/FieldInput.vue";
import { fetchWithAuth } from "@/utils/api/auth";

interface Field {
  id: string;
  submitter_id: string;
  name: string;
  type: string;
  required: boolean;
  options?: { id: string; value: string }[];
  areas?: {
    attachment_id: string;
    page: number;
    x: number;
    y: number;
    w: number;
    h: number;
  }[];
}

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
  documents: {
    id: string;
    preview_images: {
      id: string;
      url: string;
      metadata: {
        width: number;
        height: number;
      };
    }[];
  }[];
}

const route = useRoute();

const slug = ref(route.params.slug as string);
const isLoading = ref(true);
const isSubmitting = ref(false);
const error = ref("");

const template = ref<Template | null>(null);
const submitter = ref<Submitter | null>(null);
const formData = ref<Record<string, any>>({});
const fieldErrors = ref<Record<string, string>>({});

const myFields = computed(() => {
  if (!template.value || !submitter.value) {
    return [];
  }
  return template.value.fields.filter((f) => f.submitter_id === submitter.value?.id);
});

const completedFieldsCount = computed(() => {
  return myFields.value.filter((field) => {
    const value = formData.value[field.id];
    if (field.required) {
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
    }
    return true;
  }).length;
});

const isFormValid = computed(() => {
  return completedFieldsCount.value === myFields.value.length && Object.keys(fieldErrors.value).length === 0;
});

onMounted(async () => {
  await loadSubmission();
});

async function loadSubmission(): Promise<void> {
  try {
    isLoading.value = true;
    const response = await fetchWithAuth(`/api/v1/submitters/slug/${slug.value}`);

    if (!response.ok) {
      throw new Error("Submission not found");
    }

    const data = await response.json();
    template.value = data.template;
    submitter.value = data.submitter;

    // Mark as opened
    if (submitter.value?.status === "pending") {
      await fetchWithAuth(`/api/v1/submitters/${submitter.value.id}/open`, {
        method: "POST"
      });
    }

    // Initialize form data
    myFields.value.forEach((field) => {
      if (field.type === "checkbox") {
        formData.value[field.id] = false;
      } else if (field.type === "multiple") {
        formData.value[field.id] = [];
      } else {
        formData.value[field.id] = "";
      }
    });
  } catch (err) {
    error.value = (err as Error).message;
  } finally {
    isLoading.value = false;
  }
}

function getFieldsForPage(docIndex: number, pageIndex: number): Field[] {
  if (!template.value || !submitter.value) {
    return [];
  }
  const doc = template.value.documents[docIndex];
  return myFields.value.filter((field) =>
    field.areas?.some((area) => area.attachment_id === doc.id && area.page === pageIndex)
  );
}

function getFieldStyle(field: Field, page: any): Record<string, string> {
  const area = field.areas?.find((a) => a.attachment_id === page.record_id);
  if (!area) {
    return {};
  }

  const pageWidth = page.metadata.width;
  const pageHeight = page.metadata.height;

  return {
    left: `${(area.x / pageWidth) * 100}%`,
    top: `${(area.y / pageHeight) * 100}%`,
    width: `${(area.w / pageWidth) * 100}%`,
    height: `${(area.h / pageHeight) * 100}%`
  };
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

  if (field.required) {
    if (!value || (typeof value === "string" && value.trim() === "")) {
      fieldErrors.value[field.id] = "This field is required";
      return;
    }
    if (Array.isArray(value) && value.length === 0) {
      fieldErrors.value[field.id] = "Please select at least one option";
      return;
    }
  }

  fieldErrors.value[field.id] = undefined;
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
    const response = await fetchWithAuth(`/api/v1/submitters/${submitter.value.id}/complete`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        fields: formData.value
      })
    });

    if (!response.ok) {
      throw new Error("Failed to submit");
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

  const confirmed = confirm("Are you sure you want to decline this document?");
  if (!confirmed) {
    return;
  }

  isSubmitting.value = true;

  try {
    const response = await fetchWithAuth(`/api/v1/submitters/${submitter.value.id}/decline`, {
      method: "POST"
    });

    if (!response.ok) {
      throw new Error("Failed to decline");
    }

    // Reload to show declined state
    await loadSubmission();
  } catch (err) {
    error.value = (err as Error).message;
  } finally {
    isSubmitting.value = false;
  }
}

function formatDate(dateString?: string): string {
  if (!dateString) {
    return "";
  }
  return new Date(dateString).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric"
  });
}
</script>

<style scoped>
.submitter-sign-page {
  @apply min-h-screen;
}
</style>
