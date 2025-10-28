<template>
  <div :class="alertClasses" role="alert">
    <div v-if="$slots.icon" class="flex-shrink-0">
      <slot name="icon" />
    </div>
    <div class="flex-1">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  variant?: "info" | "success" | "warning" | "error";
}

const props = withDefaults(defineProps<Props>(), {
  variant: "info"
});

const alertClasses = computed(() => {
  const base = "flex items-start gap-4 p-4 rounded-xl border-l-4";

  const variants = {
    info: "bg-[var(--color-info)]/10 border-[var(--color-info)] text-[var(--color-info-content)]",
    success: "bg-[var(--color-success)]/10 border-[var(--color-success)] text-[var(--color-success-content)]",
    warning: "bg-[var(--color-warning)]/10 border-[var(--color-warning)] text-[var(--color-warning-content)]",
    error: "bg-[var(--color-error)]/10 border-[var(--color-error)] text-[var(--color-error-content)]"
  };

  return [base, variants[props.variant]].join(" ");
});
</script>
