<template>
  <div class="submissions-page">
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-3xl font-bold">Submissions</h1>
      <button
        class="rounded-[--radius-btn] bg-[--color-primary] px-4 py-2 text-white transition-colors hover:opacity-90 focus:ring-2 focus:ring-[--color-primary] focus:outline-none"
        @click="openCreateModal"
      >
        + New Submission
      </button>
    </div>

    <!-- Submissions Table -->
    <ResourceTable
      :data="submissions"
      :columns="columns"
      :is-loading="isLoading"
      searchable
      selectable
      :search-keys="['title', 'template.name']"
      search-placeholder="Search submissions..."
      @select="(items: unknown[]) => handleSelect(items as Submission[])"
      @edit="(item: unknown) => handleEdit(item as Submission)"
      @delete="(item: unknown) => handleDelete(item as Submission)"
      @page-change="handlePageChange"
    >
      <template #cell-status="{ value }">
        <span
          class="inline-flex items-center rounded-[--radius-badge] bg-[--color-primary]/10 px-3 py-1 text-sm font-medium text-[--color-primary]"
          :class="getStatusBadgeClass(value)"
        >
          {{ value }}
        </span>
      </template>

      <template #cell-progress="{ item }">
        <div class="flex items-center gap-2">
          <progress
            class="progress progress-primary w-20"
            :value="(item as Submission).completed_count"
            :max="(item as Submission).total_count"
          ></progress>
          <span class="text-xs text-[--color-base-content]/60">
            {{ (item as Submission).completed_count }}/{{ (item as Submission).total_count }}
          </span>
        </div>
      </template>

      <template #actions="{ item }">
        <div class="flex gap-2">
          <button
            v-if="(item as Submission).status === 'draft'"
            class="btn btn-primary btn-sm"
            @click="sendSubmission(item as Submission)"
          >
            Send
          </button>
          <button class="btn btn-ghost btn-sm" @click="viewSubmission(item as Submission)">View</button>
          <button
            v-if="['pending', 'in_progress'].includes((item as Submission).status)"
            class="btn btn-warning btn-sm"
            @click="sendReminder(item as Submission)"
          >
            Remind
          </button>
        </div>
      </template>
    </ResourceTable>

    <!-- Create/Edit Modal -->
    <FormModal v-model="showModal" :title="modalTitle" size="lg" @submit="handleSubmit" @cancel="handleCancel">
      <template #default="{ formData, errors }">
        <div class="space-y-4">
          <div class="form-control">
            <label class="label">
              <span class="label-text">Title</span>
            </label>
            <input
              v-model="formData.title"
              type="text"
              class="input-bordered input"
              :class="{ 'input-error': errors.title }"
              placeholder="Contract for John Doe"
            />
            <label v-if="errors.title" class="label">
              <span class="label-text-alt text-error">{{ errors.title }}</span>
            </label>
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">Template</span>
            </label>
            <select
              v-model="formData.template_id"
              class="select-bordered select"
              :class="{ 'select-error': errors.template_id }"
            >
              <option value="">Select a template</option>
              <option v-for="template in templates" :key="template.id" :value="template.id">
                {{ template.name }}
              </option>
            </select>
            <label v-if="errors.template_id" class="label">
              <span class="label-text-alt text-error">{{ errors.template_id }}</span>
            </label>
          </div>

          <div class="form-control">
            <label class="label">
              <span class="label-text">Submitters (comma separated emails)</span>
            </label>
            <textarea
              v-model="formData.submitters as string"
              class="textarea-bordered textarea"
              :class="{ 'textarea-error': errors.submitters }"
              placeholder="john@example.com, jane@example.com"
              rows="3"
            ></textarea>
            <label v-if="errors.submitters" class="label">
              <span class="label-text-alt text-error">{{ errors.submitters }}</span>
            </label>
          </div>

          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-3">
              <input v-model="formData.send_immediately" type="checkbox" class="toggle toggle-sm" />
              <span class="label-text">Send immediately after creation</span>
            </label>
          </div>
        </div>
      </template>
    </FormModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import ResourceTable from "@/components/common/ResourceTable.vue";
import FormModal from "@/components/common/FormModal.vue";
import { fetchWithAuth } from "@/utils/api/auth";

interface Submission {
  id: string;
  title: string;
  status: string;
  template: { id: string; name: string };
  completed_count: number;
  total_count: number;
  created_at: string;
  updated_at: string;
}

interface Template {
  id: string;
  name: string;
}

const router = useRouter();

const submissions = ref<Submission[]>([]);
const templates = ref<Template[]>([]);
const isLoading = ref(false);
const showModal = ref(false);
const editingSubmission = ref<Submission | null>(null);
const selectedSubmissions = ref<Submission[]>([]);

const columns = [
  { key: "title", label: "Title", sortable: true },
  { key: "template.name", label: "Template", sortable: true },
  { key: "status", label: "Status", sortable: true },
  { key: "progress", label: "Progress" },
  {
    key: "created_at",
    label: "Created",
    sortable: true,
    formatter: (value: unknown): string => new Date(String(value)).toLocaleDateString()
  }
];

const modalTitle = computed(() => {
  return editingSubmission.value ? "Edit Submission" : "Create Submission";
});

onMounted(async () => {
  await Promise.all([loadSubmissions(), loadTemplates()]);
});

async function loadSubmissions(): Promise<void> {
  isLoading.value = true;
  try {
    const response = await fetchWithAuth("/api/v1/submissions");
    if (response.ok) {
      const data = await response.json();
      submissions.value = data.data || [];
    }
  } catch (error) {
    console.error("Failed to load submissions:", error);
  } finally {
    isLoading.value = false;
  }
}

async function loadTemplates(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/templates");
    if (response.ok) {
      const data = await response.json();
      templates.value = data.data || [];
    }
  } catch (error) {
    console.error("Failed to load templates:", error);
  }
}

function openCreateModal(): void {
  editingSubmission.value = null;
  showModal.value = true;
}

function handleSelect(selected: Submission[]): void {
  selectedSubmissions.value = selected;
}

function handleEdit(submission: Submission): void {
  editingSubmission.value = submission;
  showModal.value = true;
}

async function handleDelete(submission: Submission): Promise<void> {
  if (!confirm(`Are you sure you want to delete "${submission.title}"?`)) {
    return;
  }

  try {
    const response = await fetchWithAuth(`/api/v1/submissions/${submission.id}`, {
      method: "DELETE"
    });

    if (response.ok) {
      await loadSubmissions();
    } else {
      alert("Failed to delete submission");
    }
  } catch (error) {
    console.error("Failed to delete submission:", error);
    alert("Failed to delete submission");
  }
}

async function handleSubmit(formData: any): Promise<void> {
  try {
    const url = editingSubmission.value ? `/api/v1/submissions/${editingSubmission.value.id}` : "/api/v1/submissions";

    const method = editingSubmission.value ? "PUT" : "POST";

    const response = await fetchWithAuth(url, {
      method,
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        title: formData.title,
        template_id: formData.template_id,
        submitters: formData.submitters.split(",").map((email: string) => email.trim())
      })
    });

    if (response.ok) {
      showModal.value = false;
      await loadSubmissions();

      // Send immediately if requested
      if (formData.send_immediately && !editingSubmission.value) {
        const data = await response.json();
        await sendSubmission({ id: data.data.id } as Submission);
      }
    } else {
      alert("Failed to save submission");
    }
  } catch (error) {
    console.error("Failed to save submission:", error);
    alert("Failed to save submission");
  }
}

function handleCancel(): void {
  editingSubmission.value = null;
}

function handlePageChange(page: number): void {
  console.log("Page changed:", page);
  // Implement server-side pagination if needed
}

async function sendSubmission(submission: Submission): Promise<void> {
  try {
    const response = await fetchWithAuth(`/api/v1/submissions/${submission.id}/send`, {
      method: "POST"
    });

    if (response.ok) {
      await loadSubmissions();
    } else {
      alert("Failed to send submission");
    }
  } catch (error) {
    console.error("Failed to send submission:", error);
    alert("Failed to send submission");
  }
}

async function sendReminder(submission: Submission): Promise<void> {
  try {
    const response = await fetchWithAuth(`/api/v1/submissions/${submission.id}/remind`, {
      method: "POST"
    });

    if (response.ok) {
      alert("Reminder sent successfully");
    } else {
      alert("Failed to send reminder");
    }
  } catch (error) {
    console.error("Failed to send reminder:", error);
    alert("Failed to send reminder");
  }
}

function viewSubmission(submission: Submission): void {
  router.push(`/submissions/${submission.id}`);
}

function getStatusBadgeClass(status: string): string {
  const classes: Record<string, string> = {
    draft: "badge-ghost",
    pending: "badge-warning",
    in_progress: "badge-info",
    completed: "badge-success",
    expired: "badge-error",
    cancelled: "badge-ghost"
  };
  return classes[status] || "badge-ghost";
}
</script>

<style scoped>
.submissions-page {
  @apply min-h-full bg-[--color-base-200];
}
</style>
