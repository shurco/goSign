<template>
  <div class="field-input-wrapper">
    <FileInput
      v-model="localValue"
      :accept="accept"
      :required="required"
      :disabled="disabled"
      @blur="handleBlur"
    />
    <div v-if="error" class="mt-1 text-sm text-[var(--color-error)]">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import FileInput from "@/components/ui/FileInput.vue";

interface Props {
  modelValue?: string;
  type?: "file" | "image";
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
  type: "file",
  placeholder: "",
  required: false,
  readonly: false,
  disabled: false,
  error: ""
});

const emit = defineEmits<Emits>();
const localValue = ref(props.modelValue);

const accept = props.type === "image" ? "image/*" : undefined;

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

