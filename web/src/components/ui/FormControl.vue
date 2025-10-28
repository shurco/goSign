<template>
  <div class="form-control">
    <Label v-if="label" :required="required">{{ label }}</Label>
    <slot />
    <p v-if="error" class="mt-1 text-sm text-[var(--color-error)]">{{ error }}</p>
    <p v-else-if="hint" class="mt-1 text-sm text-gray-500">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
import Label from "./Label.vue";

interface Props {
  label?: string;
  required?: boolean;
  error?: string;
  hint?: string;
}

withDefaults(defineProps<Props>(), {
  label: "",
  required: false,
  error: "",
  hint: ""
});
</script>
