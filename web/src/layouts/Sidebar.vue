<template>
  <div class="relative flex h-screen overflow-hidden bg-[var(--color-base-100)]">
    <!-- Sidebar Navigation -->
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
          <li>
            <RouterLink
              to="/dashboard"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[{ 'bg-gray-100 text-gray-900': isActive('/dashboard') }, isCollapsed ? 'justify-center' : '']"
              :title="isCollapsed ? 'Dashboard' : ''"
            >
              <SvgIcon name="dashboard" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">Dashboard</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                Dashboard
              </span>
            </RouterLink>
          </li>

          <li>
            <RouterLink
              to="/submissions"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[{ 'bg-gray-100 text-gray-900': isActive('/submissions') }, isCollapsed ? 'justify-center' : '']"
              :title="isCollapsed ? 'Submissions' : ''"
            >
              <SvgIcon name="submissions" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">Submissions</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                Submissions
              </span>
            </RouterLink>
          </li>

          <li>
            <RouterLink
              to="/templates"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[{ 'bg-gray-100 text-gray-900': isActive('/templates') }, isCollapsed ? 'justify-center' : '']"
              :title="isCollapsed ? 'Templates' : ''"
            >
              <SvgIcon name="templates" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">Templates</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                Templates
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
              :title="isCollapsed ? 'Organizations' : ''"
            >
              <SvgIcon name="organizations" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">Organizations</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                Organizations
              </span>
            </RouterLink>
          </li>

          <li>
            <RouterLink
              to="/uploads"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[{ 'bg-gray-100 text-gray-900': isActive('/uploads') }, isCollapsed ? 'justify-center' : '']"
              :title="isCollapsed ? 'Uploads' : ''"
            >
              <SvgIcon name="uploads" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">Uploads</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                Uploads
              </span>
            </RouterLink>
          </li>

          <li class="pt-3">
            <div
              v-show="!isCollapsed"
              class="mb-1.5 px-2.5 text-[11px] font-semibold tracking-wider text-gray-400 uppercase"
            >
              System
            </div>
            <div v-if="isCollapsed" class="mb-1.5 border-t border-gray-200"></div>
            <RouterLink
              to="/settings"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[{ 'bg-gray-100 text-gray-900': isActive('/settings') }, isCollapsed ? 'justify-center' : '']"
              :title="isCollapsed ? 'Settings' : ''"
            >
              <SvgIcon name="settings" class="h-4 w-4 flex-shrink-0" />
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">Settings</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                Settings
              </span>
            </RouterLink>
          </li>
        </ul>
      </nav>

      <!-- Toggle Button -->
      <div class="flex justify-center border-t border-gray-100 py-2">
        <button
          class="group relative flex h-7 w-7 items-center justify-center rounded-md text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
          :title="isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'"
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
            Expand
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
                  ? `${userData.first_name || ""} ${userData.last_name || ""}`.trim() || "User"
                  : "User"
              }}
            </p>
            <p class="truncate text-[11px] text-gray-500">{{ userData?.email || "Loading..." }}</p>
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
            :title="isCollapsed ? 'Exit' : ''"
          >
            <SvgIcon name="exit" class="h-3.5 w-3.5 flex-shrink-0" />
            <span v-show="!isCollapsed">Exit</span>
            <span
              v-if="isCollapsed"
              class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
            >
              Exit
            </span>
          </button>
        </div>
      </div>
    </aside>

    <!-- Main Content Area -->
    <main class="relative block flex-1 overflow-x-hidden overflow-y-auto px-6 py-6">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { logout } from "@/utils/auth";
import { apiGet } from "@/services/api";
import SvgIcon from "@/components/SvgIcon.vue";

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

// Load current user data
const loadUserData = async () => {
  try {
    const response = await apiGet("/api/v1/users/me");
    if (response && response.data) {
      userData.value = response.data as UserData;
    }
  } catch (error) {
    console.error("Failed to load user data:", error);
  }
};

onMounted(() => {
  loadUserData();
});

/**
 * Check if the given path is active
 */
function isActive(path: string): boolean {
  return route.path.startsWith(path);
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
/* Sidebar styles */
</style>
