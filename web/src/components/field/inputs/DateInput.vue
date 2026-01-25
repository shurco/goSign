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
    <div v-if="dateFormat && formattedDisplay" class="mt-1 text-sm text-[--color-base-content]/70">
      {{ formattedDisplay }}
    </div>
    <div v-if="error" class="mt-1 text-sm text-[var(--color-error)]">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import Input from "@/components/ui/Input.vue";
import { formatDateByPattern } from "@/utils/time";

interface Props {
  modelValue?: string;
  dateFormat?: string;
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
  dateFormat: "",
  placeholder: "",
  required: false,
  readonly: false,
  disabled: false,
  error: ""
});

const emit = defineEmits<Emits>();
const localValue = ref(props.modelValue);

const formattedDisplay = computed(() => {
  if (!props.dateFormat || !localValue.value) return "";
  return formatDateByPattern(localValue.value, props.dateFormat);
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

