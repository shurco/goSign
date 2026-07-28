<script lang="ts">
  import { onMount } from "svelte";
  import { apiGet } from "@/services/api";
  import { goto } from "$app/navigation";
  import { t } from "@/i18n/index.svelte";
  import ResourceTable from "@/components/common/ResourceTable.svelte";
  import Card from "@/components/ui/Card.svelte";
  import Badge from "@/components/ui/Badge.svelte";
  import Button from "@/components/ui/Button.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { getBadgeVariantForSubmissionStatus, statusLabel } from "@/utils/status";
  import { openCompletedDocument } from "@/utils/file";

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

  interface Stats {
    total_submissions: number;
    pending_submissions: number;
    completed_submissions: number;
    total_templates: number;
    active_templates: number;
    total_submitters: number;
  }

  let stats = $state<Stats>({
    total_submissions: 0,
    pending_submissions: 0,
    completed_submissions: 0,
    total_templates: 0,
    active_templates: 0,
    total_submitters: 0
  });

  let recentSubmissions = $state<Submission[]>([]);
  let recentEvents = $state<Event[]>([]);

  const submissionColumns = $derived([
    { key: "title", label: t("submissions.titleField"), sortable: true },
    { key: "status", label: t("submissions.status"), sortable: true },
    {
      key: "created_at",
      label: t("submissions.created"),
      sortable: true,
      formatter: (value: unknown): string => new Date(String(value)).toLocaleDateString()
    }
  ]);

  const completionRate = $derived(
    stats.total_submissions === 0 ? 0 : Math.round((stats.completed_submissions / stats.total_submissions) * 100)
  );

  onMount(async () => {
    await Promise.all([loadStats(), loadRecentSubmissions(), loadRecentEvents()]);
  });

  async function loadStats(): Promise<void> {
    try {
      // Stats API returns: { message: "Stats retrieved", data: {...} }
      const data = await apiGet("/stats");
      stats = data.data || data;
    } catch (error) {
      // Only log error if we're still on dashboard page (not redirected)
      if (window.location.pathname.includes("/dashboard")) {
        console.error("Failed to load stats:", error);
      }
    }
  }

  async function loadRecentSubmissions(): Promise<void> {
    try {
      const data = await apiGet("/signing-links?page=1&page_size=5");
      const payload = data.data || data;
      const items = (payload.items || []) as any[];
      recentSubmissions = items.map((s) => ({
        id: s.submission_id,
        title: s.template_name,
        status: s.status,
        created_at: s.created_at
      }));
    } catch (error) {
      // Only log error if we're still on dashboard page (not redirected)
      if (window.location.pathname.includes("/dashboard")) {
        console.error("Failed to load submissions:", error);
      }
    }
  }

  async function loadRecentEvents(): Promise<void> {
    try {
      // API returns: { message: "...", data: { items: [], total: 0, ... } }
      const data = await apiGet("/events?limit=6&sort=created_at:desc");
      if (data.data && data.data.items) {
        recentEvents = data.data.items || [];
      } else if (Array.isArray(data.data)) {
        recentEvents = data.data;
      } else {
        recentEvents = [];
      }
    } catch (error) {
      // Only log error if we're still on dashboard page (not redirected)
      if (window.location.pathname.includes("/dashboard")) {
        console.error("Failed to load events:", error);
      }
    }
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

  function openStatusHistory(submission: Submission): void {
    const id = String(submission?.id || "");
    if (!id) {
      return;
    }
    goto(`/submissions/${encodeURIComponent(id)}/status`);
  }
</script>

<div class="dashboard-page">
  <!-- Header -->
  <div class="page-header">
    <div>
      <h1>{t("dashboard.title")}</h1>
      <p class="page-subtitle">{t("dashboard.description")}</p>
    </div>
  </div>

  <!-- Stats Grid (WS2 metric widgets) -->
  <div class="widget-grid">
    <div class="widget">
      <div class="widget-body">
        <h4 class="widget-title text-overline">{t("dashboard.totalSubmissions")}</h4>
        <div class="widget-value tone-accent">{stats.total_submissions}</div>
        <div class="widget-subtitle">
          {stats.pending_submissions}
          {t("dashboard.pendingSubmissions").toLowerCase()}
        </div>
      </div>
      <SvgIcon name="document" class="widget-icon" />
    </div>

    <div class="widget">
      <div class="widget-body">
        <h4 class="widget-title text-overline">{t("dashboard.completedSubmissions")}</h4>
        <div class="widget-value tone-success">{stats.completed_submissions}</div>
        <div class="widget-subtitle">{completionRate}% {t("common.complete")}</div>
      </div>
      <SvgIcon name="check-circle" class="widget-icon" />
    </div>

    <div class="widget">
      <div class="widget-body">
        <h4 class="widget-title text-overline">{t("dashboard.totalTemplates")}</h4>
        <div class="widget-value tone-accent">{stats.total_templates}</div>
        <div class="widget-subtitle">{stats.active_templates || 0} {t("dashboard.active")}</div>
      </div>
      <SvgIcon name="templates" class="widget-icon" />
    </div>

    <div class="widget">
      <div class="widget-body">
        <h4 class="widget-title text-overline">{t("dashboard.submitters")}</h4>
        <div class="widget-value tone-neutral">{stats.total_submitters}</div>
        <div class="widget-subtitle">{t("dashboard.thisMonth")}</div>
      </div>
      <SvgIcon name="users" class="widget-icon" />
    </div>
  </div>

  <!-- Recent Activity -->
  <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
    <!-- Recent Submissions -->
    {#snippet cellTitle(item: unknown, value: string)}
      <button
        type="button"
        class="link text-left"
        onclick={(e) => {
          e.stopPropagation();
          openCompletedDocument((item as Submission).id);
        }}
      >
        {String(value || "")}
      </button>
    {/snippet}

    {#snippet cellStatus(item: unknown, value: string)}
      <button
        type="button"
        class="inline-flex cursor-pointer"
        onclick={(e) => {
          e.stopPropagation();
          openStatusHistory(item as Submission);
        }}
      >
        <Badge size="sm" variant={getBadgeVariantForSubmissionStatus(value)}>
          {statusLabel(value)}
        </Badge>
      </button>
    {/snippet}

    <Card>
      {#snippet header()}
        <h2 class="text-lg font-semibold">{t("dashboard.recentSubmissions")}</h2>
      {/snippet}

      <ResourceTable
        data={recentSubmissions}
        columns={submissionColumns}
        showPagination={false}
        hasActions={false}
        searchable={false}
        showFilters={false}
        emptyMessage={t("submissions.title")}
        cellSnippets={{ title: cellTitle, status: cellStatus }}
      />

      {#snippet footer()}
        <div class="flex justify-end">
          <Button variant="ghost" size="sm" onclick={() => goto("/submissions")}>
            {t("common.viewAll")} →
          </Button>
        </div>
      {/snippet}
    </Card>

    <!-- Activity Timeline -->
    <Card>
      {#snippet header()}
        <h2 class="text-lg font-semibold">{t("dashboard.recentActivity")}</h2>
      {/snippet}

      <div class="pt-4">
        {#if recentEvents.length === 0}
          <div class="py-6 text-center text-gray-500">
            {t("dashboard.noActivity")}
          </div>
        {:else}
          <div class="space-y-4">
            {#each recentEvents as event (event.id)}
              <div class="flex items-start gap-3">
                <div
                  class="mt-0.5 flex h-9 w-9 flex-none items-center justify-center rounded-full bg-[var(--color-base-200)]"
                >
                  <span class="text-base">{getEventIcon(event.type)}</span>
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
                    <div class="font-medium">{event.message}</div>
                    {#if event.document_name}
                      <span class="rounded-full bg-[var(--color-base-200)] px-2 py-0.5 text-xs text-gray-700">
                        {event.document_name}
                      </span>
                    {/if}
                  </div>
                  <div class="text-sm text-gray-500">{formatEventDate(event.created_at)}</div>
                  {#if event.reason}
                    <div class="font-mono text-sm text-gray-500">
                      {t("common.reasonLabel")}
                      {event.reason}
                    </div>
                  {/if}
                  {#if event.ip}
                    <div class="font-mono text-sm text-gray-500">IP address: {event.ip}</div>
                  {/if}
                  {#if event.location}
                    <div class="text-sm text-gray-500">Location: {event.location}</div>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </Card>
  </div>
</div>

<style>
  .dashboard-page {
    min-height: 100%;
  }

  /* WS2 metrics (modeled after MetricWidget from pmCRM) */
  .widget-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: var(--space-16);
    margin-bottom: var(--space-24);
  }
  @media (max-width: 1100px) {
    .widget-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
  @media (max-width: 640px) {
    .widget-grid {
      grid-template-columns: 1fr;
    }
  }

  .widget {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-12);
    background: var(--base-cont-top);
    border: 1px solid var(--base-line-tertiary);
    border-radius: var(--radius-12);
    box-shadow: var(--shadow-cont-minor);
    padding: var(--space-20);
  }
  .widget-body {
    display: flex;
    flex-direction: column;
    gap: var(--space-8);
    min-width: 0;
  }
  .widget-title {
    color: var(--base-txt-muted);
    margin: 0;
  }
  .widget-value {
    font-family: var(--font-family-display);
    font-size: var(--font-size-32);
    line-height: var(--line-height-40);
    font-weight: var(--font-weight-bold);
  }
  .widget-subtitle {
    font-size: var(--font-size-12);
    color: var(--base-txt-ghost);
  }
  .widget :global(.widget-icon) {
    width: 40px;
    height: 40px;
    flex-shrink: 0;
    color: var(--base-txt-ghost);
    opacity: 0.6;
  }

  .tone-accent {
    color: var(--color-interblue-700);
  }
  .tone-success {
    color: var(--color-green-600);
  }
  .tone-neutral {
    color: var(--base-txt-tertiary);
  }
</style>
