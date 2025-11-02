<template>
  <component :is="layoutComponent" v-if="layoutComponent" :key="`${route.path}-${route.name}`">
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
import { computed, ref, watch } from "vue";

const route = useRoute();
const layoutReady = ref(false);

// Watch for route changes and reset layoutReady if component is missing
watch(
  () => route.path,
  () => {
    if (!route.meta.layoutComponent) {
      layoutReady.value = false;
    } else {
      layoutReady.value = true;
    }
  },
  { immediate: true }
);

// Ensure layoutComponent is always available or fallback to Blank
const layoutComponent = computed(() => {
  return route.meta.layoutComponent || null;
});
</script>
