<script lang="ts">
  import type { Snippet } from "svelte";
  import { setContext } from "svelte";
  import { page } from "$app/state";
  import { t } from "@/i18n/index.svelte";
  import Button from "@/components/ui/Button.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

  let { children }: { children?: Snippet } = $props();

  // Map settings paths to title and description translation keys
  const pageInfo = $derived.by(() => {
    const path = page.url.pathname;

    const infoMap: Record<string, { title: string; description: string }> = {
      "/settings/general": {
        title: t("settings.generalSettings"),
        description: t("settings.generalDescription")
      },
      "/settings/webhooks": {
        title: t("webhooks.title"),
        description: t("settings.description")
      },
      "/settings/api-keys": {
        title: t("apikeys.title"),
        description: t("settings.description")
      },
      "/settings/branding": {
        title: t("branding.title"),
        description: t("branding.description")
      },
      "/settings/email/templates": {
        title: t("settings.emailTemplates"),
        description: t("settings.emailTemplatesDescription")
      }
    };

    return (
      infoMap[path] || {
        title: t("settings.title"),
        description: t("settings.description")
      }
    );
  });

  const pageTitle = $derived(pageInfo.title);
  const pageDescription = $derived(pageInfo.description);

  const isWebhooksPage = $derived(page.url.pathname === "/settings/webhooks");
  const isApiKeysPage = $derived(page.url.pathname === "/settings/api-keys");

  // Show action button for webhooks and api-keys pages
  const showActionButton = $derived(isWebhooksPage || isApiKeysPage);

  // Provide modal triggers to child pages (Ref-like shape, see PORTING.md)
  let webhookModalTrigger = $state<(() => void) | null>(null);
  let apiKeyModalTrigger = $state<(() => void) | null>(null);

  setContext("webhookModalTrigger", {
    get value() {
      return webhookModalTrigger;
    },
    set value(fn: (() => void) | null) {
      webhookModalTrigger = fn;
    }
  });
  setContext("apiKeyModalTrigger", {
    get value() {
      return apiKeyModalTrigger;
    },
    set value(fn: (() => void) | null) {
      apiKeyModalTrigger = fn;
    }
  });

  function triggerWebhookModal(): void {
    if (webhookModalTrigger) {
      webhookModalTrigger();
    }
  }

  function triggerApiKeyModal(): void {
    if (apiKeyModalTrigger) {
      apiKeyModalTrigger();
    }
  }
</script>

<div class="settings-page min-h-full">
  <!-- Header -->
  <div class="mb-6 flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold">{pageTitle}</h1>
      <p class="mt-1 text-sm text-gray-600">{pageDescription}</p>
    </div>
    {#if showActionButton}
      <div class="flex items-center gap-3">
        {#if isWebhooksPage}
          <Button variant="primary" onclick={triggerWebhookModal}>
            <SvgIcon name="plus" class="mr-2 h-5 w-5" />
            {t("webhooks.addWebhook")}
          </Button>
        {:else if isApiKeysPage}
          <Button variant="primary" onclick={triggerApiKeyModal}>
            <SvgIcon name="plus" class="mr-2 h-5 w-5" />
            {t("apikeys.createApiKey")}
          </Button>
        {/if}
      </div>
    {/if}
  </div>

  <!-- Content Area -->
  {@render children?.()}
</div>
