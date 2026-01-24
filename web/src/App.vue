<template>
  <!--
    Important: do NOT key the layout by route path/name.
    Keying by path forces a full layout remount on every navigation, which causes
    sidebar header/user blocks to flicker and re-fetch `/api/v1/users/me`.
    We only want to remount when the layout type changes (Blank/Main/Sidebar/...).
  -->
  <component :is="layoutComponent" v-if="layoutComponent" :key="layoutKey">
    <RouterView :key="route.path" />
  </component>
  <div v-else class="flex h-screen items-center justify-center">
    <div class="text-center">
      <div class="mx-auto mb-4 h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-blue-600"></div>
      <p class="text-gray-600">Loading...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterView, useRoute } from "vue-router";
import { computed } from "vue";

const route = useRoute();

// Ensure layoutComponent is always available or fallback to Blank
const layoutComponent = computed(() => {
  return route.meta.layoutComponent || null;
});

const layoutKey = computed(() => {
  // Vue will only remount the layout when this value changes.
  const key = route.meta.layout as string | undefined;
  return key || "Blank";
});
</script>
