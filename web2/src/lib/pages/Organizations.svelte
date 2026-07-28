<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { t } from "@/i18n/index.svelte";
  import { apiGet, apiPost } from "@/services/api";
  import type { Organization } from "@/models";
  import CreateOrganizationModal from "@/components/organization/CreateOrganizationModal.svelte";
  import EditOrganizationModal from "@/components/organization/EditOrganizationModal.svelte";
  import Button from "@/components/ui/Button.svelte";
  import ResourceTable from "@/components/common/ResourceTable.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

  let organizations = $state<Organization[]>([]);
  let loading = $state(true);
  let showCreateModal = $state(false);
  let showEditModal = $state(false);
  let selectedOrgForEdit = $state<Organization | null>(null);
  let currentOrganization = $state<{ id: string; name: string; role?: string } | null>(null);

  const columns = $derived([
    { key: "name", label: t("organizations.organizationName"), sortable: true },
    { key: "description", label: t("organizations.description"), sortable: false },
    {
      key: "created_at",
      label: t("submissions.created"),
      sortable: true,
      formatter: (value: unknown): string => formatDate(value as string)
    }
  ]);

  let loadOrganizationsPromise: Promise<void> | null = null;
  let hasLoadedOnce = false;
  let previousPath: string | null = null;

  const loadOrganizations = async () => {
    // Prevent multiple simultaneous loads
    if (loadOrganizationsPromise) {
      return loadOrganizationsPromise;
    }

    // Check if we're already redirecting before starting load
    if (window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin")) {
      loading = false;
      return Promise.resolve();
    }

    loading = true;
    loadOrganizationsPromise = (async () => {
      try {
        // Check if we're already redirecting before making request
        if (window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin")) {
          loading = false;
          return;
        }

        const response = await apiGet("/api/v1/organizations");

        // Check again after request if redirect happened
        if (window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin")) {
          loading = false;
          return;
        }

        // API returns: { success: true, message: "...", data: [...] } or { success: true, message: "...", data: { organizations: [...] } }
        let data = response.data;
        if (data && typeof data === "object" && "organizations" in data) {
          data = data.organizations;
        }
        // Ensure organizations is always an array
        organizations = Array.isArray(data) ? data : [];
      } catch (error: any) {
        // Don't log 401 errors if we're being redirected to login
        const isRedirecting =
          window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin");
        const is401Error = error?.status === 401 || error?.message?.includes("Unauthorized");

        // Only log if not redirecting and not a 401 error (401 will be handled by redirect)
        if (!isRedirecting && !is401Error && window.location.pathname.includes("/admin/organizations")) {
          console.error("Failed to load organizations:", error);
        }

        // Ensure organizations is always an array even on error
        if (!Array.isArray(organizations)) {
          organizations = [];
        }

        // If auth failed, redirect will happen automatically
        if (!window.location.pathname.includes("/admin/organizations")) {
          loading = false;
          return;
        }
      } finally {
        // Always set loading to false, even if redirecting
        loading = false;
        loadOrganizationsPromise = null;
      }
    })();

    return loadOrganizationsPromise;
  };

  const selectOrganization = async (org: Organization) => {
    try {
      const response = await apiPost(`/api/v1/organizations/${org.id}/switch`);

      // Store new tokens
      localStorage.setItem("access_token", response.data.access_token);
      localStorage.setItem("refresh_token", response.data.refresh_token);

      // Update current organization in localStorage
      const orgData = {
        id: org.id,
        name: org.name,
        role: response.data.role
      };
      localStorage.setItem("current_organization", JSON.stringify(orgData));
      currentOrganization = orgData;

      // Reload organizations to refresh the list and update current organization indicator
      await loadOrganizations();
    } catch (error) {
      console.error("Failed to switch organization:", error);
    }
  };

  const exitOrganization = async () => {
    try {
      const response = await apiPost("/api/v1/organizations/switch");

      // Store new tokens
      localStorage.setItem("access_token", response.data.access_token);
      localStorage.setItem("refresh_token", response.data.refresh_token);

      // Clear current organization from localStorage
      localStorage.removeItem("current_organization");
      currentOrganization = null;

      // Reload organizations to refresh the list
      await loadOrganizations();
    } catch (error) {
      console.error("Failed to exit organization:", error);
      alert(t("organizations.exitError") || "Failed to exit organization");
    }
  };

  const openCreateModal = () => {
    showCreateModal = true;
  };

  const onOrganizationCreated = (newOrg: Organization) => {
    if (newOrg && newOrg.id) {
      // Check if organization already exists to avoid duplicates
      const exists = organizations.some((org) => org.id === newOrg.id);
      if (!exists) {
        organizations.push(newOrg);
      }
    } else {
      console.error("Invalid organization data received:", newOrg);
      // Reload organizations to get fresh data
      loadOrganizations();
    }
    showCreateModal = false;
  };

  const editOrganization = (org: Organization) => {
    selectedOrgForEdit = org;
    showEditModal = true;
  };

  const onOrganizationUpdated = (updatedOrg: Organization) => {
    const index = organizations.findIndex((o) => o.id === updatedOrg.id);
    if (index !== -1) {
      // Update organization in place to ensure reactivity
      Object.assign(organizations[index], updatedOrg);
    } else {
      // If not found, reload organizations
      loadOrganizations();
    }
    showEditModal = false;
    selectedOrgForEdit = null;
  };

  const manageMembers = (org: Organization) => {
    goto(`/admin/organizations/${org.id}/members`);
  };

  const formatDate = (dateString: string | undefined) => {
    if (!dateString) {
      return "";
    }
    return new Date(dateString).toLocaleDateString();
  };

  const updateCurrentOrganization = () => {
    const storedOrg = localStorage.getItem("current_organization");
    if (storedOrg) {
      try {
        currentOrganization = JSON.parse(storedOrg);
      } catch (e) {
        console.error("Failed to parse current organization:", e);
        currentOrganization = null;
      }
    } else {
      currentOrganization = null;
    }
  };

  const reloadCurrentOrganizationFromStorage = () => {
    const storedOrg = localStorage.getItem("current_organization");
    if (storedOrg) {
      try {
        currentOrganization = JSON.parse(storedOrg);
      } catch (e) {
        console.error("Failed to parse current organization:", e);
        currentOrganization = null;
      }
    } else {
      currentOrganization = null;
    }
  };

  onMount(() => {
    loadOrganizations().then(() => {
      hasLoadedOnce = true;
    });

    // Load current organization from localStorage
    const storedOrg = localStorage.getItem("current_organization");
    if (storedOrg) {
      try {
        currentOrganization = JSON.parse(storedOrg);
      } catch (e) {
        console.error("Failed to parse current organization:", e);
      }
    }

    // Listen for storage events (when localStorage changes in another tab/window)
    window.addEventListener("storage", updateCurrentOrganization);

    // Also check periodically (for same-tab updates)
    const intervalId = setInterval(updateCurrentOrganization, 1000);

    return () => {
      window.removeEventListener("storage", updateCurrentOrganization);
      clearInterval(intervalId);
    };
  });

  // Reload organizations when component is activated (reused by router)
  // Only reload if we haven't loaded yet or data might be stale
  $effect(() => {
    const path = page.url.pathname;
    if (path === "/admin/organizations" && (!hasLoadedOnce || organizations.length === 0)) {
      loadOrganizations().then(() => {
        hasLoadedOnce = true;
      });
    }
  });

  // Reload organizations when navigating to this page via router
  // Only if component is reused and data is empty or stale
  $effect(() => {
    const newPath = page.url.pathname;
    if (previousPath === null) {
      previousPath = newPath;
      return;
    }
    const oldPath = previousPath;
    previousPath = newPath;

    // Only reload if navigating TO this page (not from it) and we need fresh data
    if (newPath === "/admin/organizations" && oldPath !== newPath) {
      // Only reload if data is empty or component was reused
      if (organizations.length === 0 || !hasLoadedOnce) {
        setTimeout(() => {
          if (newPath === page.url.pathname && !loading) {
            loadOrganizations().then(() => {
              hasLoadedOnce = true;
            });
          }
        }, 50);
      }
      // Reload current organization from localStorage when navigating to this page
      reloadCurrentOrganizationFromStorage();
    }
  });
</script>

<div class="organizations-page">
  <!-- Header -->
  <div class="mb-6 flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold">{t("organizations.title")}</h1>
      <p class="mt-1 text-sm text-gray-600">{t("organizations.description")}</p>
    </div>
    <div class="flex items-center gap-3">
      <Button variant="primary" onclick={openCreateModal}>
        <SvgIcon name="plus" class="mr-2 h-5 w-5" />
        {t("organizations.createOrganization")}
      </Button>
    </div>
  </div>

  <!-- Organizations Table -->
  {#snippet cellName(item: unknown, _value: string)}
    <div class="flex items-center gap-2">
      <button
        class="cursor-pointer text-left font-medium text-gray-900 hover:text-blue-600"
        onclick={() => selectOrganization(item as Organization)}
      >
        {(item as Organization).name}
      </button>
      {#if currentOrganization && (item as Organization).id === currentOrganization.id}
        <span class="inline-flex items-center rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-800">
          {t("organizations.current")}
        </span>
      {/if}
    </div>
  {/snippet}

  {#snippet cellDescription(_item: unknown, value: string)}
    <span class="text-sm text-gray-500">
      {value || t("organizations.noDescription")}
    </span>
  {/snippet}

  {#snippet cellCreatedAt(_item: unknown, value: string)}
    <span class="text-sm text-gray-500">{formatDate(value)}</span>
  {/snippet}

  {#snippet rowActions(item: unknown)}
    <div class="flex items-center justify-end gap-2">
      {#if currentOrganization && (item as Organization).id === currentOrganization.id}
        <button
          class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
          title={t("organizations.exitOrganization")}
          onclick={(e) => {
            e.stopPropagation();
            exitOrganization();
          }}
        >
          <SvgIcon name="exit" class="h-5 w-5 stroke-[2]" />
        </button>
      {/if}
      <button
        class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
        title={t("organizations.editOrganization")}
        onclick={(e) => {
          e.stopPropagation();
          editOrganization(item as Organization);
        }}
      >
        <SvgIcon name="settings" class="h-5 w-5 stroke-[2]" />
      </button>
      <button
        class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
        title={t("organizations.manageMembers")}
        onclick={(e) => {
          e.stopPropagation();
          manageMembers(item as Organization);
        }}
      >
        <SvgIcon name="users" class="h-5 w-5 stroke-[2]" />
      </button>
    </div>
  {/snippet}

  <ResourceTable
    data={organizations}
    {columns}
    isLoading={loading}
    searchable
    searchKeys={["name", "description"]}
    searchPlaceholder={t("organizations.searchOrganizations")}
    emptyMessage={t("organizations.noOrganizations")}
    showEdit={false}
    showDelete={false}
    cellSnippets={{ name: cellName, description: cellDescription, created_at: cellCreatedAt }}
    actions={rowActions}
  />

  <!-- Create Organization Modal -->
  <CreateOrganizationModal bind:open={showCreateModal} onCreated={onOrganizationCreated} />

  <!-- Edit Organization Modal -->
  <EditOrganizationModal
    bind:open={showEditModal}
    organization={selectedOrgForEdit}
    onUpdated={onOrganizationUpdated}
  />
</div>
