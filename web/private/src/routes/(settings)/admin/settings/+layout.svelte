<script lang="ts">
  import type { Snippet } from "svelte";
  import { page } from "$app/state";
  import { t } from "@/i18n/index.svelte";

  let { children }: { children?: Snippet } = $props();

  // Map admin settings paths to title and description translation keys
  const pageInfo = $derived.by(() => {
    const path = page.url.pathname;

    const infoMap: Record<string, { title: string; description: string }> = {
      "/admin/settings/smtp": {
        title: t("settings.smtpConfiguration"),
        description: t("settings.smtpDescription")
      },
      "/admin/settings/sms": {
        title: t("settings.smsConfiguration"),
        description: t("settings.smsDescription")
      },
      "/admin/settings/storage": {
        title: t("settings.storageConfiguration"),
        description: t("settings.storageDescription")
      },
      "/admin/settings/geolocation": {
        title: t("settings.geolocation"),
        description: t("settings.geolocationSectionDescription")
      }
    };

    return (
      infoMap[path] || {
        title: t("settings.adminSettings"),
        description: t("settings.adminSettingsDescription")
      }
    );
  });

  const pageTitle = $derived(pageInfo.title);
  const pageDescription = $derived(pageInfo.description);
</script>

<div class="settings-page min-h-full">
  <!-- Header -->
  <div class="page-header">
    <div>
      <h1>{pageTitle}</h1>
      <p class="page-subtitle">{pageDescription}</p>
    </div>
  </div>

  <!-- Content Area -->
  {@render children?.()}
</div>
