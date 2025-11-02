<template>
  <div class="flex gap-2">
    <button
      v-for="option in normalizedOptions"
      :key="getOptionValue(option)"
      class="flex-1 rounded-md border px-3 py-2 text-sm font-medium transition-colors"
      :class="[
        isSelected(option)
          ? 'border-blue-500 bg-blue-50 text-blue-700'
          : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
      ]"
      :disabled="disabled"
      @click="handleClick(option)"
    >
      <div class="flex items-center justify-center gap-2">
        <SvgIcon v-if="getOptionIcon(option)" :name="getOptionIcon(option)" width="16" height="16" />
        <span>{{ getOptionLabel(option) }}</span>
      </div>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import SvgIcon from "@/components/SvgIcon.vue";

export interface ButtonGroupOption {
  value: string | number;
  label: string;
  icon?: string;
}

interface Props {
  modelValue: string | number;
  options: ButtonGroupOption[] | string[] | Array<{ value: any; label: string; icon?: string }>;
  disabled?: boolean;
}

interface Emits {
  (e: "update:modelValue", value: string | number): void;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false
});

const emit = defineEmits<Emits>();

// Normalize options to handle different input formats
const normalizedOptions = computed(() => {
  return props.options.map((option) => {
    if (typeof option === "string") {
      return { value: option, label: option, icon: undefined };
    }
    return option;
  });
});

function getOptionValue(option: any): string | number {
  if (typeof option === "string") {
    return option;
  }
  return option.value;
}

function getOptionLabel(option: any): string {
  if (typeof option === "string") {
    return option;
  }
  return option.label || String(option.value);
}

function getOptionIcon(option: any): string | undefined {
  if (typeof option === "string") {
    return undefined;
  }
  return option.icon;
}

function isSelected(option: any): boolean {
  return getOptionValue(option) === props.modelValue;
}

function handleClick(option: any): void {
  if (props.disabled) {
    return;
  }
  emit("update:modelValue", getOptionValue(option));
}
</script>
