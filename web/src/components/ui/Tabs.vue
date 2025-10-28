<template>
  <div class="tabs-wrapper">
    <div class="flex gap-2 rounded-lg bg-gray-100 p-1">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { provide, ref } from "vue";

interface Props {
  modelValue?: string;
}

interface Emits {
  (e: "update:modelValue", value: string): void;
}

const props = defineProps<Props>();
const emit = defineEmits<Emits>();

const activeTab = ref(props.modelValue || "");

function setActiveTab(value: string): void {
  activeTab.value = value;
  emit("update:modelValue", value);
}

provide("activeTab", activeTab);
provide("setActiveTab", setActiveTab);
</script>
