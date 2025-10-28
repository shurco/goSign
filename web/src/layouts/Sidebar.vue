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
              <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-4 w-4 flex-shrink-0"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
                />
              </svg>
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
              <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-4 w-4 flex-shrink-0"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
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
              to="/edit"
              class="group relative flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
              :class="[{ 'bg-gray-100 text-gray-900': isActive('/edit') }, isCollapsed ? 'justify-center' : '']"
              :title="isCollapsed ? 'Editor' : ''"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-4 w-4 flex-shrink-0"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                />
              </svg>
              <span v-show="!isCollapsed" class="text-[13px] whitespace-nowrap">Editor</span>
              <span
                v-if="isCollapsed"
                class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
              >
                Editor
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
              <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-4 w-4 flex-shrink-0"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                />
              </svg>
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
              <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-4 w-4 flex-shrink-0"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                />
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                />
              </svg>
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
          @click="toggleSidebar"
          :title="isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-3.5 w-3.5 transition-transform duration-150"
            :class="isCollapsed ? 'rotate-180' : ''"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
          </svg>
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
            U
          </div>
          <div class="flex-1 overflow-hidden">
            <p class="truncate text-[13px] font-medium text-gray-900">User</p>
            <p class="truncate text-[11px] text-gray-500">user@example.com</p>
          </div>
        </div>
        <div v-else class="flex justify-center">
          <div
            class="flex h-8 w-8 items-center justify-center rounded-full bg-gray-200 text-xs font-medium text-gray-600"
          >
            U
          </div>
        </div>
        <div class="mt-2" :class="isCollapsed ? 'flex justify-center' : ''">
          <a
            href="/"
            class="group relative flex items-center justify-center gap-1.5 rounded-md border border-gray-200 px-2.5 py-1.5 text-[13px] font-medium text-gray-700 transition-colors hover:bg-gray-50"
            :class="isCollapsed ? 'h-8 w-8 p-0' : 'w-full'"
            :title="isCollapsed ? 'Exit' : ''"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-3.5 w-3.5 flex-shrink-0"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
              />
            </svg>
            <span v-show="!isCollapsed">Exit</span>
            <span
              v-if="isCollapsed"
              class="invisible absolute left-full ml-2 rounded-md bg-gray-900 px-2 py-1 text-xs text-white opacity-0 transition-all group-hover:visible group-hover:opacity-100"
            >
              Exit
            </span>
          </a>
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
import { ref } from "vue";
import { useRoute } from "vue-router";
import SvgIcon from "@/components/SvgIcon.vue";

const route = useRoute();
const isCollapsed = ref(false);

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
</script>

<style scoped>
/* Sidebar styles */
</style>
