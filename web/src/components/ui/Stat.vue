<template>
  <div class="stat-item p-6">
    <div v-if="$slots.figure" class="stat-figure mb-2">
      <slot name="figure" />
    </div>
    <div v-if="title" class="stat-title mb-1 text-sm text-gray-600">{{ title }}</div>
    <div v-if="value" class="stat-value mb-1 text-3xl font-bold" :class="valueClass">{{ value }}</div>
    <div v-if="description" class="stat-desc text-sm text-gray-500">{{ description }}</div>
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  title?: string;
  value?: string | number;
  description?: string;
  variant?: "primary" | "success" | "warning" | "error" | "info";
}

const props = withDefaults(defineProps<Props>(), {
  title: "",
  value: "",
  description: "",
  variant: "primary"
});

const valueClass = computed(() => {
  const variants = {
    primary: "text-[var(--color-primary)]",
    success: "text-[var(--color-success)]",
    warning: "text-[var(--color-warning)]",
    error: "text-[var(--color-error)]",
    info: "text-[var(--color-info)]"
  };

  return variants[props.variant];
});
</script>
