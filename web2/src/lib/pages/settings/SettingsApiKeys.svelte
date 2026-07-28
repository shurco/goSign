<script lang="ts">
  import type { Snippet } from "svelte";
  import { onMount, getContext } from "svelte";
  import { t } from "@/i18n/index.svelte";
  import ResourceTable from "@/components/common/ResourceTable.svelte";
  import FormModal from "@/components/common/FormModal.svelte";
  import FieldInput from "@/components/common/FieldInput.svelte";
  import Button from "@/components/ui/Button.svelte";
  import Badge from "@/components/ui/Badge.svelte";
  import Alert from "@/components/ui/Alert.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { fetchWithAuth } from "@/utils/auth";

  interface APIKey {
    id: string;
    name: string;
    enabled: boolean;
    last_used_at?: string;
    expires_at?: string;
  }

  let apiKeys = $state<APIKey[]>([]);
  let showAPIKeyModal = $state(false);
  let newAPIKey = $state("");

  // Get trigger ref from parent and register open function
  const apiKeyModalTrigger = getContext<{ value: (() => void) | null }>("apiKeyModalTrigger");

  const apiKeyColumns = $derived.by(() => [
    { key: "name", label: t("apikeys.name"), sortable: true },
    {
      key: "last_used_at",
      label: t("apikeys.lastUsed"),
      formatter: (value: unknown): string => (value ? new Date(String(value)).toLocaleDateString() : t("apikeys.never"))
    },
    {
      key: "expires_at",
      label: t("apikeys.expires"),
      formatter: (value: unknown): string => (value ? new Date(String(value)).toLocaleDateString() : t("apikeys.never"))
    },
    { key: "enabled", label: t("apikeys.status") }
  ]);

  onMount(async () => {
    await loadAPIKeys();
  });

  function openAPIKeyModal(): void {
    newAPIKey = "";
    showAPIKeyModal = true;
  }

  // Register open function with parent
  if (apiKeyModalTrigger) {
    apiKeyModalTrigger.value = () => {
      openAPIKeyModal();
    };
  }

  async function loadAPIKeys(): Promise<void> {
    try {
      const response = await fetchWithAuth("/api/v1/apikeys");
      if (response.ok) {
        const data = await response.json();
        if (Array.isArray(data.data)) {
          apiKeys = data.data;
        } else {
          apiKeys = [];
        }
      }
    } catch (error) {
      if (!window.location.pathname.includes("/auth/") && !window.location.pathname.includes("/signin")) {
        console.error("Failed to load API keys:", error);
      }
    }
  }

  async function saveAPIKey(formData: Record<string, unknown>): Promise<void> {
    try {
      const response = await fetchWithAuth("/api/v1/apikeys", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(formData)
      });

      if (response.ok) {
        const data = await response.json();
        newAPIKey = data.data.key;
        await loadAPIKeys();
        setTimeout(() => {
          if (confirm(t("apikeys.savedKeyConfirm"))) {
            showAPIKeyModal = false;
            newAPIKey = "";
          }
        }, 2000);
      } else {
        alert(t("apikeys.createError"));
      }
    } catch (error) {
      console.error("Failed to create API key:", error);
      alert(t("apikeys.createError"));
    }
  }

  async function deleteAPIKey(apiKey: APIKey): Promise<void> {
    if (!confirm(t("apikeys.deleteConfirm"))) {
      return;
    }

    try {
      const response = await fetchWithAuth(`/api/v1/apikeys/${apiKey.id}`, {
        method: "DELETE"
      });

      if (response.ok) {
        await loadAPIKeys();
      } else {
        alert(t("apikeys.deleteError"));
      }
    } catch (error) {
      console.error("Failed to delete API key:", error);
      alert(t("apikeys.deleteError"));
    }
  }

  async function toggleAPIKey(apiKey: APIKey): Promise<void> {
    const action = apiKey.enabled ? "disable" : "enable";

    try {
      const response = await fetchWithAuth(`/api/v1/apikeys/${apiKey.id}/${action}`, {
        method: "PUT"
      });

      if (response.ok) {
        await loadAPIKeys();
      } else {
        alert(t(`apikeys.${action}Error`));
      }
    } catch (error) {
      console.error(`Failed to ${action} API key:`, error);
      alert(t(`apikeys.${action}Error`));
    }
  }
</script>

{#snippet enabledCell(_item: unknown, value: string)}
  <Badge variant={value ? "success" : "ghost"}>
    {value ? t("apikeys.active") : t("apikeys.disabled")}
  </Badge>
{/snippet}

{#snippet actionsCell(item: unknown)}
  <Button
    size="sm"
    variant={(item as APIKey).enabled ? "warning" : "success"}
    onclick={() => toggleAPIKey(item as APIKey)}
  >
    {(item as APIKey).enabled ? t("apikeys.disable") : t("apikeys.enable")}
  </Button>
{/snippet}

<div class="space-y-4">
  <ResourceTable
    data={apiKeys}
    columns={apiKeyColumns}
    hasActions={true}
    showPagination={false}
    searchable={false}
    showFilters={false}
    showEdit={false}
    emptyMessage={t("apikeys.noApiKeys")}
    cellSnippets={{ enabled: enabledCell as Snippet<[item: unknown, value: string]> }}
    actions={actionsCell}
    onDelete={(item) => deleteAPIKey(item as APIKey)}
  />

  <!-- API Key Modal -->
  <FormModal bind:open={showAPIKeyModal} title={t("apikeys.createApiKey")} onSubmit={saveAPIKey}>
    {#snippet children(formData)}
      <div class="space-y-4">
        <FieldInput bind:value={formData.name} type="text" placeholder={t("apikeys.keyNamePlaceholder")} required />

        <FieldInput bind:value={formData.expires_at} type="date" />

        {#if newAPIKey}
          <Alert variant="info">
            {#snippet icon()}
              <SvgIcon name="info" class="h-6 w-6 shrink-0" />
            {/snippet}
            <div>
              <p class="font-bold">{t("apikeys.saveThisKey")}</p>
              <p class="text-sm">{newAPIKey}</p>
              <p class="mt-1 text-xs">{t("apikeys.wontBeShownAgain")}</p>
            </div>
          </Alert>
        {/if}
      </div>
    {/snippet}
  </FormModal>
</div>
