<template>
  <div class="min-h-full">
    <div class="mb-6 flex flex-wrap items-start justify-between gap-4">
      <div class="min-w-0">
        <div class="truncate text-2xl font-bold">
          {{ detail?.template_name || "—" }}
        </div>
        <div v-if="detail?.created_at" class="mt-1 text-sm text-gray-600">
          {{ t("submissionStatus.createdAt", { date: formatDate(detail.created_at) }) }}
        </div>
      </div>

      <div class="flex flex-wrap items-center justify-end gap-3">
        <Badge v-if="detail" size="sm" :variant="getBadgeVariantForSubmissionStatus(detail.status)">
          {{ statusLabel(detail.status) }}
        </Badge>
        <div v-if="detail" class="text-sm text-gray-600">
          {{ t("submissionStatus.progressCompleted", { completed: detail.completed_count, total: detail.total_count }) }}
        </div>
        <Button
          v-if="detail && String(detail.status) === 'completed'"
          variant="ghost"
          size="sm"
          @click="openCompletedDocument(detail.submission_id)"
        >
          {{ t("common.download") }}
        </Button>
      </div>
    </div>

    <Card>

      <div v-if="loading" class="py-10 text-center text-gray-600">{{ t("common.loading") }}</div>
      <div v-else-if="error" class="py-10 text-center text-red-600">{{ error }}</div>

      <div v-else-if="detail" class="grid grid-cols-1 gap-6 lg:grid-cols-5">
        <!-- Timeline -->
        <div class="min-w-0 lg:col-span-2">
          <div class="mb-3 flex items-center justify-between">
            <div class="text-sm font-semibold text-gray-700">{{ t("submissionStatus.timeline.title") }}</div>
          </div>

          <div class="pt-4">
            <div v-if="timeline.length === 0" class="py-6 text-center text-gray-500">
              {{ t("submissionStatus.timeline.empty") }}
            </div>

            <div v-else class="space-y-4">
              <div v-for="e in timeline" :key="e.key" class="flex items-start gap-3">
                <div class="mt-0.5 flex h-9 w-9 flex-none items-center justify-center rounded-full bg-[var(--color-base-200)]">
                  <span class="text-base">{{ e.icon }}</span>
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
                    <div class="font-medium">{{ e.title }}</div>
                    <span v-if="e.signer" class="rounded-full bg-[var(--color-base-200)] px-2 py-0.5 text-xs text-gray-700">
                      {{ e.signer }}
                    </span>
                  </div>
                  <div class="text-sm text-gray-500">{{ formatDate(e.at) }}</div>
                  <div v-if="e.reason" class="text-sm text-gray-500 font-mono">{{ t('common.reasonLabel') }} {{ e.reason }}</div>
                  <div v-if="e.ip" class="text-sm text-gray-500 font-mono">IP address: {{ e.ip }}</div>
                  <div v-if="e.location" class="text-sm text-gray-500">Location: {{ e.location }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Signers -->
        <div class="min-w-0 lg:col-span-3">
          <div class="mb-3 text-sm font-semibold text-gray-700">{{ t("submissionStatus.signers.title") }}</div>

          <ResourceTable
            :data="signers"
            :columns="signerColumns"
            :show-pagination="false"
            :has-actions="false"
            :searchable="false"
            :show-filters="false"
            :empty-message="t('submissionStatus.signers.empty')"
          >
            <template #cell-signer="{ item }">
              <div class="font-medium">{{ (item as any).display_name }}</div>
              <div v-if="(item as any).email" class="text-xs text-gray-500">{{ (item as any).email }}</div>
              <div v-if="(item as any).phone" class="text-xs text-gray-500">{{ (item as any).phone }}</div>
            </template>

            <template #cell-status="{ value }">
              <Badge size="sm" :variant="getBadgeVariantForSubmitterStatus(value)">
                {{ statusLabel(value) }}
              </Badge>
            </template>

            <template #cell-opened_at="{ value }">
              <span class="text-gray-600">{{ formatDate(value) }}</span>
            </template>
            <template #cell-completed_at="{ value }">
              <span class="text-gray-600">{{ formatDate(value) }}</span>
            </template>
            <template #cell-declined_at="{ value }">
              <span class="text-gray-600">{{ formatDate(value) }}</span>
            </template>
          </ResourceTable>
        </div>
      </div>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import ResourceTable from "@/components/common/ResourceTable.vue";
import Card from "@/components/ui/Card.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import { fetchWithAuth } from "@/utils/auth";
import { getBadgeVariantForSubmissionStatus, getBadgeVariantForSubmitterStatus, getI18nStatusKey } from "@/utils/status";
import { useI18n } from "vue-i18n";
import { openBlobInNewTab } from "@/utils/file";

type SigningLinkDetail = {
  submission_id: string;
  template_id: string;
  template_name: string;
  created_at: string;
  created_ip?: string;
  status: string;
  completed_count: number;
  total_count: number;
  submitters: Array<Record<string, any>>;
  decline_events?: Array<{ at: string; submitter_id: string; submitter_name: string; ip?: string; reason?: string }>;
  opened_events?: Array<{ at: string; submitter_id: string; submitter_name: string; ip?: string }>;
  completed_events?: Array<{ at: string; submitter_id: string; submitter_name: string; ip?: string }>;
};

type TimelineItem = {
  key: string;
  at: string;
  title: string;
  icon: string;
  signer?: string;
  ip?: string;
  location?: string;
  reason?: string;
};

const route = useRoute();
const router = useRouter();
const { t, te } = useI18n();

const loading = ref(false);
const error = ref<string | null>(null);
const detail = ref<SigningLinkDetail | null>(null);

const submissionID = computed(() => String(route.params.submission_id || ""));

onMounted(async () => {
  await load();
});

async function load(): Promise<void> {
  if (!submissionID.value) {
    error.value = t("submissionStatus.errors.missingSubmissionId");
    return;
  }
  loading.value = true;
  error.value = null;
  try {
    const res = await fetchWithAuth(`/api/v1/signing-links/${encodeURIComponent(submissionID.value)}`);
    if (!res.ok) {
      error.value = t("submissionStatus.errors.failedToLoad");
      detail.value = null;
      return;
    }
    const json = await res.json();
    detail.value = (json?.data || json) as SigningLinkDetail;
  } catch (e) {
    error.value = e instanceof Error ? e.message : t("submissionStatus.errors.failedToLoad");
    detail.value = null;
  } finally {
    loading.value = false;
  }
}

const timeline = computed<TimelineItem[]>(() => {
  const d = detail.value;
  if (!d) {
    return [];
  }

  const items: TimelineItem[] = [];

  if (d.created_at) {
    items.push({
      key: `created:${d.submission_id}`,
      at: d.created_at,
      title: t("submissionStatus.timeline.events.created"),
      icon: "📄",
      ip: d.created_ip
    });
  }

  for (const e of d.opened_events || []) {
    items.push({
      key: `opened:${String(e.submitter_id)}:${String(e.at)}`,
      at: String(e.at),
      title: t("submissionStatus.timeline.events.opened"),
      icon: "👀",
      signer: e.submitter_name || t("submissionStatus.signerFallback"),
      ip: e.ip
    });
  }

  for (const e of d.completed_events || []) {
    items.push({
      key: `completed:${String(e.submitter_id)}:${String(e.at)}`,
      at: String(e.at),
      title: t("submissionStatus.timeline.events.completed"),
      icon: "✅",
      signer: e.submitter_name || t("submissionStatus.signerFallback"),
      ip: e.ip
    });
  }

  for (const e of d.decline_events || []) {
    items.push({
      key: `declined:${String(e.submitter_id)}:${String(e.at)}`,
      at: String(e.at),
      title: t("submissionStatus.timeline.events.declined"),
      icon: "✕",
      signer: e.submitter_name || t("submissionStatus.signerFallback"),
      ip: e.ip,
      reason: e.reason || undefined
    });
  }

  const hasOpened = (d.opened_events?.length ?? 0) > 0;
  const hasCompleted = (d.completed_events?.length ?? 0) > 0;
  const hasDeclined = (d.decline_events?.length ?? 0) > 0;
  if (!hasOpened || !hasCompleted || !hasDeclined) {
    for (const s of d.submitters || []) {
      const signerName = String(s.name || t("submissionStatus.signerFallback"));
      if (!hasOpened && s.opened_at) {
        items.push({
          key: `opened:${String(s.id)}:${String(s.opened_at)}`,
          at: String(s.opened_at),
          title: t("submissionStatus.timeline.events.opened"),
          icon: "👀",
          signer: signerName,
          ip: s.opened_ip,
          location: s.opened_location || null
        });
      }
      if (!hasCompleted && s.completed_at) {
        items.push({
          key: `completed:${String(s.id)}:${String(s.completed_at)}`,
          at: String(s.completed_at),
          title: t("submissionStatus.timeline.events.completed"),
          icon: "✅",
          signer: signerName,
          ip: s.completed_ip,
          location: s.completed_location || null
        });
      }
      if (!hasDeclined && s.declined_at) {
        items.push({
          key: `declined:${String(s.id)}:${String(s.declined_at)}`,
          at: String(s.declined_at),
          title: t("submissionStatus.timeline.events.declined"),
          icon: "✕",
          signer: signerName,
          ip: s.declined_ip,
          location: s.declined_location || null,
          reason: s.decline_reason || undefined
        });
      }
    }
  }

  // Sort reverse-chronologically (newest first), keep stable ordering for identical timestamps.
  return items
    .filter((i) => !!i.at)
    .sort((a, b) => {
      const at = new Date(a.at).getTime();
      const bt = new Date(b.at).getTime();
      if (at === bt) {
        return a.key.localeCompare(b.key);
      }
      return bt - at;
    });
});

function statusLabel(status: unknown): string {
  const key = getI18nStatusKey(status);
  return te(key) ? t(key) : String(status || "");
}

const signerColumns = computed(() => [
  { key: "signer", label: t("submissionStatus.signers.columns.signer"), sortable: false },
  { key: "status", label: t("submissionStatus.signers.columns.status"), sortable: false, headerClass: "w-32" },
  { key: "opened_at", label: t("submissionStatus.signers.columns.opened"), sortable: false, headerClass: "w-36" },
  { key: "completed_at", label: t("submissionStatus.signers.columns.completed"), sortable: false, headerClass: "w-36" },
  { key: "declined_at", label: t("submissionStatus.signers.columns.declined"), sortable: false, headerClass: "w-36" }
]);

const signers = computed(() => {
  const d = detail.value;
  if (!d) {
    return [];
  }
  return (d.submitters || []).map((s) => ({
    id: String(s.id || ""),
    display_name: String(s.name || t("submissionStatus.signerFallback")),
    email: String(s.email || ""),
    phone: String(s.phone || ""),
    status: s.status,
    opened_at: s.opened_at,
    completed_at: s.completed_at,
    declined_at: s.declined_at,
    decline_reason: s.decline_reason,
    slug: s.slug
  }));
});

function formatDate(value: unknown): string {
  const s = String(value || "");
  if (!s) {
    return "—";
  }
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) {
    return s;
  }
  return d.toLocaleString();
}

async function openCompletedDocument(submissionId: string): Promise<void> {
  const id = String(submissionId || "");
  if (!id) {
    return;
  }
  try {
    const res = await fetchWithAuth(`/api/v1/signing-links/${encodeURIComponent(id)}/document`, { method: "GET" });
    if (res.status === 409) {
      alert(t("submissionStatus.errors.notCompletedYet"));
      return;
    }
    if (res.status === 404 || res.status === 403) {
      alert(t("submissionStatus.errors.onlyOwnerCanView"));
      return;
    }
    if (!res.ok) {
      alert(t("submissionStatus.errors.failedToLoadDocument"));
      return;
    }
    const buf = await res.arrayBuffer();
    openBlobInNewTab(new Blob([buf], { type: "application/pdf" }));
  } catch (e) {
    console.error("Failed to open completed document:", e);
    alert(t("submissionStatus.errors.failedToLoadDocument"));
  }
}
</script>

