<template>
  <label :class="labelClasses">
    <slot />
    <span v-if="required" class="ml-1 text-[var(--color-error)]">*</span>
  </label>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  required?: boolean;
}

withDefaults(defineProps<Props>(), {
  required: false
});

const labelClasses = computed(() => {
  return "block text-sm font-medium text-gray-700 mb-1";
});
</script>
