<template>
  <div class="flex flex-wrap items-center justify-center gap-1 overflow-x-auto">
    <button
      v-for="(field, index) in fields"
      :key="field.id"
      type="button"
      class="field-dot relative h-2 w-2 shrink-0 cursor-pointer rounded-full focus:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-1"
      :class="dotClasses(field)"
      :title="getFieldLabel ? getFieldLabel(field) : (field.label || field.name || `Field ${index + 1}`)"
      :aria-label="getFieldLabel ? getFieldLabel(field) : (field.label || field.name || `Field ${index + 1}`)"
      :aria-current="currentFieldId === field.id ? 'true' : undefined"
      @click="onSelect(field.id)"
    />
  </div>
</template>

<script setup lang="ts">
import type { Field } from "@/models/template";

const props = defineProps<{
  fields: Field[];
  filledFieldIds: string[];
  currentFieldId: string | null;
  getFieldLabel?: (field: Field) => string;
}>();

const emit = defineEmits<{
  fieldSelect: [fieldId: string];
}>();

function dotClasses(field: Field): Record<string, boolean> {
  const filled = props.filledFieldIds.includes(field.id);
  const current = props.currentFieldId === field.id;
  return {
    "bg-neutral-400 hover:bg-neutral-500": !filled && !current,
    "bg-success": filled && !current,
    "bg-primary": current,
  };
}

function onSelect(fieldId: string): void {
  emit("fieldSelect", fieldId);
}
</script>
