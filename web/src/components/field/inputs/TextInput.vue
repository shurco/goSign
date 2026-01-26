<template>
  <div class="field-input-wrapper">
    <Input
      v-model="localValue"
      :type="inputType"
      :placeholder="placeholder"
      :required="required"
      :readonly="readonly"
      :disabled="disabled"
      :min="min"
      :max="max"
      :step="step"
      @blur="handleBlur"
    />
    <div v-if="error" class="mt-1 text-sm text-[var(--color-error)]">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import Input from "@/components/ui/Input.vue";

interface Props {
  modelValue?: string;
  type?: string;
  placeholder?: string;
  required?: boolean;
  readonly?: boolean;
  disabled?: boolean;
  error?: string;
  min?: number | string;
  max?: number | string;
  step?: string;
}

interface Emits {
  (e: "update:modelValue", value: string): void;
  (e: "blur"): void;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: "",
  type: "text",
  placeholder: "",
  required: false,
  readonly: false,
  disabled: false,
  error: "",
  min: undefined,
  max: undefined,
  step: undefined
});

const emit = defineEmits<Emits>();
const localValue = ref(props.modelValue);

const inputType = computed(() => {
  if (props.type === "number") return "number";
  if (props.type === "phone") return "tel";
  return "text";
});

watch(
  () => props.modelValue,
  (newValue) => {
    localValue.value = newValue || "";
  }
);

watch(localValue, (newValue) => {
  emit("update:modelValue", newValue || "");
});

function handleBlur(): void {
  emit("blur");
}
</script>

