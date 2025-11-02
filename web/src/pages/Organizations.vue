<template>
  <div class="organizations-page">
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold">Organizations</h1>
        <p class="mt-1 text-sm text-gray-600">Manage your organizations and team members</p>
      </div>
      <Button variant="primary" @click="showCreateModal = true">
        <PlusIcon class="mr-2 h-5 w-5" />
        Create Organization
      </Button>
    </div>

    <!-- Content -->
    <div class="organizations-content">
      <!-- Organizations Grid -->
      <div class="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="org in (organizations || []).filter((o) => o != null)"
          :key="org?.id || Math.random()"
          class="cursor-pointer overflow-hidden rounded-lg border border-gray-200 bg-white transition-colors hover:border-gray-300"
          @click="selectOrganization(org)"
        >
          <div class="p-6">
            <div class="flex items-center justify-between">
              <div class="flex items-center">
                <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg bg-blue-500">
                  <BuildingOfficeIcon class="h-6 w-6 text-white" />
                </div>
                <div class="ml-4">
                  <h3 class="text-lg font-medium text-gray-900">{{ org.name }}</h3>
                  <p class="text-sm text-gray-500">{{ org.description || "No description" }}</p>
                </div>
              </div>
              <div class="flex items-center space-x-2">
                <span
                  v-if="org.owner_id === currentUserId"
                  class="inline-flex items-center rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800"
                >
                  Owner
                </span>
                <button
                  class="rounded-full p-1 text-gray-400 hover:text-gray-600"
                  @click.stop="openOrganizationMenu(org)"
                >
                  <EllipsisVerticalIcon class="h-5 w-5" />
                </button>
              </div>
            </div>

            <div class="mt-4 flex items-center justify-between">
              <div class="flex items-center text-sm text-gray-500">
                <UsersIcon class="mr-1 h-4 w-4" />
                <span>Members</span>
              </div>
              <div class="text-sm text-gray-500">Created {{ formatDate(org.created_at) }}</div>
            </div>
          </div>
        </div>

        <!-- Empty State -->
        <div
          v-if="organizations.length === 0 && !loading"
          class="col-span-full rounded-lg border-2 border-dashed border-gray-300 bg-white p-12 text-center"
        >
          <BuildingOfficeIcon class="mx-auto h-12 w-12 text-gray-400" />
          <h3 class="mt-2 text-sm font-medium text-gray-900">No organizations</h3>
          <p class="mt-1 text-sm text-gray-500">Get started by creating your first organization.</p>
          <div class="mt-6">
            <Button variant="primary" @click="showCreateModal = true">
              <PlusIcon class="mr-2 h-5 w-5" />
              Create Organization
            </Button>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="col-span-full flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-600"></div>
      </div>
    </div>

    <!-- Create Organization Modal -->
    <CreateOrganizationModal v-if="showCreateModal" @close="showCreateModal = false" @created="onOrganizationCreated" />

    <!-- Organization Menu -->
    <OrganizationMenu
      v-if="selectedOrgForMenu"
      :organization="selectedOrgForMenu"
      @close="selectedOrgForMenu = null"
      @edit="editOrganization"
      @delete="deleteOrganization(selectedOrgForMenu)"
      @manage-members="manageMembers(selectedOrgForMenu)"
    />
  </div>
</template>

<script setup lang="ts">
import { onActivated, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { BuildingOfficeIcon, EllipsisVerticalIcon, PlusIcon, UsersIcon } from "@heroicons/vue/24/outline";
import { apiDelete, apiGet, apiPost } from "@/services/api";
import { Organization } from "@/models";
import CreateOrganizationModal from "@/components/organization/CreateOrganizationModal.vue";
import OrganizationMenu from "@/components/organization/OrganizationMenu.vue";
import Button from "@/components/ui/Button.vue";

const router = useRouter();
const route = useRoute();
const organizations = ref<Organization[]>([]);
const loading = ref(true);
const showCreateModal = ref(false);
const selectedOrgForMenu = ref<Organization | null>(null);
const currentUserId = ref("");

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

      const data = response.data?.organizations || response.data;
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
    localStorage.setItem(
      "current_organization",
      JSON.stringify({
        id: org.id,
        name: org.name,
        role: response.data.role
      })
    );

    // Navigate to dashboard or organization overview
    router.push("/dashboard");
  } catch (error) {
    console.error("Failed to switch organization:", error);
  }
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

const openOrganizationMenu = (org: Organization) => {
  selectedOrgForMenu.value = org;
};

const editOrganization = () => {
  // TODO: Implement edit organization
  selectedOrgForMenu.value = null;
};

const deleteOrganization = async (org: Organization) => {
  if (!confirm(`Are you sure you want to delete "${org.name}"? This action cannot be undone.`)) {
    return;
  }

  try {
    await apiDelete(`/api/v1/organizations/${org.id}`);
    organizations.value = organizations.value.filter((o) => o.id !== org.id);
  } catch (error) {
    console.error("Failed to delete organization:", error);
  } finally {
    selectedOrgForMenu.value = null;
  }
};

const manageMembers = (org: Organization) => {
  router.push(`/organizations/${org.id}/members`);
  selectedOrgForMenu.value = null;
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
    }
  },
  { immediate: false }
);
</script>
