<template>
  <div class="field-input-wrapper">
    <FileDropZone
      :accept="accept"
      :disabled="disabled"
      :selected-label="selectedFileName"
      @change="handleFileChange"
      @clear="handleClear"
    />
    <div v-if="type === 'image' && modelValue" class="mt-2 rounded-md border border-[var(--color-base-300)] bg-[--color-base-100] p-2">
      <img :src="modelValue" alt="" class="max-h-32 w-full object-contain" />
    </div>
    <div v-if="error" class="mt-2 text-sm text-[var(--color-error)]">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import FileDropZone from "@/components/ui/FileDropZone.vue";

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
const selectedFileName = ref("");

const accept = props.type === "image" ? "image/*" : undefined;

watch(
  () => props.modelValue,
  (newValue) => {
    if (!newValue || newValue === "") {
      selectedFileName.value = "";
    }
  }
);

function handleFileChange(file: File): void {
  selectedFileName.value = file.name;

  if (props.type === "image") {
    const reader = new FileReader();
    reader.onload = (e) => {
      const result = e.target?.result as string;
      if (result) emit("update:modelValue", result);
      emit("blur");
    };
    reader.readAsDataURL(file);
  } else {
    emit("update:modelValue", file.name);
    emit("blur");
  }
}

function handleClear(): void {
  selectedFileName.value = "";
  emit("update:modelValue", "");
  emit("blur");
}
</script>
