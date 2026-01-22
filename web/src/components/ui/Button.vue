<template>
  <button :class="buttonClasses" :disabled="disabled || loading" v-bind="$attrs">
    <LoadingSpinner v-if="loading" :size="spinnerSize" class="mr-2" />
    <slot />
  </button>
</template>

<script setup lang="ts">
import { computed } from "vue";
import LoadingSpinner from "./LoadingSpinner.vue";

interface Props {
  variant?: "primary" | "ghost" | "success" | "warning" | "error" | "info";
  size?: "sm" | "md" | "lg";
  loading?: boolean;
  disabled?: boolean;
  circle?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  variant: "primary",
  size: "md",
  loading: false,
  disabled: false,
  circle: false
});

const buttonClasses = computed(() => {
  const base =
    "inline-flex items-center justify-center gap-2 font-medium cursor-pointer select-none text-center transition-colors focus:outline-none no-underline normal-case";

  const variants = {
    primary: "rounded-md border border-blue-500 bg-blue-50 text-blue-700 hover:bg-blue-100",
    ghost: "rounded-md border border-gray-300 bg-white text-gray-700 hover:bg-gray-50",
    success: "rounded-md border border-green-500 bg-green-50 text-green-700 hover:bg-green-100",
    warning: "rounded-md border border-yellow-500 bg-yellow-50 text-yellow-700 hover:bg-yellow-100",
    error: "rounded-md border border-red-500 bg-red-50 text-red-700 hover:bg-red-100",
    info: "rounded-md border border-blue-500 bg-blue-50 text-blue-700 hover:bg-blue-100"
  };

  const sizes = props.circle
    ? {
        sm: "h-8 w-8 rounded-full",
        md: "h-12 w-12 rounded-full",
        lg: "h-16 w-16 rounded-full"
      }
    : {
        sm: "px-3 py-2 text-xs rounded-md",
        md: "px-4 py-2 text-sm rounded-md",
        lg: "px-6 py-3 text-base rounded-md"
      };

  const disabled = props.disabled || props.loading ? "opacity-60 cursor-not-allowed pointer-events-none" : "";

  return [base, variants[props.variant], sizes[props.size], disabled].filter(Boolean).join(" ");
});

const spinnerSize = computed(() => {
  const sizeMap = {
    sm: "sm",
    md: "sm",
    lg: "md"
  } as const;
  return sizeMap[props.size];
});
</script>
