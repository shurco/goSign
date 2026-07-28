<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { t, te } from "@/i18n/index.svelte";
  import ResourceTable from "@/components/common/ResourceTable.svelte";
  import Card from "@/components/ui/Card.svelte";
  import Badge from "@/components/ui/Badge.svelte";
  import Button from "@/components/ui/Button.svelte";
  import { fetchWithAuth } from "@/utils/auth";
  import {
    getBadgeVariantForSubmissionStatus,
    getBadgeVariantForSubmitterStatus,
    getI18nStatusKey
  } from "@/utils/status";
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

  let loading = $state(false);
  let error = $state<string | null>(null);
  let detail = $state<SigningLinkDetail | null>(null);

  const submissionID = $derived(String(page.params.submission_id || ""));

  onMount(async () => {
    await load();
  });

  async function load(): Promise<void> {
    if (!submissionID) {
      error = t("submissionStatus.errors.missingSubmissionId");
      return;
    }
    loading = true;
    error = null;
    try {
      const res = await fetchWithAuth(`/api/v1/signing-links/${encodeURIComponent(submissionID)}`);
      if (!res.ok) {
        error = t("submissionStatus.errors.failedToLoad");
        detail = null;
        return;
      }
      const json = await res.json();
      detail = (json?.data || json) as SigningLinkDetail;
    } catch (e) {
      error = e instanceof Error ? e.message : t("submissionStatus.errors.failedToLoad");
      detail = null;
    } finally {
      loading = false;
    }
  }

  const timeline = $derived.by((): TimelineItem[] => {
    const d = detail;
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
            location: s.opened_location || undefined
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
            location: s.completed_location || undefined
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
            location: s.declined_location || undefined,
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

  const signerColumns = $derived([
    { key: "signer", label: t("submissionStatus.signers.columns.signer"), sortable: false },
    { key: "status", label: t("submissionStatus.signers.columns.status"), sortable: false, headerClass: "w-32" },
    { key: "opened_at", label: t("submissionStatus.signers.columns.opened"), sortable: false, headerClass: "w-36" },
    {
      key: "completed_at",
      label: t("submissionStatus.signers.columns.completed"),
      sortable: false,
      headerClass: "w-36"
    },
    { key: "declined_at", label: t("submissionStatus.signers.columns.declined"), sortable: false, headerClass: "w-36" }
  ]);

  const signers = $derived.by(() => {
    const d = detail;
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

{#snippet cellSigner(item: unknown)}
  <div class="font-medium">{(item as any).display_name}</div>
  {#if (item as any).email}
    <div class="text-xs text-gray-500">{(item as any).email}</div>
  {/if}
  {#if (item as any).phone}
    <div class="text-xs text-gray-500">{(item as any).phone}</div>
  {/if}
{/snippet}

{#snippet cellStatus(_item: unknown, value: string)}
  <Badge size="sm" variant={getBadgeVariantForSubmitterStatus(value)}>
    {statusLabel(value)}
  </Badge>
{/snippet}

{#snippet cellOpenedAt(_item: unknown, value: string)}
  <span class="text-gray-600">{formatDate(value)}</span>
{/snippet}

{#snippet cellCompletedAt(_item: unknown, value: string)}
  <span class="text-gray-600">{formatDate(value)}</span>
{/snippet}

{#snippet cellDeclinedAt(_item: unknown, value: string)}
  <span class="text-gray-600">{formatDate(value)}</span>
{/snippet}

<div class="min-h-full">
  <div class="mb-6 flex flex-wrap items-start justify-between gap-4">
    <div class="min-w-0">
      <div class="truncate text-2xl font-bold">
        {detail?.template_name || "—"}
      </div>
      {#if detail?.created_at}
        <div class="mt-1 text-sm text-gray-600">
          {t("submissionStatus.createdAt", { date: formatDate(detail.created_at) })}
        </div>
      {/if}
    </div>

    <div class="flex flex-wrap items-center justify-end gap-3">
      {#if detail}
        <Badge size="sm" variant={getBadgeVariantForSubmissionStatus(detail.status)}>
          {statusLabel(detail.status)}
        </Badge>
        <div class="text-sm text-gray-600">
          {t("submissionStatus.progressCompleted", {
            completed: detail.completed_count,
            total: detail.total_count
          })}
        </div>
        {#if String(detail.status) === "completed"}
          <Button variant="ghost" size="sm" onclick={() => openCompletedDocument(detail!.submission_id)}>
            {t("common.download")}
          </Button>
        {/if}
      {/if}
    </div>
  </div>

  <Card>
    {#if loading}
      <div class="py-10 text-center text-gray-600">{t("common.loading")}</div>
    {:else if error}
      <div class="py-10 text-center text-red-600">{error}</div>
    {:else if detail}
      <div class="grid grid-cols-1 gap-6 lg:grid-cols-5">
        <!-- Timeline -->
        <div class="min-w-0 lg:col-span-2">
          <div class="mb-3 flex items-center justify-between">
            <div class="text-sm font-semibold text-gray-700">{t("submissionStatus.timeline.title")}</div>
          </div>

          <div class="pt-4">
            {#if timeline.length === 0}
              <div class="py-6 text-center text-gray-500">
                {t("submissionStatus.timeline.empty")}
              </div>
            {:else}
              <div class="space-y-4">
                {#each timeline as e (e.key)}
                  <div class="flex items-start gap-3">
                    <div
                      class="mt-0.5 flex h-9 w-9 flex-none items-center justify-center rounded-full bg-[var(--color-base-200)]"
                    >
                      <span class="text-base">{e.icon}</span>
                    </div>
                    <div class="min-w-0 flex-1">
                      <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
                        <div class="font-medium">{e.title}</div>
                        {#if e.signer}
                          <span class="rounded-full bg-[var(--color-base-200)] px-2 py-0.5 text-xs text-gray-700">
                            {e.signer}
                          </span>
                        {/if}
                      </div>
                      <div class="text-sm text-gray-500">{formatDate(e.at)}</div>
                      {#if e.reason}
                        <div class="font-mono text-sm text-gray-500">
                          {t("common.reasonLabel")}
                          {e.reason}
                        </div>
                      {/if}
                      {#if e.ip}
                        <div class="font-mono text-sm text-gray-500">IP address: {e.ip}</div>
                      {/if}
                      {#if e.location}
                        <div class="text-sm text-gray-500">Location: {e.location}</div>
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        </div>

        <!-- Signers -->
        <div class="min-w-0 lg:col-span-3">
          <div class="mb-3 text-sm font-semibold text-gray-700">{t("submissionStatus.signers.title")}</div>

          <ResourceTable
            data={signers}
            columns={signerColumns}
            showPagination={false}
            hasActions={false}
            searchable={false}
            showFilters={false}
            emptyMessage={t("submissionStatus.signers.empty")}
            cellSnippets={{
              signer: cellSigner,
              status: cellStatus,
              opened_at: cellOpenedAt,
              completed_at: cellCompletedAt,
              declined_at: cellDeclinedAt
            }}
          />
        </div>
      </div>
    {/if}
  </Card>
</div>
