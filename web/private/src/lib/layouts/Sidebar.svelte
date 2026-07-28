<script lang="ts">
  import type { Snippet } from "svelte";
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { t } from "@/i18n/index.svelte";
  import { logout } from "@/utils/auth";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { useCurrentUser } from "@/composables/useCurrentUser.svelte";

  let { children }: { children?: Snippet } = $props();

  const currentUser = useCurrentUser();

  let isCollapsed = $state(false);

  onMount(() => {
    currentUser.loadUserData();
  });

  /**
   * Check if the given path is active
   */
  function isActive(path: string): boolean {
    return page.url.pathname.startsWith(path);
  }

  /**
   * Toggle sidebar collapsed state
   */
  function toggleSidebar(): void {
    isCollapsed = !isCollapsed;
  }

  /**
   * Handle logout - clear tokens and redirect to login
   */
  async function handleLogout(): Promise<void> {
    currentUser.clearUser();
    await logout();
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
    { path: "/admin/settings", icon: "settings", label: t("navigation.settings"), active: isActive("/admin/settings") }
  ]);

  const userInitial = $derived(
    currentUser.userData?.first_name
      ? currentUser.userData.first_name[0].toUpperCase()
      : currentUser.userData?.email
        ? currentUser.userData.email[0].toUpperCase()
        : "U"
  );

  const userName = $derived(
    currentUser.userData?.first_name || currentUser.userData?.last_name
      ? `${currentUser.userData?.first_name || ""} ${currentUser.userData?.last_name || ""}`.trim() ||
          t("navigation.user")
      : t("navigation.user")
  );
</script>

<div class="relative flex h-screen overflow-hidden bg-[var(--color-base-100)]">
  <!-- Sidebar Navigation -->
  <aside
    class="flex h-screen flex-col border-e border-[#e7e2df] bg-white transition-all duration-150 {isCollapsed
      ? 'w-16'
      : 'w-48'}"
  >
    <!-- Logo -->
    <div class="flex h-14 items-center border-b border-gray-100 px-3">
      <div class="flex items-center overflow-hidden">
        <SvgIcon name="logo" stroke="currentColor" class="h-6 w-6 flex-shrink-0" />
        <span
          class="ml-2.5 text-base font-bold whitespace-nowrap text-gray-800 transition-opacity duration-150 {isCollapsed
            ? 'opacity-0'
            : 'opacity-100'}"
          hidden={isCollapsed}
        >
          goSign
        </span>
      </div>
    </div>

    <!-- Navigation Menu -->
    <nav class="flex-1 overflow-hidden px-2 py-3">
      <ul class="space-y-0.5">
        {#each navItems as item (item.path)}
          <li>
            <a
              href={item.path}
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900 {item.active
                ? 'bg-gray-100 text-gray-900'
                : ''} {isCollapsed ? 'justify-center' : ''}"
              title={isCollapsed ? item.label : ""}
            >
              <SvgIcon name={item.icon} class="h-4 w-4 flex-shrink-0" />
              <span class="text-[13px] whitespace-nowrap" hidden={isCollapsed}>{item.label}</span>
              {#if isCollapsed}
                <span
                  class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
                >
                  {item.label}
                </span>
              {/if}
            </a>
          </li>
        {/each}

        <!-- Administrator Section (only for admins) -->
        <li class="pt-3" hidden={!currentUser.isAdmin}>
          <div
            class="mb-1.5 px-2.5 text-[11px] font-semibold tracking-wider text-gray-400 uppercase"
            hidden={isCollapsed}
          >
            {t("navigation.administrator")}
          </div>
          {#if isCollapsed}
            <div class="mb-1.5 border-t border-gray-200"></div>
          {/if}
        </li>
        {#each adminNavItems as item (item.path)}
          <li hidden={!currentUser.isAdmin}>
            <a
              href={item.path}
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900 {item.active
                ? 'bg-gray-100 text-gray-900'
                : ''} {isCollapsed ? 'justify-center' : ''}"
              title={isCollapsed ? item.label : ""}
            >
              <SvgIcon name={item.icon} class="h-4 w-4 flex-shrink-0" />
              <span class="text-[13px] whitespace-nowrap" hidden={isCollapsed}>{item.label}</span>
              {#if isCollapsed}
                <span
                  class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
                >
                  {item.label}
                </span>
              {/if}
            </a>
          </li>
        {/each}
      </ul>
    </nav>

    <!-- Toggle Button -->
    <div class="flex justify-center border-t border-gray-100 py-2">
      <button
        class="group relative flex h-7 w-7 items-center justify-center rounded-md text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
        title={isCollapsed ? t("navigation.expandSidebar") : t("navigation.collapseSidebar")}
        onclick={toggleSidebar}
      >
        <SvgIcon
          name="sidebar-toggle"
          class="h-3.5 w-3.5 transition-transform duration-150 {isCollapsed ? 'rotate-180' : ''}"
        />
        {#if isCollapsed}
          <span
            class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
          >
            {t("navigation.expand")}
          </span>
        {/if}
      </button>
    </div>

    <!-- User Section -->
    <div class="border-t border-gray-100 p-2.5">
      {#if !isCollapsed}
        <div class="flex items-center gap-2.5 px-0.5">
          <div
            class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-gray-200 text-xs font-medium text-gray-600"
          >
            {userInitial}
          </div>
          <div class="flex-1 overflow-hidden">
            <p class="truncate text-[13px] font-medium text-gray-900">{userName}</p>
            <p class="truncate text-[11px] text-gray-500">
              {currentUser.userData?.email || t("navigation.loading")}
            </p>
          </div>
        </div>
      {:else}
        <div class="flex justify-center">
          <div
            class="flex h-8 w-8 items-center justify-center rounded-full bg-gray-200 text-xs font-medium text-gray-600"
          >
            {userInitial}
          </div>
        </div>
      {/if}
      <div class="mt-2 {isCollapsed ? 'flex justify-center' : ''}">
        <button
          type="button"
          class="group relative flex w-full items-center justify-center gap-1.5 rounded-md border border-gray-200 px-2.5 py-1.5 text-[13px] font-medium text-gray-700 transition-colors hover:bg-gray-50 {isCollapsed
            ? 'h-8 w-8 p-0'
            : 'w-full'}"
          title={isCollapsed ? t("navigation.exit") : ""}
          onclick={handleLogout}
        >
          <SvgIcon name="exit" class="h-3.5 w-3.5 flex-shrink-0" />
          <span hidden={isCollapsed}>{t("navigation.exit")}</span>
          {#if isCollapsed}
            <span
              class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
            >
              {t("navigation.exit")}
            </span>
          {/if}
        </button>
      </div>
    </div>
  </aside>

  <!-- Main Content Area -->
  <main class="relative block flex-1 overflow-x-hidden overflow-y-auto px-6 py-6">
    {@render children?.()}
  </main>
</div>
