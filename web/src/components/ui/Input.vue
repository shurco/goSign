<template>
  <input
    :class="inputClasses"
    :type="type"
    :value="modelValue"
    @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    v-bind="$attrs"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  modelValue?: string | number;
  type?: "text" | "email" | "password" | "number" | "date" | "tel" | "url" | "search" | "color";
  error?: boolean;
  size?: "sm" | "md" | "lg";
}

interface Emits {
  (e: "update:modelValue", value: string | number): void;
}

const props = withDefaults(defineProps<Props>(), {
  type: "text",
  error: false,
  size: "md"
});

const emit = defineEmits<Emits>();

const inputClasses = computed(() => {
  const base =
    "w-full rounded-md border border-gray-300 shadow-sm bg-[var(--color-base-100)] text-[var(--color-base-content)] transition-all duration-200 focus:outline-none focus:outline-offset-2 focus:ring-2 focus:ring-[var(--color-primary)]";

  const borderColor = props.error
    ? "border-[var(--color-error)] focus:border-[var(--color-error)] focus:ring-[var(--color-error)]"
    : "hover:border-[var(--color-base-content)]/20 focus:border-[var(--color-primary)]";

  const sizes = {
    sm: "h-8 px-3 text-xs min-h-[2rem]",
    md: "h-12 px-4 text-sm min-h-[3rem]",
    lg: "h-16 px-4 text-base min-h-[4rem]"
  };

  const disabled = "$attrs.disabled" in ["", true] ? "opacity-60 cursor-not-allowed bg-[var(--color-base-200)]" : "";

  return [base, borderColor, sizes[props.size], disabled].filter(Boolean).join(" ");
});
</script>
