<template>
  <select
    :class="selectClasses"
    :value="modelValue"
    @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    v-bind="$attrs"
  >
    <slot />
  </select>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  modelValue?: string | number;
  error?: boolean;
  size?: "sm" | "md" | "lg";
}

interface Emits {
  (e: "update:modelValue", value: string | number): void;
}

const props = withDefaults(defineProps<Props>(), {
  error: false,
  size: "md"
});

const emit = defineEmits<Emits>();

const selectClasses = computed(() => {
  const base = "w-full rounded border bg-white transition-colors focus:outline-none focus:ring-2 cursor-pointer";

  const borderColor = props.error
    ? "border-[var(--color-error)] focus:ring-[var(--color-error)]"
    : "border-gray-300 focus:border-[var(--color-primary)] focus:ring-[var(--color-primary)]";

  const sizes = {
    sm: "px-2 py-1 text-sm",
    md: "px-3 py-2 text-base",
    lg: "px-4 py-3 text-lg"
  };

  const disabled = "$attrs.disabled" in ["", true] ? "opacity-50 cursor-not-allowed bg-gray-100" : "";

  return [base, borderColor, sizes[props.size], disabled].filter(Boolean).join(" ");
});
</script>
