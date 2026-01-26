<template>
  <div class="submissions-page">
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold">{{ $t('submissions.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ $t('submissions.description') }}</p>
      </div>
      <div class="flex items-center gap-3">
        <Button variant="primary" @click="openCreateModal">
          <SvgIcon name="plus" class="mr-2 h-5 w-5" />
          {{ $t('submissions.newSubmission') }}
        </Button>
      </div>
    </div>

    <!-- Submissions Table -->
    <ResourceTable
      :data="submissions"
      :columns="columns"
      :is-loading="isLoading"
      searchable
      :search-keys="['template_name', 'status']"
      :search-placeholder="$t('submissions.searchSubmissions')"
      @page-change="handlePageChange"
    >
      <template #cell-template_name="{ value, item }">
        <button
          type="button"
          class="link text-left"
          @click="openCompletedDocument(item as Signing)"
        >
          {{ String(value || "") }}
        </button>
      </template>

      <template #cell-status="{ value, item }">
        <button type="button" class="inline-flex cursor-pointer" @click="openStatusHistory(item as Signing)">
          <Badge size="sm" :variant="getBadgeVariantForSubmissionStatus(value)">
            {{ statusLabel(value) }}
          </Badge>
        </button>
      </template>

      <template #cell-progress="{ item }">
        <button type="button" class="flex cursor-pointer items-center gap-2 text-left" @click="openStatusHistory(item as Signing)">
          <progress
            class="progress progress-primary w-20"
            :value="(item as Signing).completed_count"
            :max="(item as Signing).total_count"
          ></progress>
          <span class="text-xs text-[--color-base-content]/60">
            {{ (item as Signing).completed_count }}/{{ (item as Signing).total_count }}
          </span>
        </button>
      </template>

      <template #actions="{ item }">
        <div class="flex items-center justify-end gap-2">
          <button
            :class="ICON_BUTTON_CLASS"
            type="button"
            :title="t('signing.links')"
            :aria-label="t('signing.links')"
            @click="openLinksModal(item as Signing)"
          >
            <SvgIcon name="link" :class="ICON_SVG_CLASS" />
          </button>

          <button
            v-if="String((item as Signing).status) === 'completed'"
            :class="ICON_BUTTON_CLASS"
            type="button"
            :title="t('common.download')"
            :aria-label="t('common.download')"
            @click="openCompletedDocument(item as Signing)"
          >
            <SvgIcon name="document" :class="ICON_SVG_CLASS" />
          </button>
        </div>
      </template>
    </ResourceTable>

    <!-- Create submission modal -->
    <FormModal
      v-model="showModal"
      :title="modalTitle"
      size="lg"
      :submit-text="$t('common.save')"
      :cancel-text="$t('common.cancel')"
      :onSubmit="handleSubmit"
    >
      <template #default="{ formData, errors }">
        <div class="space-y-6">
          <p class="text-sm text-[var(--color-base-content)]/70">
            {{ $t('submissions.createDescription') }}
          </p>

          <FormControl
            :label="$t('submissions.template')"
            :error="errors.template_id"
          >
            <Select
              :model-value="(formData as any).template_id"
              :error="!!errors.template_id"
              @update:model-value="(v) => { const id = String(v ?? ''); (formData as any).template_id = id; onTemplateChange(id); }"
            >
              <option value="">{{ $t('submissions.selectTemplate') }}</option>
              <option v-for="template in templates" :key="template.id" :value="template.id">
                {{ getTemplateDisplayName(template) }}
              </option>
            </Select>
          </FormControl>

          <template v-if="expectedSubmittersCount > 0">
            <SigningModeSelector
              :signing-mode="createSigningMode"
              :submitters="createSubmittersForOrder"
              :editable="true"
              :hide-order-list="true"
              @update:signing-mode="createSigningMode = $event"
              @update:submitter-order="onCreateSubmitterOrder($event)"
            />

            <FormControl
              :label="$t('submissions.signers')"
              :hint="createSigningMode === 'sequential' && createSubmitters.length > 1 ? $t('submissions.signersOrderDragHint') : $t('submissions.signersHint')"
            >
              <div
                class="space-y-4"
                @dragover.prevent
                @drop="onSignerDrop"
              >
                <template v-for="(s, idx) in createSubmitters" :key="idx">
                  <div
                    v-if="insertBeforeIndex === idx && draggedSignerIndex !== null"
                    class="h-1 flex-shrink-0 rounded-full bg-[var(--color-primary)] opacity-90"
                    aria-hidden="true"
                  />
                  <div
                    :data-signer-index="idx"
                    class="rounded-lg border border-[var(--color-base-300)] bg-[var(--color-base-100)]/50 p-4 transition-colors"
                    :class="{
                      'border-[var(--color-primary)]/50': createSigningMode === 'sequential' && createSubmitters.length > 1,
                      'ring-2 ring-[var(--color-primary)]/40': draggedSignerIndex === idx
                    }"
                    @dragenter.prevent="onSignerDragOver($event, idx)"
                    @dragover.prevent="onSignerDragOver($event, idx)"
                    @drop.prevent="onSignerDrop"
                  >
                  <div
                    class="mb-3 flex items-center justify-between gap-2"
                    :class="{ 'cursor-move': createSigningMode === 'sequential' && createSubmitters.length > 1 }"
                    :draggable="createSigningMode === 'sequential' && createSubmitters.length > 1"
                    @dragstart="onSignerDragStart($event, idx)"
                    @dragend="onSignerDragEnd"
                  >
                    <span class="text-sm font-medium text-[var(--color-base-content)]">
                      {{ idx + 1 }}. {{ $t('signing.signer') }} {{ idx + 1 }}
                    </span>
                    <SvgIcon
                      v-if="createSigningMode === 'sequential' && createSubmitters.length > 1"
                      name="drag"
                      class="h-4 w-4 shrink-0 text-[var(--color-base-content)]/50 pointer-events-none"
                      width="16"
                      height="16"
                    />
                  </div>
                  <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
                    <div v-for="field in signerFieldKeys" :key="field.key">
                      <label class="mb-1 block text-xs font-medium text-[var(--color-base-content)]/80">
                        {{ $t(field.labelKey) }}
                      </label>
                      <input
                        v-model="s[field.key]"
                        :type="field.type"
                        class="w-full rounded-lg border border-[var(--color-base-300)] bg-[var(--color-base-100)] px-3 py-2 text-sm text-[var(--color-base-content)] transition-colors hover:border-[var(--color-base-content)]/20 focus:border-[var(--color-primary)] focus:outline-2 focus:outline-offset-0 focus:outline-[var(--color-primary)]"
                        :placeholder="$t('common.optional')"
                      />
                    </div>
                  </div>
                </div>
                </template>
                <div
                  v-if="insertBeforeIndex === createSubmitters.length && draggedSignerIndex !== null"
                  class="h-1 flex-shrink-0 rounded-full bg-[var(--color-primary)] opacity-90"
                  aria-hidden="true"
                />
              </div>
            </FormControl>
          </template>

          <Alert v-if="createdLinks.length" variant="success" class="mt-4">
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <span class="font-medium">{{ t('signing.links') }}</span>
                <button
                  type="button"
                  class="rounded p-1.5 transition-colors hover:bg-[var(--color-base-content)]/10"
                  :title="t('signing.copyAllLinks')"
                  :aria-label="t('signing.copyAllLinks')"
                  @click="copyAllCreatedLinks"
                >
                  <SvgIcon name="copy" class="h-5 w-5" />
                </button>
              </div>
              <p class="text-sm opacity-90">{{ t('signing.sendEachLinkHint') }}</p>
              <div class="space-y-2">
                <div
                  v-for="l in createdLinks"
                  :key="l.slug"
                  class="flex gap-2 rounded border border-[var(--color-base-content)]/20 bg-[var(--color-base-100)] p-2"
                >
                  <input
                    class="min-w-0 flex-1 rounded border-0 bg-transparent px-2 py-1.5 text-sm focus:outline-none"
                    :value="l.full_url"
                    readonly
                  />
                  <div class="flex gap-1">
                    <button
                      type="button"
                      class="rounded p-1.5 transition-colors hover:bg-[var(--color-base-content)]/10"
                      :title="t('signing.copyLink')"
                      :aria-label="t('signing.copyLink')"
                      @click="copyText(l.full_url)"
                    >
                      <SvgIcon name="copy" class="h-4 w-4" />
                    </button>
                    <button
                      type="button"
                      class="rounded p-1.5 transition-colors hover:bg-[var(--color-base-content)]/10"
                      :title="t('submissionStatus.signers.openLink')"
                      :aria-label="t('submissionStatus.signers.openLink')"
                      @click="openUrl(l.full_url)"
                    >
                      <SvgIcon name="arrow-left" class="h-4 w-4 rotate-180" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </Alert>
        </div>
      </template>
      <template #footer="{ submit, cancel, isSubmitting }">
        <div class="flex justify-end gap-3">
          <Button variant="ghost" :disabled="isSubmitting" @click="cancel">
            {{ $t('common.cancel') }}
          </Button>
          <Button variant="primary" :loading="isSubmitting" :disabled="isSubmitting" @click="submit">
            {{ $t('common.save') }}
          </Button>
        </div>
      </template>
    </FormModal>

    <!-- Links Modal (available anytime after creation) -->
    <Modal
      v-model="linksModalOpen"
      :title="activeSigning ? t('signing.linksTitleWithTemplate', { template: activeSigning.template_name }) : t('signing.links')"
      size="lg"
    >
      <template #header>
        <span>{{ activeSigning ? t('signing.linksTitleWithTemplate', { template: activeSigning.template_name }) : t('signing.links') }}</span>
      </template>

      <div v-if="activeSigning" class="space-y-4">
        <ResourceTable
          :data="activeSigning.submitters"
          :columns="signerColumns"
          :has-actions="true"
          :show-edit="false"
          :show-delete="false"
          :show-filters="false"
          :show-pagination="false"
          :searchable="false"
          :selectable="false"
          :page-size="100"
        >
          <template #cell-index="{ item }">
            {{ (activeSigning.submitters.findIndex((s) => s.id === (item as any).id) ?? 0) + 1 }}
          </template>

          <template #cell-signer="{ item }">
            <div class="font-medium">{{ (item as any).name || 'Signer' }}</div>
            <div v-if="(item as any).phone" class="text-xs text-[--color-base-content]/60">{{ (item as any).phone }}</div>
          </template>

          <template #cell-status="{ value }">
            <Badge size="sm" :variant="getBadgeVariantForSubmitterStatus(value)">
              {{ statusLabel(value) }}
            </Badge>
          </template>

          <template #cell-link="{ item }">
            <input class="input input-bordered w-full min-w-[360px]" :value="fullSignerUrl((item as any).slug)" readonly />
          </template>

          <template #actions="{ item }">
            <div class="flex items-center gap-2">
            <button
              :class="ICON_BUTTON_CLASS"
              type="button"
              :title="t('signing.copyLink')"
              :aria-label="t('signing.copyLink')"
              @click="copyText(fullSignerUrl((item as any).slug))"
            >
              <SvgIcon name="copy" :class="ICON_SVG_CLASS" />
            </button>
            </div>
          </template>
        </ResourceTable>
      </div>
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import ResourceTable from "@/components/common/ResourceTable.vue";
import FormModal from "@/components/common/FormModal.vue";
import Modal from "@/components/ui/Modal.vue";
import Button from "@/components/ui/Button.vue";
import Badge from "@/components/ui/Badge.vue";
import Select from "@/components/ui/Select.vue";
import FormControl from "@/components/ui/FormControl.vue";
import Alert from "@/components/ui/Alert.vue";
import SigningModeSelector from "@/components/field/SigningModeSelector.vue";
import SvgIcon from "@/components/SvgIcon.vue";
import { fetchWithAuth } from "@/utils/auth";
import { apiGet } from "@/services/api";
import { getBadgeVariantForSubmissionStatus, getBadgeVariantForSubmitterStatus, getI18nStatusKey } from "@/utils/status";
import { openBlobInNewTab } from "@/utils/file";

const ICON_BUTTON_CLASS =
  "cursor-pointer rounded-full p-1.5 text-[--color-base-content]/50 transition-colors hover:text-[--color-base-content]";
const ICON_SVG_CLASS = "h-5 w-5 stroke-[2]";

const signerFieldKeys = [
  { key: "name", labelKey: "submissions.signerName", type: "text" },
  { key: "email", labelKey: "submissions.signerEmail", type: "email" },
  { key: "phone", labelKey: "submissions.signerPhone", type: "text" }
] as const;

interface CreatedLink {
  submitter_id: string;
  slug: string;
  direct_url: string;
  full_url: string;
}

interface Signing {
  submission_id: string;
  template_id: string;
  template_name: string;
  created_at: string;
  status: string;
  completed_count: number;
  total_count: number;
  submitters: Array<{ id: string; name: string; phone: string; email: string; slug: string; status: string }>;
  links: Array<{ submitter_id: string; slug: string; direct_url: string }>;
}

interface Template {
  id: string;
  name: string;
  folder_id?: string;
}

const { t, te } = useI18n();
const router = useRouter();

const submissions = ref<Signing[]>([]);
const templates = ref<Template[]>([]);
const folders = ref<{ id: string; name: string }[]>([]);
const isLoading = ref(false);
const showModal = ref(false);

const expectedSubmittersCount = ref(0);
const createSubmitters = ref<Array<{ name: string; email: string; phone: string }>>([]);
const createSigningMode = ref<"sequential" | "parallel">("parallel");
const createdLinks = ref<CreatedLink[]>([]);

const linksModalOpen = ref(false);
const activeSigning = ref<Signing | null>(null);
const draggedSignerIndex = ref<number | null>(null);
/** Drop indicator: insert before this index (0 = before first, length = after last). Null = hide. */
const insertBeforeIndex = ref<number | null>(null);

let loadSubmissionsPromise: Promise<void> | null = null;
let loadTemplatesPromise: Promise<void> | null = null;

const columns = computed(() => [
  { key: "template_name", label: t('submissions.template'), sortable: true },
  { key: "status", label: t('submissions.status'), sortable: true },
  { key: "progress", label: t('signing.progress') },
  {
    key: "created_at",
    label: t('submissions.created'),
    sortable: true,
    formatter: (value: unknown): string => new Date(String(value)).toLocaleDateString()
  }
]);

const signerColumns = computed(() => [
  { key: "index", label: "#", sortable: false, headerClass: "w-16" },
  { key: "signer", label: "Signer", sortable: false },
  { key: "status", label: "Status", sortable: false, headerClass: "w-32" },
  { key: "link", label: "Link", sortable: false },
]);

const createSubmittersForOrder = computed(() =>
  createSubmitters.value.map((s, idx) => ({
    id: String(idx),
    name: s.name || `${t("signing.signer")} ${idx + 1}`,
    colorIndex: idx % 8
  }))
);

const modalTitle = computed(() => {
  return t('submissions.create');
});

onMounted(async () => {
  await Promise.all([loadSubmissions(), loadTemplates(), loadFolders()]);
});

async function loadSubmissions(): Promise<void> {
  // Prevent multiple simultaneous loads
  if (isLoading.value || loadSubmissionsPromise) {
    return loadSubmissionsPromise || Promise.resolve();
  }

  isLoading.value = true;
  loadSubmissionsPromise = (async () => {
    try {
      const response = await fetchWithAuth("/api/v1/signing-links");
      if (response.ok) {
        const data = await response.json();
        const payload = data.data || data;
        submissions.value = (payload.items || []) as Signing[];

        // Keep active signing in sync after refresh.
        if (activeSigning.value) {
          const updated = submissions.value.find((s) => s.submission_id === activeSigning.value?.submission_id) || null;
          activeSigning.value = updated;
        }
      }
    } catch (error) {
      console.error("Failed to load submissions:", error);
    } finally {
      isLoading.value = false;
      loadSubmissionsPromise = null;
    }
  })();

  return loadSubmissionsPromise;
}

async function loadTemplates(): Promise<void> {
  // Prevent multiple simultaneous loads
  if (loadTemplatesPromise) {
    return loadTemplatesPromise;
  }

  loadTemplatesPromise = (async () => {
    try {
      // Use search endpoint which properly filters by organization
      // This will return all templates from root and folders in current organization
      const response = await apiGet("/api/v1/templates/search?limit=1000");
      
      if (response && response.data) {
        // Search endpoint returns: { success: true, message: "templates", data: { templates: [...], total: ... } }
        const result = response.data;
        if (result.templates && Array.isArray(result.templates)) {
          templates.value = result.templates;
        } else if (Array.isArray(result)) {
          // Fallback if API returns array directly
          templates.value = result;
        } else {
          templates.value = [];
        }
      } else {
        templates.value = [];
      }
    } catch (error) {
      console.error("Failed to load templates:", error);
      templates.value = [];
    } finally {
      loadTemplatesPromise = null;
    }
  })();

  return loadTemplatesPromise;
}

async function loadFolders(): Promise<void> {
  try {
    const response = await apiGet("/api/v1/templates/folders");
    if (response && response.data) {
      const result = response.data;
      if (Array.isArray(result)) {
        folders.value = result.map((f: any) => ({ id: f.id, name: f.name }));
      } else if (result.folders && Array.isArray(result.folders)) {
        folders.value = result.folders.map((f: any) => ({ id: f.id, name: f.name }));
      } else {
        folders.value = [];
      }
    }
  } catch (error) {
    console.error("Failed to load folders:", error);
    folders.value = [];
  }
}

function getTemplateDisplayName(template: Template): string {
  if (!template.folder_id) {
    return template.name;
  }
  
  const folder = folders.value.find((f) => f.id === template.folder_id);
  if (folder) {
    return `${folder.name} / ${template.name}`;
  }
  
  return template.name;
}

function openCreateModal(): void {
  expectedSubmittersCount.value = 0;
  createSubmitters.value = [];
  createdLinks.value = [];
  showModal.value = true;
}

async function onTemplateChange(templateID: string): Promise<void> {
  if (!templateID) {
    expectedSubmittersCount.value = 0;
    createSubmitters.value = [];
    createdLinks.value = [];
    return;
  }

  try {
    const res: any = await apiGet(`/api/v1/templates/${templateID}`);
    const tpl = res.data;
    const count = Array.isArray(tpl?.submitters) ? tpl.submitters.length : 0;
    expectedSubmittersCount.value = count;
    createSubmitters.value = Array.from({ length: count }).map((_, idx) => ({
      name: tpl?.submitters?.[idx]?.name || "",
      email: "",
      phone: ""
    }));
    createdLinks.value = [];
  } catch (e) {
    console.error("Failed to load template:", e);
    expectedSubmittersCount.value = 0;
    createSubmitters.value = [];
  }
}

function onCreateSubmitterOrder(ordered: Array<{ id: string; name: string }>): void {
  const indices = ordered.map((s) => parseInt(s.id, 10));
  if (indices.some((i) => Number.isNaN(i))) return;
  const current = [...createSubmitters.value];
  createSubmitters.value = indices.map((i) => current[i]).filter(Boolean);
}

function onSignerDragStart(event: DragEvent, index: number): void {
  draggedSignerIndex.value = index;
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", String(index));
  }
}

function onSignerDragEnd(): void {
  draggedSignerIndex.value = null;
  insertBeforeIndex.value = null;
}

function onSignerDragOver(event: DragEvent, cardIndex: number): void {
  event.preventDefault();
  if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
  if (draggedSignerIndex.value === null) return;
  const el = event.currentTarget as HTMLElement;
  const rect = el.getBoundingClientRect();
  const mid = rect.top + rect.height / 2;
  insertBeforeIndex.value = event.clientY < mid ? cardIndex : cardIndex + 1;
}

function onSignerDrop(event: DragEvent): void {
  event.preventDefault();
  event.stopPropagation();
  if (draggedSignerIndex.value === null) return;
  const pos = insertBeforeIndex.value;
  if (pos === null) {
    draggedSignerIndex.value = null;
    insertBeforeIndex.value = null;
    return;
  }
  const arr = [...createSubmitters.value];
  const [item] = arr.splice(draggedSignerIndex.value, 1);
  const insertAt = draggedSignerIndex.value < pos ? pos - 1 : pos;
  arr.splice(insertAt, 0, item);
  createSubmitters.value = arr;
  draggedSignerIndex.value = null;
  insertBeforeIndex.value = null;
}

async function handleSubmit(formData: any): Promise<void> {
  try {
    if (!formData.template_id) {
      alert(String(t("submissions.selectTemplateRequired")));
      return;
    }

    if (expectedSubmittersCount.value <= 0 || createSubmitters.value.length !== expectedSubmittersCount.value) {
      alert(String(t("submissions.invalidTemplateSubmitters")));
      return;
    }

    const response = await fetchWithAuth("/api/v1/signing-links", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        template_id: formData.template_id,
        signing_mode: createSigningMode.value,
        submitters: createSubmitters.value.map((s, idx) => ({
          name: s.name || `${t("signing.signer")} ${idx + 1}`,
          email: s.email || "",
          phone: s.phone || ""
        }))
      })
    });

    if (response.ok) {
      const data = await response.json();
      const payload = data.data || data;
      const links = (payload.links || []) as Array<{ submitter_id: string; slug: string; direct_url: string }>;
      createdLinks.value = links.map((l) => ({
        ...l,
        full_url: `${window.location.origin}${l.direct_url}`
      }));

      await loadSubmissions();
      // Close modal after successful creation (requested UX).
      showModal.value = false;
    } else {
      alert(String(t("submissions.saveFailed")));
    }
  } catch (error) {
    console.error("Failed to save submission:", error);
    alert(String(t("submissions.saveFailed")));
  }
}

async function copyText(text: string): Promise<void> {
  try {
    await navigator.clipboard.writeText(text);
  } catch {
    const input = document.createElement("input");
    input.value = text;
    document.body.appendChild(input);
    input.select();
    document.execCommand("copy");
    document.body.removeChild(input);
  }
}

async function copyAllCreatedLinks(): Promise<void> {
  if (!createdLinks.value.length) {
    return;
  }
  await copyText(createdLinks.value.map((l) => l.full_url).join("\n"));
}

async function copyLinks(signing: Signing): Promise<void> {
  const urls = (signing.links || []).map((l) => `${window.location.origin}${l.direct_url}`);
  if (!urls.length) {
    return;
  }
  await copyText(urls.join("\n"));
}

function openLinksModal(signing: Signing): void {
  activeSigning.value = signing;
  linksModalOpen.value = true;
}

function handlePageChange(page: number): void {
  console.log("Page changed:", page);
  // Implement server-side pagination if needed
}

function openUrl(url: string): void {
  window.open(url, "_blank", "noopener,noreferrer");
}

function fullSignerUrl(slug: string): string {
  return `${window.location.origin}/s/${slug}`;
}

function statusLabel(status: unknown): string {
  const key = getI18nStatusKey(status);
  return te(key) ? t(key) : String(status || "");
}

function openStatusHistory(signing: Signing): void {
  const id = String(signing?.submission_id || "");
  if (!id) {
    return;
  }
  router.push(`/submissions/${encodeURIComponent(id)}/status`);
}

async function openCompletedDocument(signing: Signing): Promise<void> {
  const id = String(signing?.submission_id || "");
  if (!id) {
    return;
  }
  try {
    const res = await fetchWithAuth(`/api/v1/signing-links/${encodeURIComponent(id)}/document`, { method: "GET" });
    if (res.status === 409) {
      alert("Document is not completed yet.");
      return;
    }
    if (res.status === 404 || res.status === 403) {
      alert("Only the owner can view the completed document.");
      return;
    }
    if (!res.ok) {
      alert("Failed to load completed document.");
      return;
    }
    const buf = await res.arrayBuffer();
    openBlobInNewTab(new Blob([buf], { type: "application/pdf" }));
  } catch (e) {
    console.error("Failed to open completed document:", e);
    alert("Failed to load completed document.");
  }
}
</script>

<style scoped>
.submissions-page {
  @apply min-h-full bg-[--color-base-200];
}
</style>
