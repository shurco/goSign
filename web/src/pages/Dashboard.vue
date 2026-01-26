<template>
  <div class="dashboard-page">
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold">{{ $t('dashboard.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ $t('dashboard.description') }}</p>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="mb-6 grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <div class="flex items-center justify-between">
          <div>
            <div class="mb-1 text-sm text-gray-600">{{ $t('dashboard.totalSubmissions') }}</div>
            <div class="mb-1 text-3xl font-bold text-[var(--color-primary)]">{{ stats.total_submissions }}</div>
            <div class="text-sm text-gray-500">{{ stats.pending_submissions }} {{ $t('dashboard.pendingSubmissions').toLowerCase() }}</div>
          </div>
          <SvgIcon name="document" class="h-12 w-12 text-[var(--color-primary)] opacity-20" />
        </div>
      </Card>

      <Card>
        <div class="flex items-center justify-between">
          <div>
            <div class="mb-1 text-sm text-gray-600">{{ $t('dashboard.completedSubmissions') }}</div>
            <div class="mb-1 text-3xl font-bold text-[var(--color-success)]">{{ stats.completed_submissions }}</div>
            <div class="text-sm text-gray-500">{{ completionRate }}% {{ $t('common.complete') }}</div>
          </div>
          <SvgIcon name="check-circle" class="h-12 w-12 text-[var(--color-success)] opacity-20" />
        </div>
      </Card>

      <Card>
        <div class="flex items-center justify-between">
          <div>
            <div class="mb-1 text-sm text-gray-600">{{ $t('dashboard.totalTemplates') }}</div>
            <div class="mb-1 text-3xl font-bold text-[var(--color-primary)]">{{ stats.total_templates }}</div>
            <div class="text-sm text-gray-500">{{ stats.active_templates || 0 }} {{ $t('dashboard.active') }}</div>
          </div>
          <SvgIcon name="templates" class="h-12 w-12 text-[var(--color-primary)] opacity-20" />
        </div>
      </Card>

      <Card>
        <div class="flex items-center justify-between">
          <div>
            <div class="mb-1 text-sm text-gray-600">{{ $t('dashboard.submitters') }}</div>
            <div class="mb-1 text-3xl font-bold text-[var(--color-info)]">{{ stats.total_submitters }}</div>
            <div class="text-sm text-gray-500">{{ $t('dashboard.thisMonth') }}</div>
          </div>
          <SvgIcon name="users" class="h-12 w-12 text-[var(--color-info)] opacity-20" />
        </div>
      </Card>
    </div>

    <!-- Recent Activity -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- Recent Submissions -->
      <Card>
        <template #header>
          <h2 class="text-lg font-semibold">{{ $t('dashboard.recentSubmissions') }}</h2>
        </template>
        <ResourceTable
          :data="recentSubmissions"
          :columns="submissionColumns"
          :show-pagination="false"
          :has-actions="false"
          :searchable="false"
          :show-filters="false"
          :empty-message="$t('submissions.title')"
          @row-click="viewSubmission"
        >
          <template #cell-title="{ value, item }">
            <button
              type="button"
              class="link text-left"
              @click.stop="openCompletedDocument(item as Submission)"
            >
              {{ String(value || "") }}
            </button>
          </template>

          <template #cell-status="{ value, item }">
            <button type="button" class="inline-flex cursor-pointer" @click.stop="openStatusHistory(item as Submission)">
              <Badge size="sm" :variant="getBadgeVariantForSubmissionStatus(value)">
                {{ statusLabel(value) }}
              </Badge>
            </button>
          </template>
        </ResourceTable>
        <template #footer>
          <div class="flex justify-end">
            <Button variant="ghost" size="sm" @click="$router.push('/submissions')">{{ $t('common.viewAll') }} →</Button>
          </div>
        </template>
      </Card>

      <!-- Activity Timeline -->
      <Card>
        <template #header>
          <h2 class="text-lg font-semibold">{{ $t('dashboard.recentActivity') }}</h2>
        </template>
        <div class="pt-4">
          <div v-if="recentEvents.length === 0" class="py-6 text-center text-gray-500">{{ $t('dashboard.noActivity') }}</div>
          <div v-else class="space-y-4">
            <div v-for="event in recentEvents" :key="event.id" class="flex items-start gap-3">
              <div class="mt-0.5 flex h-9 w-9 flex-none items-center justify-center rounded-full bg-[var(--color-base-200)]">
                <span class="text-base">{{ getEventIcon(event.type) }}</span>
              </div>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
                  <div class="font-medium">{{ event.message }}</div>
                  <span v-if="event.document_name" class="rounded-full bg-[var(--color-base-200)] px-2 py-0.5 text-xs text-gray-700">
                    {{ event.document_name }}
                  </span>
                </div>
                <div class="text-sm text-gray-500">{{ formatEventDate(event.created_at) }}</div>
                <div v-if="event.reason" class="text-sm text-gray-500 font-mono">{{ $t('common.reasonLabel') }} {{ event.reason }}</div>
                <div v-if="event.ip" class="text-sm text-gray-500 font-mono">IP address: {{ event.ip }}</div>
                <div v-if="event.location" class="text-sm text-gray-500">Location: {{ event.location }}</div>
              </div>
            </div>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import ResourceTable from "@/components/common/ResourceTable.vue";
import Card from "@/components/ui/Card.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import SvgIcon from "@/components/SvgIcon.vue";
import { fetchWithAuth } from "@/utils/auth";
import { getBadgeVariantForSubmissionStatus, getI18nStatusKey } from "@/utils/status";
import { openBlobInNewTab } from "@/utils/file";

interface Submission {
  id: string;
  title: string;
  status: string;
  created_at: string;
}

interface Event {
  id: string;
  type: string;
  message: string;
  document_name?: string;
  created_at: string;
  ip?: string;
  location?: string;
  reason?: string;
}

const router = useRouter();

const stats = ref<Stats>({
  total_submissions: 0,
  pending_submissions: 0,
  completed_submissions: 0,
  total_templates: 0,
  active_templates: 0,
  total_submitters: 0
});

const recentSubmissions = ref<Submission[]>([]);
const recentEvents = ref<Event[]>([]);

const { t, te } = useI18n();

const submissionColumns = computed(() => [
  { key: "title", label: t('submissions.titleField'), sortable: true },
  { key: "status", label: t('submissions.status'), sortable: true },
  {
    key: "created_at",
    label: t('submissions.created'),
    sortable: true,
    formatter: (value: unknown): string => new Date(String(value)).toLocaleDateString()
  }
]);

const completionRate = computed(() => {
  if (stats.value.total_submissions === 0) {
    return 0;
  }
  return Math.round((stats.value.completed_submissions / stats.value.total_submissions) * 100);
});

onMounted(async () => {
  await Promise.all([loadStats(), loadRecentSubmissions(), loadRecentEvents()]);
});

async function loadStats(): Promise<void> {
  try {
    // Token check is redundant - fetchWithAuth handles authentication
    const response = await fetchWithAuth("/api/v1/stats");
    if (response.ok) {
      const data = await response.json();
      // Stats API returns: { message: "Stats retrieved", data: {...} }
      if (data.data) {
        stats.value = data.data;
      } else {
        stats.value = data;
      }
    } else if (response.status === 401) {
      // Redirect to login is handled by fetchWithAuth
      return;
    }
  } catch (error) {
    // Only log error if we're still on dashboard page (not redirected)
    if (window.location.pathname.includes("/dashboard")) {
      console.error("Failed to load stats:", error);
    }
  }
}

async function loadRecentSubmissions(): Promise<void> {
  try {
    // Token check is redundant - fetchWithAuth handles authentication
    const response = await fetchWithAuth("/api/v1/signing-links?page=1&page_size=5");
    if (response.ok) {
      const data = await response.json();
      const payload = data.data || data;
      const items = (payload.items || []) as any[];
      recentSubmissions.value = items.map((s) => ({
        id: s.submission_id,
        title: s.template_name,
        status: s.status,
        created_at: s.created_at
      }));
    } else if (response.status === 401) {
      // Redirect to login is handled by fetchWithAuth
      return;
    }
  } catch (error) {
    // Only log error if we're still on dashboard page (not redirected)
    if (window.location.pathname.includes("/dashboard")) {
      console.error("Failed to load submissions:", error);
    }
  }
}

async function loadRecentEvents(): Promise<void> {
  try {
    // Token check is redundant - fetchWithAuth handles authentication
    const response = await fetchWithAuth("/api/v1/events?limit=6&sort=created_at:desc");
    if (response.ok) {
      const data = await response.json();
      // API returns: { message: "...", data: { items: [], total: 0, ... } }
      if (data.data && data.data.items) {
        recentEvents.value = data.data.items || [];
      } else if (Array.isArray(data.data)) {
        recentEvents.value = data.data;
      } else {
        recentEvents.value = [];
      }
      // Debug: log events data to check location
      if (recentEvents.value.length > 0) {
        console.log("Events data:", recentEvents.value);
      }
    } else if (response.status === 401) {
      // Redirect to login is handled by fetchWithAuth
      return;
    }
  } catch (error) {
    // Only log error if we're still on dashboard page (not redirected)
    if (window.location.pathname.includes("/dashboard")) {
      console.error("Failed to load events:", error);
    }
  }
}

function statusLabel(status: unknown): string {
  const key = getI18nStatusKey(status);
  return te(key) ? t(key) : String(status || "");
}

function getEventIcon(type: string): string {
  const icons: Record<string, string> = {
    submission_created: "📄",
    submission_sent: "✉️",
    submission_completed: "✅",
    submitter_opened: "👀",
    submitter_completed: "✓",
    submitter_declined: "✕",
    template_created: "📝"
  };
  return icons[type] || "•";
}

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 1) {
    return t('time.justNow');
  }
  if (minutes < 60) {
    return t('time.minutesAgo', { minutes });
  }
  if (hours < 24) {
    return t('time.hoursAgo', { hours });
  }
  if (days < 7) {
    return t('time.daysAgo', { days });
  }

  return date.toLocaleDateString();
}

function formatEventDate(value: unknown): string {
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

function viewSubmission(submission: Submission): void {
  openStatusHistory(submission);
}

function openStatusHistory(submission: Submission): void {
  const id = String(submission?.id || "");
  if (!id) {
    return;
  }
  router.push(`/submissions/${encodeURIComponent(id)}/status`);
}

async function openCompletedDocument(submission: Submission): Promise<void> {
  const id = String(submission?.id || "");
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
.dashboard-page {
  @apply min-h-full;
}
</style>
