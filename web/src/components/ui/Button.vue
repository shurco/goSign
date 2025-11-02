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
    "inline-flex items-center justify-center gap-2 font-semibold border-0 cursor-pointer select-none text-center transition-all duration-200 focus:outline-none active:scale-95 no-underline normal-case";

  const variants = {
    primary: "bg-[var(--color-primary)] text-[var(--color-primary-content)] hover:bg-[var(--color-primary-focus)]",
    ghost: "bg-transparent hover:bg-[var(--color-base-200)] text-[var(--color-base-content)] border border-transparent",
    success: "bg-[var(--color-success)] text-[var(--color-success-content)] hover:brightness-110",
    warning: "bg-[var(--color-warning)] text-[var(--color-warning-content)] hover:brightness-110",
    error: "bg-[var(--color-error)] text-[var(--color-error-content)] hover:brightness-110",
    info: "bg-[var(--color-info)] text-[var(--color-info-content)] hover:brightness-110"
  };

  const sizes = props.circle
    ? {
        sm: "h-8 w-8 rounded-full",
        md: "h-12 w-12 rounded-full",
        lg: "h-16 w-16 rounded-full"
      }
    : {
        sm: "h-8 px-3 text-xs rounded-lg min-h-[2rem]",
        md: "h-12 px-4 text-sm rounded-lg min-h-[3rem]",
        lg: "h-16 px-6 text-base rounded-lg min-h-[4rem]"
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
