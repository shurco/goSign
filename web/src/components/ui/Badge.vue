<template>
  <span :class="badgeClasses">
    <slot />
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  variant?: "ghost" | "primary" | "success" | "warning" | "error" | "info";
  size?: "sm" | "md";
}

const props = withDefaults(defineProps<Props>(), {
  variant: "ghost",
  size: "md"
});

const badgeClasses = computed(() => {
  const base = "inline-flex items-center justify-center font-semibold border-0";

  const variants = {
    ghost: "bg-[var(--color-base-200)] text-[var(--color-base-content)]",
    primary: "bg-[var(--color-primary)] text-[var(--color-primary-content)]",
    success: "bg-[var(--color-success)] text-[var(--color-success-content)]",
    warning: "bg-[var(--color-warning)] text-[var(--color-warning-content)]",
    error: "bg-[var(--color-error)] text-[var(--color-error-content)]",
    info: "bg-[var(--color-info)] text-[var(--color-info-content)]"
  };

  const sizes = {
    sm: "h-5 px-2 text-xs rounded-full",
    md: "h-6 px-3 text-sm rounded-full"
  };

  return [base, variants[props.variant], sizes[props.size]].join(" ");
});
</script>
