<template>
  <button :class="tabClasses" @click="handleClick">
    <slot />
  </button>
</template>

<script setup lang="ts">
import { computed, inject, type Ref } from "vue";

interface Props {
  value: string;
}

const props = defineProps<Props>();

const activeTab = inject<Ref<string>>("activeTab");
const setActiveTab = inject<(value: string) => void>("setActiveTab");

const isActive = computed(() => activeTab?.value === props.value);

const tabClasses = computed(() => {
  const base = "px-4 py-2 rounded-md text-sm font-medium transition-colors";
  const active = isActive.value ? "bg-white text-gray-900 shadow-sm" : "text-gray-600 hover:text-gray-900";

  return [base, active].join(" ");
});

function handleClick(): void {
  setActiveTab?.(props.value);
}
</script>
