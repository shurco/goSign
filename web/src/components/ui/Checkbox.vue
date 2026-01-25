<template>
  <input
    type="checkbox"
    :class="checkboxClasses"
    :checked="isChecked"
    v-bind="restAttrs"
    @change="onChange"
  />
</template>

<script setup lang="ts">
import { computed, useAttrs } from "vue";

interface Props {
  modelValue?: boolean | string[];
  value?: string;
  size?: "sm" | "md";
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  value: undefined,
  size: "md"
});

const emit = defineEmits<{ (e: "update:modelValue", value: boolean | string[]): void }>();

const attrs = useAttrs();
const restAttrs = computed(() => {
  const { modelValue: _mv, value: _v, ...rest } = attrs as Record<string, unknown>;
  return rest;
});

const isChecked = computed(() => {
  if (props.value !== undefined && props.value !== "") {
    const arr = Array.isArray(props.modelValue) ? props.modelValue : [];
    return arr.includes(props.value);
  }
  return props.modelValue === true;
});

function onChange(e: Event): void {
  const target = e.target as HTMLInputElement;
  if (props.value !== undefined && props.value !== "") {
    const arr = Array.isArray(props.modelValue) ? [...props.modelValue] : [];
    const i = arr.indexOf(props.value);
    if (target.checked) {
      if (i === -1) arr.push(props.value);
    } else {
      if (i !== -1) arr.splice(i, 1);
    }
    emit("update:modelValue", arr);
  } else {
    emit("update:modelValue", target.checked);
  }
}

const checkboxClasses = computed(() => {
  const base =
    "rounded border-gray-300 text-[var(--color-primary)] focus:ring-[var(--color-primary)] focus:ring-2 transition-colors cursor-pointer";

  const sizes = {
    sm: "h-4 w-4",
    md: "h-5 w-5"
  };

  const disabled = "$attrs.disabled" in ["", true] ? "opacity-50 cursor-not-allowed" : "";

  return [base, sizes[props.size], disabled].filter(Boolean).join(" ");
});
</script>
