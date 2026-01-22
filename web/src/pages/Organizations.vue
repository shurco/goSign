<template>
  <div class="organizations-page">
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold">{{ $t('organizations.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ $t('organizations.description') }}</p>
      </div>
      <div class="flex items-center gap-3">
        <Button variant="primary" @click="openCreateModal">
          <SvgIcon name="plus" class="mr-2 h-5 w-5" />
          {{ $t('organizations.createOrganization') }}
        </Button>
      </div>
    </div>

    <!-- Organizations Table -->
    <ResourceTable
      :data="organizations"
      :columns="columns"
      :is-loading="loading"
      searchable
      :search-keys="['name', 'description']"
      :search-placeholder="$t('organizations.searchOrganizations')"
      :empty-message="$t('organizations.noOrganizations')"
      :show-edit="false"
      :show-delete="false"
    >
      <template #cell-name="{ item }">
        <div class="flex items-center gap-2">
          <button
            class="cursor-pointer text-left font-medium text-gray-900 hover:text-blue-600"
            @click="selectOrganization(item as Organization)"
          >
            {{ (item as Organization).name }}
          </button>
          <span
            v-if="currentOrganization && (item as Organization).id === currentOrganization.id"
            class="inline-flex items-center rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-800"
          >
            {{ $t('organizations.current') }}
          </span>
        </div>
      </template>

      <template #cell-description="{ value, item }">
        <span class="text-sm text-gray-500">
          {{ value || $t('organizations.noDescription') }}
        </span>
      </template>

      <template #cell-created_at="{ value }">
        <span class="text-sm text-gray-500">{{ formatDate(value) }}</span>
      </template>

      <template #actions="{ item }">
        <div class="flex items-center justify-end gap-2">
          <button
            v-if="currentOrganization && (item as Organization).id === currentOrganization.id"
            class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
            @click.stop="exitOrganization"
            :title="$t('organizations.exitOrganization')"
          >
            <SvgIcon name="exit" class="h-5 w-5 stroke-[2]" />
          </button>
          <button
            class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
            @click.stop="editOrganization(item as Organization)"
            :title="$t('organizations.editOrganization')"
          >
            <SvgIcon name="settings" class="h-5 w-5 stroke-[2]" />
          </button>
          <button
            class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
            @click.stop="manageMembers(item as Organization)"
            :title="$t('organizations.manageMembers')"
          >
            <SvgIcon name="users" class="h-5 w-5 stroke-[2]" />
          </button>
        </div>
      </template>
    </ResourceTable>

    <!-- Create Organization Modal -->
    <CreateOrganizationModal v-model="showCreateModal" @created="onOrganizationCreated" />

    <!-- Edit Organization Modal -->
    <EditOrganizationModal
      v-model="showEditModal"
      :organization="selectedOrgForEdit"
      @updated="onOrganizationUpdated"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { apiDelete, apiGet, apiPost } from "@/services/api";
import { Organization } from "@/models";
import CreateOrganizationModal from "@/components/organization/CreateOrganizationModal.vue";
import EditOrganizationModal from "@/components/organization/EditOrganizationModal.vue";
import Button from "@/components/ui/Button.vue";
import ResourceTable from "@/components/common/ResourceTable.vue";
import SvgIcon from "@/components/SvgIcon.vue";

const router = useRouter();
const route = useRoute();
const { t } = useI18n();
const organizations = ref<Organization[]>([]);
const loading = ref(true);
const showCreateModal = ref(false);
const showEditModal = ref(false);
const selectedOrgForEdit = ref<Organization | null>(null);
const currentUserId = ref("");
const currentOrganization = ref<{ id: string; name: string; role?: string } | null>(null);

const columns = computed(() => [
  { key: "name", label: t('organizations.organizationName'), sortable: true },
  { key: "description", label: t('organizations.description'), sortable: false },
  {
    key: "created_at",
    label: t('submissions.created'),
    sortable: true,
    formatter: (value: unknown): string => formatDate(value as string)
  }
]);

let loadOrganizationsPromise: Promise<void> | null = null;

const loadOrganizations = async () => {
  // Prevent multiple simultaneous loads
  if (loadOrganizationsPromise) {
    return loadOrganizationsPromise;
  }

  // Check if we're already redirecting before starting load
  if (window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin")) {
    loading.value = false;
    return Promise.resolve();
  }

  loading.value = true;
  loadOrganizationsPromise = (async () => {
    try {
      // Check if we're already redirecting before making request
      if (window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin")) {
        loading.value = false;
        return;
      }

      const response = await apiGet("/api/v1/organizations");

      // Check again after request if redirect happened
      if (window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin")) {
        loading.value = false;
        return;
      }

      // API returns: { success: true, message: "...", data: [...] } or { success: true, message: "...", data: { organizations: [...] } }
      let data = response.data;
      if (data && typeof data === 'object' && 'organizations' in data) {
        data = data.organizations;
      }
      // Ensure organizations is always an array
      organizations.value = Array.isArray(data) ? data : [];
    } catch (error: any) {
      // Don't log 401 errors if we're being redirected to login
      const isRedirecting = window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin");
      const is401Error = error?.status === 401 || error?.message?.includes("Unauthorized");

      // Only log if not redirecting and not a 401 error (401 will be handled by redirect)
      if (!isRedirecting && !is401Error && window.location.pathname.includes("/organizations")) {
        console.error("Failed to load organizations:", error);
      }

      // Ensure organizations is always an array even on error
      if (!Array.isArray(organizations.value)) {
        organizations.value = [];
      }

      // If auth failed, redirect will happen automatically
      if (!window.location.pathname.includes("/organizations")) {
        loading.value = false;
        return;
      }
    } finally {
      // Always set loading to false, even if redirecting
      loading.value = false;
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
    currentOrganization.value = orgData;

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
    currentOrganization.value = null;

    // Reload organizations to refresh the list
    await loadOrganizations();
  } catch (error) {
    console.error("Failed to exit organization:", error);
    alert(t('organizations.exitError') || 'Failed to exit organization');
  }
};

const openCreateModal = () => {
  showCreateModal.value = true;
};

const handleCloseModal = () => {
  showCreateModal.value = false;
};

const onOrganizationCreated = (newOrg: Organization) => {
  if (newOrg && newOrg.id) {
    // Check if organization already exists to avoid duplicates
    const exists = organizations.value.some((org) => org.id === newOrg.id);
    if (!exists) {
      organizations.value.push(newOrg);
    }
  } else {
    console.error("Invalid organization data received:", newOrg);
    // Reload organizations to get fresh data
    loadOrganizations();
  }
  showCreateModal.value = false;
};

const editOrganization = (org: Organization) => {
  selectedOrgForEdit.value = org;
  showEditModal.value = true;
};

const onOrganizationUpdated = (updatedOrg: Organization) => {
  const index = organizations.value.findIndex((o) => o.id === updatedOrg.id);
  if (index !== -1) {
    // Update organization in place to ensure Vue reactivity
    Object.assign(organizations.value[index], updatedOrg);
    console.log("Organization updated in list at index:", index, "New value:", organizations.value[index]);
  } else {
    // If not found, reload organizations
    loadOrganizations();
  }
  showEditModal.value = false;
  selectedOrgForEdit.value = null;
};

const manageMembers = (org: Organization) => {
  router.push(`/organizations/${org.id}/members`);
};

const formatDate = (dateString: string | undefined) => {
  if (!dateString) {
    return "";
  }
  return new Date(dateString).toLocaleDateString();
};

let hasLoadedOnce = false;

onMounted(() => {
  loadOrganizations().then(() => {
    hasLoadedOnce = true;
  });
  // TODO: Get current user ID
  currentUserId.value = "user-id"; // Replace with actual user ID
  
  // Load current organization from localStorage
  const storedOrg = localStorage.getItem("current_organization");
  if (storedOrg) {
    try {
      currentOrganization.value = JSON.parse(storedOrg);
    } catch (e) {
      console.error("Failed to parse current organization:", e);
    }
  }
});

// Reload organizations when component is activated (reused by router)
// Only reload if we haven't loaded yet or data might be stale
onActivated(() => {
  if (route.path === "/organizations" && (!hasLoadedOnce || organizations.value.length === 0)) {
    loadOrganizations().then(() => {
      hasLoadedOnce = true;
    });
  }
});

// Reload organizations when navigating to this page via router
// Only if component is reused and data is empty or stale
watch(
  () => route.path,
  (newPath, oldPath) => {
    // Only reload if navigating TO this page (not from it) and we need fresh data
    if (newPath === "/organizations" && oldPath !== newPath) {
      // Only reload if data is empty or component was reused
      if (organizations.value.length === 0 || !hasLoadedOnce) {
        setTimeout(() => {
          if (newPath === route.path && !loading.value) {
            loadOrganizations().then(() => {
              hasLoadedOnce = true;
            });
          }
        }, 50);
      }
      // Reload current organization from localStorage when navigating to this page
      const storedOrg = localStorage.getItem("current_organization");
      if (storedOrg) {
        try {
          currentOrganization.value = JSON.parse(storedOrg);
        } catch (e) {
          console.error("Failed to parse current organization:", e);
          currentOrganization.value = null;
        }
      } else {
        currentOrganization.value = null;
      }
    }
  },
  { immediate: false }
);

// Watch for changes in localStorage to update current organization
// This handles cases where organization is switched from another page
const updateCurrentOrganization = () => {
  const storedOrg = localStorage.getItem("current_organization");
  if (storedOrg) {
    try {
      currentOrganization.value = JSON.parse(storedOrg);
    } catch (e) {
      console.error("Failed to parse current organization:", e);
      currentOrganization.value = null;
    }
  } else {
    currentOrganization.value = null;
  }
};

// Listen for storage events (when localStorage changes in another tab/window)
window.addEventListener("storage", updateCurrentOrganization);

// Also check periodically (for same-tab updates)
setInterval(updateCurrentOrganization, 1000);
</script>
