<script lang="ts">
  import type { Snippet } from "svelte";
  import { page } from "$app/state";
  import { t } from "@/i18n/index.svelte";
  import RailSidebar from "@/components/layout/Sidebar.svelte";
  import SectionNav from "@/components/layout/SectionNav.svelte";

  let { children }: { children?: Snippet } = $props();

  // Check if current route is admin settings
  const isAdminSettings = $derived(page.url.pathname.startsWith("/admin/settings"));

  // Organization settings tabs (available to all users; templates are per-organization)
  const organizationTabs = $derived([
    { href: "/settings/general", label: t("settings.general") },
    { href: "/settings/email/templates", label: t("settings.emailTemplates") },
    { href: "/settings/webhooks", label: t("settings.webhooks") },
    { href: "/settings/api-keys", label: t("settings.apiKeys") },
    { href: "/settings/branding", label: t("settings.branding") }
  ]);

  // Admin/Global settings tabs (only for admins)
  const adminTabs = $derived([
    { href: "/admin/settings/smtp", label: `${t("settings.email")} (${t("settings.smtp")})` },
    { href: "/admin/settings/sms", label: "SMS (Twilio)" },
    { href: "/admin/settings/storage", label: t("settings.storage") },
    { href: "/admin/settings/geolocation", label: t("settings.geolocation") }
  ]);

  const tabs = $derived(isAdminSettings ? adminTabs : organizationTabs);
  const sectionTitle = $derived(isAdminSettings ? t("settings.adminSettings") : t("navigation.settings"));
</script>

<div class="app-shell">
  <RailSidebar />

  <main class="app-main">
    <SectionNav title={sectionTitle} items={tabs} />
    <div class="section-page">
      <div class="page">
        {@render children?.()}
      </div>
    </div>
  </main>
</div>
