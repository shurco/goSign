<template>
  <div class="dashboard-page">
    <h1 class="mb-6 text-3xl font-bold">Dashboard</h1>

    <!-- Stats Grid -->
    <div class="mb-6 grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Stats>
        <Stat
          title="Total Submissions"
          :value="stats.total_submissions"
          :description="`${stats.pending_submissions} pending`"
          variant="primary"
        >
          <template #figure>
            <SvgIcon name="document" class="inline-block h-8 w-8 text-[var(--color-primary)]" />
          </template>
        </Stat>
      </Stats>

      <Stats>
        <Stat
          title="Completed"
          :value="stats.completed_submissions"
          :description="`${completionRate}% completion rate`"
          variant="success"
        >
          <template #figure>
            <SvgIcon name="check-circle" class="inline-block h-8 w-8 text-[var(--color-success)]" />
          </template>
        </Stat>
      </Stats>

      <Stats>
        <Stat
          title="Templates"
          :value="stats.total_templates"
          :description="`${stats.active_templates} active`"
          variant="primary"
        >
          <template #figure>
            <SvgIcon name="templates" class="inline-block h-8 w-8 text-[var(--color-primary)]" />
          </template>
        </Stat>
      </Stats>

      <Stats>
        <Stat title="Submitters" :value="stats.total_submitters" description="This month" variant="info">
          <template #figure>
            <SvgIcon name="users" class="inline-block h-8 w-8 text-[var(--color-info)]" />
          </template>
        </Stat>
      </Stats>
    </div>

    <!-- Recent Activity -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- Recent Submissions -->
      <Card>
        <template #header>
          <h2 class="text-lg font-semibold">Recent Submissions</h2>
        </template>
        <ResourceTable
          :data="recentSubmissions"
          :columns="submissionColumns"
          :show-pagination="false"
          :has-actions="false"
          :searchable="false"
          :show-filters="false"
          empty-message="No submissions yet"
          @row-click="viewSubmission"
        >
          <template #cell-status="{ value }">
            <Badge :variant="getStatusBadgeVariant(value)">
              {{ value }}
            </Badge>
          </template>
        </ResourceTable>
        <template #footer>
          <div class="flex justify-end">
            <Button variant="ghost" size="sm" @click="$router.push('/submissions')">View All →</Button>
          </div>
        </template>
      </Card>

      <!-- Activity Timeline -->
      <Card>
        <template #header>
          <h2 class="text-lg font-semibold">Recent Activity</h2>
        </template>
        <div class="space-y-4">
          <div v-for="event in recentEvents" :key="event.id" class="flex gap-4">
            <div class="flex-shrink-0">
              <div class="flex h-10 w-10 items-center justify-center rounded-full bg-[var(--color-base-200)]">
                <span class="text-lg">{{ getEventIcon(event.type) }}</span>
              </div>
            </div>
            <div class="flex-1">
              <p class="font-medium">{{ event.message }}</p>
              <p class="text-sm text-gray-500">{{ formatDate(event.created_at) }}</p>
            </div>
          </div>
          <div v-if="recentEvents.length === 0" class="py-8 text-center text-gray-500">No recent activity</div>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import ResourceTable from "@/components/common/ResourceTable.vue";
import Stats from "@/components/ui/Stats.vue";
import Stat from "@/components/ui/Stat.vue";
import Card from "@/components/ui/Card.vue";
import Badge from "@/components/ui/Badge.vue";
import Button from "@/components/ui/Button.vue";
import SvgIcon from "@/components/SvgIcon.vue";
import { fetchWithAuth } from "@/utils/auth";

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
  created_at: string;
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

const submissionColumns = [
  { key: "title", label: "Title", sortable: true },
  { key: "status", label: "Status", sortable: true },
  {
    key: "created_at",
    label: "Created",
    sortable: true,
    formatter: (value: unknown): string => new Date(String(value)).toLocaleDateString()
  }
];

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
    const response = await fetchWithAuth("/api/v1/submissions?limit=5&sort=created_at:desc");
    if (response.ok) {
      const data = await response.json();
      // API returns: { message: "...", data: { items: [], total: 0, ... } }
      if (data.data && data.data.items) {
        recentSubmissions.value = data.data.items || [];
      } else if (Array.isArray(data.data)) {
        recentSubmissions.value = data.data;
      } else {
        recentSubmissions.value = [];
      }
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
    const response = await fetchWithAuth("/api/v1/events?limit=10&sort=created_at:desc");
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

function getStatusBadgeVariant(status: string): "ghost" | "primary" | "success" | "warning" | "error" | "info" {
  const variants: Record<string, "ghost" | "primary" | "success" | "warning" | "error" | "info"> = {
    draft: "ghost",
    pending: "warning",
    in_progress: "info",
    completed: "success",
    expired: "error",
    cancelled: "ghost"
  };
  return variants[status] || "ghost";
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
    return "Just now";
  }
  if (minutes < 60) {
    return `${minutes}m ago`;
  }
  if (hours < 24) {
    return `${hours}h ago`;
  }
  if (days < 7) {
    return `${days}d ago`;
  }

  return date.toLocaleDateString();
}

function viewSubmission(submission: Submission): void {
  router.push(`/submissions/${submission.id}`);
}
</script>

<style scoped>
.dashboard-page {
  @apply min-h-full;
}
</style>
