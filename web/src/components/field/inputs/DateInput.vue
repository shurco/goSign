<template>
  <div class="field-input-wrapper">
    <Input
      v-model="localValue"
      type="date"
      :placeholder="placeholder"
      :required="required"
      :readonly="readonly"
      :disabled="disabled"
      @blur="handleBlur"
    />
    <div v-if="error" class="mt-1 text-sm text-[var(--color-error)]">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import Input from "@/components/ui/Input.vue";

interface Props {
  modelValue?: string;
  placeholder?: string;
  required?: boolean;
  readonly?: boolean;
  disabled?: boolean;
  error?: string;
}

interface Emits {
  (e: "update:modelValue", value: string): void;
  (e: "blur"): void;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: "",
  placeholder: "",
  required: false,
  readonly: false,
  disabled: false,
  error: ""
});

const emit = defineEmits<Emits>();
const localValue = ref(props.modelValue);

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

