<template>
  <input type="checkbox" :class="switchClasses" v-bind="$attrs" />
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  size?: "xs" | "sm" | "md" | "lg";
}

const props = withDefaults(defineProps<Props>(), {
  size: "md"
});

const switchClasses = computed(() => {
  const base = "toggle";

  const sizes = {
    xs: "toggle-xs",
    sm: "toggle-sm",
    md: "toggle-md",
    lg: "toggle-lg"
  };

  return [base, sizes[props.size]].filter(Boolean).join(" ");
});
</script>
