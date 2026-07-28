<script lang="ts">
  /**
   * Dark 64px icon rail (WS2 primaryMenu / twing-m):
   * logo at top, section icons in the middle, user and logout at bottom.
   */
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { t } from "@/i18n/index.svelte";
  import { logout } from "@/utils/auth";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { useCurrentUser } from "@/composables/useCurrentUser.svelte";

  const currentUser = useCurrentUser();

  onMount(() => {
    currentUser.loadUserData();
  });

  function isActive(path: string): boolean {
    return page.url.pathname.startsWith(path);
  }

  const navItems = $derived([
    { path: "/dashboard", icon: "dashboard", label: t("navigation.dashboard"), active: isActive("/dashboard") },
    { path: "/submissions", icon: "submissions", label: t("navigation.submissions"), active: isActive("/submissions") },
    { path: "/templates", icon: "templates", label: t("navigation.templates"), active: isActive("/templates") },
    {
      path: "/settings",
      icon: "settings",
      label: t("navigation.settings"),
      active: isActive("/settings") && !isActive("/admin/settings")
    }
  ]);

  const adminNavItems = $derived([
    {
      path: "/admin/organizations",
      icon: "organizations",
      label: t("navigation.organizations"),
      active: isActive("/admin/organizations")
    },
    {
      path: "/admin/settings",
      icon: "settings",
      label: `${t("navigation.administrator")}: ${t("navigation.settings")}`,
      active: isActive("/admin/settings")
    }
  ]);

  const userInitial = $derived(
    currentUser.userData?.first_name
      ? currentUser.userData.first_name[0].toUpperCase()
      : currentUser.userData?.email
        ? currentUser.userData.email[0].toUpperCase()
        : "U"
  );

  const userTitle = $derived.by(() => {
    const name = `${currentUser.userData?.first_name || ""} ${currentUser.userData?.last_name || ""}`.trim();
    const email = currentUser.userData?.email || "";
    return [name, email].filter(Boolean).join(" — ") || t("navigation.user");
  });

  async function handleLogout(): Promise<void> {
    currentUser.clearUser();
    await logout();
  }
</script>

<aside class="app-sidebar">
  <a href="/dashboard" class="sb-brand" title="goSign">
    <span class="sb-logo">
      <SvgIcon name="logo" class="sb-logo-icon" />
    </span>
  </a>

  <nav class="sb-nav">
    {#each navItems as item (item.path)}
      <a href={item.path} class="sb-item" class:active={item.active} title={item.label} aria-label={item.label}>
        <SvgIcon name={item.icon} class="sb-icon" />
      </a>
    {/each}

    {#if currentUser.isAdmin}
      <div class="sb-divider" role="separator" title={t("navigation.administrator")}></div>
      {#each adminNavItems as item (item.path)}
        <a href={item.path} class="sb-item" class:active={item.active} title={item.label} aria-label={item.label}>
          <SvgIcon name={item.icon} class="sb-icon" />
        </a>
      {/each}
    {/if}
  </nav>

  <div class="sb-footer">
    <a href="/settings/general" class="sb-avatar" title={userTitle} aria-label={userTitle}>
      {userInitial}
    </a>
    <button
      type="button"
      class="sb-item sb-item--logout"
      onclick={handleLogout}
      title={t("navigation.exit")}
      aria-label={t("navigation.exit")}
    >
      <SvgIcon name="exit" class="sb-icon" />
    </button>
  </div>
</aside>

<style>
  .sb-brand {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--space-12) 0 var(--space-8);
    flex-shrink: 0;
  }
  .sb-logo {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: var(--size-control-36);
    height: var(--size-control-36);
    border-radius: var(--radius-10);
    background: linear-gradient(135deg, var(--base-hlt-invert), var(--base-hlt-b-invert));
    color: var(--base-txt-alt-light);
  }
  .sb-logo :global(.sb-logo-icon) {
    width: 22px;
    height: 22px;
  }

  .sb-nav {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-8) 0;
    overflow-y: auto;
    overflow-x: hidden;
    scrollbar-width: none;
  }

  .sb-divider {
    width: 28px;
    height: 1px;
    margin: var(--space-4) 0;
    background: var(--color-graphite-alpha-d-200);
    flex-shrink: 0;
  }

  .sb-item {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 42px;
    border: none;
    background: transparent;
    border-radius: var(--radius-10);
    cursor: pointer;
    transition: var(--transition-colors);
    flex-shrink: 0;
  }
  .sb-item :global(.sb-icon) {
    width: 21px;
    height: 21px;
    flex-shrink: 0;
    color: var(--sidebar-ico-base);
    transition: var(--transition-colors);
  }
  .sb-item:hover {
    background: var(--sidebar-cont-hover);
  }
  .sb-item:hover :global(.sb-icon) {
    color: var(--sidebar-ico-hover);
  }
  .sb-item.active {
    background: var(--sidebar-cont-active);
  }
  .sb-item.active :global(.sb-icon) {
    color: var(--sidebar-ico-active);
  }
  .sb-item:focus-visible {
    outline: none;
    box-shadow: var(--shadow-focus);
  }

  .sb-footer {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-8);
    padding: var(--space-8) 0 var(--space-12);
    border-top: 1px solid var(--color-graphite-alpha-d-100);
    flex-shrink: 0;
  }
  .sb-avatar {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: var(--radius-full);
    background: var(--sidebar-cont-active);
    color: var(--sidebar-ico-active);
    font-size: var(--font-size-11);
    font-weight: var(--font-weight-bold);
    text-decoration: none;
    transition: var(--transition-colors);
  }
  .sb-avatar:hover {
    background: var(--sidebar-cont-hover);
    color: var(--sidebar-ico-active);
  }
  .sb-item--logout:hover {
    background: var(--sidebar-cont-hover);
  }
  .sb-item--logout:hover :global(.sb-icon) {
    color: var(--base-txt-alert-minor);
  }
</style>
