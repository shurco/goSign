<template>
  <div class="relative flex h-screen overflow-hidden bg-[var(--color-base-100)]">
    <!-- Main Sidebar Navigation -->
    <aside
      class="flex h-screen flex-col border-e border-[#e7e2df] bg-white transition-all duration-150"
      :class="isCollapsed ? 'w-16' : 'w-48'"
    >
      <!-- Logo -->
      <div class="flex h-14 items-center border-b border-gray-100 px-3">
        <div class="flex items-center overflow-hidden">
          <SvgIcon name="logo" stroke="currentColor" class="h-6 w-6 flex-shrink-0" />
          <span
            v-show="!isCollapsed"
            class="ml-2.5 text-base font-bold whitespace-nowrap text-gray-800 transition-opacity duration-150"
            :class="isCollapsed ? 'opacity-0' : 'opacity-100'"
          >
            goSign
          </span>
        </div>
      </div>

      <!-- Navigation Menu -->
      <nav class="flex-1 overflow-hidden px-2 py-3">
        <ul class="space-y-0.5">
          <!-- Organization Selector -->
          <li v-if="!isCollapsed" class="mb-4">
            <div class="px-2.5">
              <label class="mb-1 block text-[11px] font-semibold tracking-wider text-gray-400 uppercase">
                {{ $t('navigation.organization') }}
              </label>
              <Select
                :model-value="currentOrganizationId"
                @update:model-value="(value) => handleOrganizationChange(String(value))"
                size="sm"
                class="text-[13px]"
              >
                <option value="">{{ $t('navigation.noOrganization') }}</option>
                <option
                  v-for="org in organizations"
                  :key="org.id"
                  :value="org.id"
                >
                  {{ org.name }}
                </option>
              </Select>
            </div>
          </li>
          <li v-else class="mb-4 flex justify-center">
            <div
              class="group relative flex h-8 w-8 items-center justify-center rounded-lg text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
              :title="currentOrganizationName || $t('navigation.noOrganization')"
            >
              <SvgIcon name="organizations" class="h-4 w-4 flex-shrink-0" />
              <span
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100 whitespace-nowrap"
              >
                {{ currentOrganizationName || $t('navigation.noOrganization') }}
              </span>
            </div>
          </li>

          <li>
            <RouterLink
              to="/dashboard"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[{ 'bg-gray-100 text-gray-900': isActive('/dashboard') }, isCollapsed ? 'justify-center' : '']"
              :title="isCollapsed ? $t('navigation.dashboard') : ''"
            >
              <SvgIcon name="dashboard" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">{{ $t('navigation.dashboard') }}</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                {{ $t('navigation.dashboard') }}
              </span>
            </RouterLink>
          </li>

          <li>
            <RouterLink
              to="/submissions"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[{ 'bg-gray-100 text-gray-900': isActive('/submissions') }, isCollapsed ? 'justify-center' : '']"
              :title="isCollapsed ? $t('navigation.submissions') : ''"
            >
              <SvgIcon name="submissions" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">{{ $t('navigation.submissions') }}</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                {{ $t('navigation.submissions') }}
              </span>
            </RouterLink>
          </li>

          <li>
            <RouterLink
              to="/templates"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[{ 'bg-gray-100 text-gray-900': isActive('/templates') }, isCollapsed ? 'justify-center' : '']"
              :title="isCollapsed ? $t('navigation.templates') : ''"
            >
              <SvgIcon name="templates" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">{{ $t('navigation.templates') }}</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                {{ $t('navigation.templates') }}
              </span>
            </RouterLink>
          </li>

          <li>
            <RouterLink
              to="/organizations"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[
                { 'bg-gray-100 text-gray-900': isActive('/organizations') },
                isCollapsed ? 'justify-center' : ''
              ]"
              :title="isCollapsed ? $t('navigation.organizations') : ''"
            >
              <SvgIcon name="organizations" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">{{ $t('navigation.organizations') }}</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                {{ $t('navigation.organizations') }}
              </span>
            </RouterLink>
          </li>

          <li>
            <RouterLink
              to="/settings"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[{ 'bg-gray-100 text-gray-900': isActive('/settings') && !isActive('/admin/settings') }, isCollapsed ? 'justify-center' : '']"
              :title="isCollapsed ? $t('navigation.settings') : ''"
            >
              <SvgIcon name="settings" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">{{ $t('navigation.settings') }}</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                {{ $t('navigation.settings') }}
              </span>
            </RouterLink>
          </li>

          <!-- Administrator Section (only for admins) -->
          <li v-show="isAdmin" class="pt-3">
            <div
              v-show="!isCollapsed"
              class="mb-1.5 px-2.5 text-[11px] font-semibold tracking-wider text-gray-400 uppercase"
            >
              {{ $t('navigation.administrator') }}
            </div>
            <div v-if="isCollapsed" class="mb-1.5 border-t border-gray-200"></div>
            <RouterLink
              to="/admin/settings"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[{ 'bg-gray-100 text-gray-900': isActive('/admin/settings') }, isCollapsed ? 'justify-center' : '']"
              :title="isCollapsed ? $t('navigation.settings') : ''"
            >
              <SvgIcon name="settings" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">{{ $t('navigation.settings') }}</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                {{ $t('navigation.settings') }}
              </span>
            </RouterLink>
          </li>
        </ul>
      </nav>

      <!-- Toggle Button -->
      <div class="flex justify-center border-t border-gray-100 py-2">
        <button
          class="group relative flex h-7 w-7 items-center justify-center rounded-md text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
          :title="isCollapsed ? $t('navigation.expandSidebar') : $t('navigation.collapseSidebar')"
          @click="toggleSidebar"
        >
          <SvgIcon
            name="sidebar-toggle"
            class="h-3.5 w-3.5 transition-transform duration-150"
            :class="isCollapsed ? 'rotate-180' : ''"
          />
          <span
            v-if="isCollapsed"
            class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
          >
            {{ $t('navigation.expand') }}
          </span>
        </button>
      </div>

      <!-- User Section -->
      <div class="border-t border-gray-100 p-2.5">
        <div v-if="!isCollapsed" class="flex items-center gap-2.5 px-0.5">
          <div
            class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-gray-200 text-xs font-medium text-gray-600"
          >
            {{
              userData?.first_name
                ? userData.first_name[0].toUpperCase()
                : userData?.email
                  ? userData.email[0].toUpperCase()
                  : "U"
            }}
          </div>
          <div class="flex-1 overflow-hidden">
            <p class="truncate text-[13px] font-medium text-gray-900">
              {{
                userData?.first_name || userData?.last_name
                  ? `${userData.first_name || ""} ${userData.last_name || ""}`.trim() || $t('navigation.user')
                  : $t('navigation.user')
              }}
            </p>
            <p class="truncate text-[11px] text-gray-500">{{ userData?.email || $t('navigation.loading') }}</p>
          </div>
        </div>
        <div v-else class="flex justify-center">
          <div
            class="flex h-8 w-8 items-center justify-center rounded-full bg-gray-200 text-xs font-medium text-gray-600"
          >
            {{
              userData?.first_name
                ? userData.first_name[0].toUpperCase()
                : userData?.email
                  ? userData.email[0].toUpperCase()
                  : "U"
            }}
          </div>
        </div>
        <div class="mt-2" :class="isCollapsed ? 'flex justify-center' : ''">
          <button
            type="button"
            @click="handleLogout"
            class="group relative flex w-full items-center justify-center gap-1.5 rounded-md border border-gray-200 px-2.5 py-1.5 text-[13px] font-medium text-gray-700 transition-colors hover:bg-gray-50"
            :class="isCollapsed ? 'h-8 w-8 p-0' : 'w-full'"
            :title="isCollapsed ? $t('navigation.exit') : ''"
          >
            <SvgIcon name="exit" class="h-3.5 w-3.5 flex-shrink-0" />
            <span v-show="!isCollapsed">{{ $t('navigation.exit') }}</span>
            <span
              v-if="isCollapsed"
              class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
            >
              {{ $t('navigation.exit') }}
            </span>
          </button>
        </div>
      </div>
    </aside>

    <!-- Settings Sidebar -->
    <aside class="flex h-screen flex-col border-e border-[#e7e2df] bg-white w-64">
      <!-- Settings Header -->
      <div class="flex h-14 items-center border-b border-gray-100 px-4">
        <h2 class="text-base font-semibold text-gray-900">
          {{ isAdminSettings ? $t('settings.adminSettings') : $t('navigation.settings') }}
        </h2>
      </div>

      <!-- Settings Navigation Menu -->
      <nav class="flex-1 overflow-y-auto px-2 py-3">
        <div class="space-y-1">
          <router-link
            v-for="tab in tabs"
            :key="tab.id"
            :to="{ name: tab.routeName }"
            :class="[
              'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
              isSettingsActive(tab.routeName)
                ? 'bg-gray-100 text-gray-900'
                : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
            ]"
          >
            <span>{{ tab.label }}</span>
          </router-link>
        </div>
      </nav>
    </aside>

    <!-- Main Content Area -->
    <main class="relative block flex-1 overflow-x-hidden overflow-y-auto px-6 py-6">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { logout } from "@/utils/auth";
import { apiGet, apiPost } from "@/services/api";
import SvgIcon from "@/components/SvgIcon.vue";
import Select from "@/components/ui/Select.vue";
import { Organization } from "@/models";

const { t } = useI18n();

const route = useRoute();
const isCollapsed = ref(false);

// User data
interface UserData {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: number;
}

const userData = ref<UserData | null>(null);

// Cache user role to prevent flickering
const cachedUserRole = ref<number | null>(null);

// Organizations data
const organizations = ref<Organization[]>([]);
const currentOrganizationId = ref<string>("");
const currentOrganizationName = ref<string>("");

// Check if user is admin (role === 3)
// Use cached role if userData is not loaded yet to prevent flickering
const isAdmin = computed(() => {
  const role = userData.value?.role ?? cachedUserRole.value;
  return role === 3;
});

// Check if current route is admin settings
const isAdminSettings = computed(() => route.path.startsWith('/admin/settings'));

// Organization settings tabs (available to all users)
const organizationTabs = computed(() => [
  { id: "general", label: t('settings.general'), routeName: "settings-general" },
  { id: "webhooks", label: t('settings.webhooks'), routeName: "settings-webhooks" },
  { id: "api_keys", label: t('settings.apiKeys'), routeName: "settings-api-keys" },
  { id: "branding", label: t('settings.branding'), routeName: "settings-branding" }
]);

// Admin/Global settings tabs (only for admins)
const adminTabs = computed(() => [
  { id: "smtp", label: `${t('settings.email')} (${t('settings.smtp')})`, routeName: "admin-settings-smtp" },
  { id: "sms", label: "SMS (Twilio)", routeName: "admin-settings-sms" },
  { id: "storage", label: t('settings.storage'), routeName: "admin-settings-storage" },
  { id: "geolocation", label: t('settings.geolocation'), routeName: "admin-settings-geolocation" },
  { id: "email_templates", label: t('settings.emailTemplates'), routeName: "admin-settings-email-templates" }
]);

// Active tabs based on current route
const tabs = computed(() => {
  if (isAdminSettings.value) {
    return adminTabs.value;
  }
  return organizationTabs.value;
});

// Load organizations
const loadOrganizations = async () => {
  try {
    const response = await apiGet("/api/v1/organizations");
    let data = response.data;
    if (data && typeof data === 'object' && 'organizations' in data) {
      data = data.organizations;
    }
    organizations.value = Array.isArray(data) ? data : [];
    
    // Update current organization from localStorage
    updateCurrentOrganization();
  } catch (error) {
    console.error("Failed to load organizations:", error);
    organizations.value = [];
  }
};

// Update current organization from localStorage
const updateCurrentOrganization = () => {
  const storedOrg = localStorage.getItem("current_organization");
  if (storedOrg) {
    try {
      const org = JSON.parse(storedOrg);
      currentOrganizationId.value = org.id || "";
      currentOrganizationName.value = org.name || "";
    } catch (e) {
      console.error("Failed to parse current organization:", e);
      currentOrganizationId.value = "";
      currentOrganizationName.value = "";
    }
  } else {
    currentOrganizationId.value = "";
    currentOrganizationName.value = "";
  }
};

// Handle organization change
const handleOrganizationChange = async (orgId: string) => {
  if (orgId === "") {
    // Exit organization
    try {
      const response = await apiPost("/api/v1/organizations/switch");
      localStorage.setItem("access_token", response.data.access_token);
      localStorage.setItem("refresh_token", response.data.refresh_token);
      localStorage.removeItem("current_organization");
      currentOrganizationId.value = "";
      currentOrganizationName.value = "";
      // Reload page to refresh data
      window.location.reload();
    } catch (error) {
      console.error("Failed to exit organization:", error);
    }
  } else {
    // Switch to organization
    const org = organizations.value.find((o) => o.id === orgId);
    if (org) {
      try {
        const response = await apiPost(`/api/v1/organizations/${orgId}/switch`);
        localStorage.setItem("access_token", response.data.access_token);
        localStorage.setItem("refresh_token", response.data.refresh_token);
        const orgData = {
          id: org.id,
          name: org.name,
          role: response.data.role
        };
        localStorage.setItem("current_organization", JSON.stringify(orgData));
        currentOrganizationId.value = orgId;
        currentOrganizationName.value = org.name;
        // Reload page to refresh data
        window.location.reload();
      } catch (error) {
        console.error("Failed to switch organization:", error);
      }
    }
  }
};

// Load current user data
const loadUserData = async () => {
  try {
    // Try to load cached role from localStorage first
    const cachedRole = localStorage.getItem("user_role");
    if (cachedRole) {
      cachedUserRole.value = parseInt(cachedRole, 10);
    }
    
    const response = await apiGet("/api/v1/users/me");
    if (response && response.data) {
      userData.value = response.data as UserData;
      // Cache role in localStorage and ref
      if (response.data.role !== undefined) {
        cachedUserRole.value = response.data.role;
        localStorage.setItem("user_role", String(response.data.role));
      }
    }
  } catch (error) {
    console.error("Failed to load user data:", error);
  }
};

// Watch for localStorage changes
const watchStorage = () => {
  window.addEventListener("storage", updateCurrentOrganization);
  // Also check periodically for same-tab updates
  setInterval(updateCurrentOrganization, 1000);
};

onMounted(() => {
  loadUserData();
  loadOrganizations();
  updateCurrentOrganization();
  watchStorage();
});

/**
 * Check if the given path is active
 */
function isActive(path: string): boolean {
  return route.path.startsWith(path);
}

/**
 * Check if the given settings route is active
 */
function isSettingsActive(routeName: string): boolean {
  return route.name === routeName;
}

/**
 * Toggle sidebar collapsed state
 */
function toggleSidebar(): void {
  isCollapsed.value = !isCollapsed.value;
}

/**
 * Handle logout - clear tokens and redirect to login
 */
async function handleLogout(): Promise<void> {
  await logout();
}
</script>

<style scoped>
/* Settings Sidebar styles */
</style>
