<script lang="ts">
  import type { Snippet } from "svelte";
  import { apiDelete, apiGet, apiPost } from "@/services/api";
  import { onMount, getContext } from "svelte";
  import { t } from "@/i18n/index.svelte";
  import ResourceTable from "@/components/common/ResourceTable.svelte";
  import FormModal from "@/components/common/FormModal.svelte";
  import FieldInput from "@/components/common/FieldInput.svelte";
  import FormControl from "@/components/ui/FormControl.svelte";
  import Badge from "@/components/ui/Badge.svelte";
  import Switch from "@/components/ui/Switch.svelte";

  interface Webhook {
    id: string;
    url: string;
    events: string[];
    enabled: boolean;
    secret?: string;
  }

  let webhooks = $state<Webhook[]>([]);
  let showWebhookModal = $state(false);

  // Get trigger ref from parent and register open function
  const webhookModalTrigger = getContext<{ value: (() => void) | null }>("webhookModalTrigger");

  const availableEvents = [
    "submission.created",
    "submission.sent",
    "submission.completed",
    "submitter.opened",
    "submitter.completed",
    "submitter.declined"
  ];

  const webhookColumns = $derived.by(() => [
    { key: "url", label: t("webhooks.url"), sortable: true },
    { key: "events", label: t("webhooks.events") },
    { key: "enabled", label: t("webhooks.status") }
  ]);

  onMount(async () => {
    await loadWebhooks();
  });

  function openWebhookModal(): void {
    showWebhookModal = true;
  }

  // Register open function with parent
  if (webhookModalTrigger) {
    webhookModalTrigger.value = () => {
      openWebhookModal();
    };
  }

  async function loadWebhooks(): Promise<void> {
    try {
      const data = await apiGet("/settings/webhooks");
      if (data.data && data.data.items) {
        webhooks = data.data.items || [];
      } else if (Array.isArray(data.data)) {
        webhooks = data.data;
      } else {
        webhooks = [];
      }
    } catch (error) {
      if (!window.location.pathname.includes("/auth/") && !window.location.pathname.includes("/signin")) {
        console.error("Failed to load webhooks:", error);
      }
    }
  }

  async function saveWebhook(formData: Record<string, unknown>): Promise<void> {
    try {
      await apiPost("/settings/webhooks", formData);
      showWebhookModal = false;
      await loadWebhooks();
    } catch (error) {
      console.error("Failed to save webhook:", error);
      alert(t("webhooks.saveError"));
    }
  }

  // Webhook editing is not implemented yet; the table's edit action is a no-op.
  function editWebhook(_webhook: Webhook): void {}

  async function deleteWebhook(webhook: Webhook): Promise<void> {
    if (!confirm(t("webhooks.deleteConfirm"))) {
      return;
    }

    try {
      await apiDelete(`/settings/webhooks/${webhook.id}`);
      await loadWebhooks();
    } catch (error) {
      console.error("Failed to delete webhook:", error);
      alert(t("webhooks.deleteError"));
    }
  }

  function toggleEvent(formData: Record<string, unknown>, event: string): void {
    if (!Array.isArray(formData.events)) {
      formData.events = [];
    }

    const events = formData.events as string[];
    const index = events.indexOf(event);
    if (index > -1) {
      events.splice(index, 1);
    } else {
      events.push(event);
    }
  }
</script>

{#snippet enabledCell(_item: unknown, value: string)}
  <Badge variant={value ? "success" : "ghost"}>
    {value ? t("webhooks.active") : t("webhooks.inactive")}
  </Badge>
{/snippet}

{#snippet eventsCell(item: unknown, _value: string)}
  {@const events = (item as Webhook).events}
  <div class="flex flex-wrap gap-1">
    {#each events.slice(0, 3) as event (event)}
      <Badge size="sm">{event}</Badge>
    {/each}
    {#if events.length > 3}
      <Badge variant="ghost" size="sm">+{events.length - 3} {t("webhooks.more")}</Badge>
    {/if}
  </div>
{/snippet}

<div class="space-y-4">
  <ResourceTable
    data={webhooks}
    columns={webhookColumns}
    hasActions={true}
    showPagination={false}
    searchable={false}
    showFilters={false}
    emptyMessage={t("webhooks.noWebhooks")}
    cellSnippets={{
      enabled: enabledCell as Snippet<[item: unknown, value: string]>,
      events: eventsCell as Snippet<[item: unknown, value: string]>
    }}
    onEdit={(item) => editWebhook(item as Webhook)}
    onDelete={(item) => deleteWebhook(item as Webhook)}
  />

  <!-- Webhook Modal -->
  <FormModal bind:open={showWebhookModal} title={t("webhooks.configureWebhook")} onSubmit={saveWebhook}>
    {#snippet children(formData)}
      <div class="space-y-4">
        <FieldInput bind:value={formData.url} type="text" placeholder="https://example.com/webhook" required />

        <FormControl label={t("webhooks.eventsToSubscribe")}>
          <div class="space-y-2">
            {#each availableEvents as event (event)}
              <label class="flex cursor-pointer items-center gap-3">
                <Switch
                  size="sm"
                  checked={Array.isArray(formData.events) && formData.events.includes(event)}
                  onchange={() => toggleEvent(formData, event)}
                />
                <span class="text-sm text-gray-700">{event}</span>
              </label>
            {/each}
          </div>
        </FormControl>

        <FieldInput bind:value={formData.secret} type="text" placeholder={t("common.optional")} />
      </div>
    {/snippet}
  </FormModal>
</div>
