<template>
  <input type="radio" :class="radioClasses" v-bind="$attrs" />
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  size?: "sm" | "md";
}

const props = withDefaults(defineProps<Props>(), {
  size: "md"
});

const radioClasses = computed(() => {
  const base =
    "border-gray-300 text-[var(--color-primary)] focus:ring-[var(--color-primary)] focus:ring-2 transition-colors cursor-pointer";

  const sizes = {
    sm: "h-4 w-4",
    md: "h-5 w-5"
  };

  const disabled = "$attrs.disabled" in ["", true] ? "opacity-50 cursor-not-allowed" : "";

  return [base, sizes[props.size], disabled].filter(Boolean).join(" ");
});
</script>
