<template>
  <div class="field-input-wrapper">
    <FileInput
      ref="fileInputRef"
      :accept="accept"
      :required="required"
      :disabled="disabled"
      @change="handleFileChange"
      @blur="handleBlur"
    />
    <div v-if="selectedFileName" class="mt-2 text-sm text-gray-600">
      {{ selectedFileName }}
    </div>
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
const fileInputRef = ref<HTMLInputElement | null>(null);
const selectedFileName = ref("");

const accept = props.type === "image" ? "image/*" : undefined;

watch(
  () => props.modelValue,
  (newValue) => {
    if (!newValue || newValue === "") {
      selectedFileName.value = "";
      if (fileInputRef.value) {
        (fileInputRef.value as HTMLInputElement).value = "";
      }
    }
  }
);

function handleFileChange(event: Event): void {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  
  if (file) {
    selectedFileName.value = file.name;
    
    // For image type, convert to base64
    if (props.type === "image") {
      const reader = new FileReader();
      reader.onload = (e) => {
        const result = e.target?.result as string;
        if (result) {
          emit("update:modelValue", result);
        }
      };
      reader.readAsDataURL(file);
    } else {
      // For file type, just store the filename as indicator that file was selected
      emit("update:modelValue", file.name);
    }
  } else {
    selectedFileName.value = "";
    emit("update:modelValue", "");
  }
}

function handleBlur(): void {
  emit("blur");
}
</script>

