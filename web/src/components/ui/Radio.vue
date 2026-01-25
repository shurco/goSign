<template>
  <input
    type="radio"
    :class="radioClasses"
    :name="name"
    :checked="isChecked"
    :value="value"
    v-bind="restAttrs"
    @change="onChange"
  />
</template>

<script setup lang="ts">
import { computed, useAttrs } from "vue";

interface Props {
  modelValue?: string;
  value?: string;
  name?: string;
  size?: "sm" | "md";
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: "",
  value: "",
  name: "",
  size: "md"
});

const emit = defineEmits<{ (e: "update:modelValue", value: string): void }>();

const attrs = useAttrs();
const restAttrs = computed(() => {
  const { modelValue: _mv, value: _v, name: _n, ...rest } = attrs as Record<string, unknown>;
  return rest;
});

const isChecked = computed(() => {
  const current = props.modelValue != null ? String(props.modelValue) : "";
  const opt = props.value != null ? String(props.value) : "";
  return current === opt;
});

function onChange(e: Event): void {
  const target = e.target as HTMLInputElement;
  if (target.checked) {
    emit("update:modelValue", props.value != null ? String(props.value) : "");
  }
}

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
