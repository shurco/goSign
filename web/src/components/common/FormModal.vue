<template>
  <Modal
    v-model="isOpen"
    :title="title"
    :size="size"
    :show-close="closeButton"
    :close-on-backdrop="closeOnOutsideClick"
    :close-on-escape="closeOnEscape"
    @close="handleClose"
  >
    <template #default>
      <div class="py-4">
        <slot :form-data="formData" :errors="errors" />
      </div>
    </template>

    <template #footer>
      <slot name="footer" :submit="handleSubmit" :cancel="handleClose" :is-submitting="isSubmitting">
        <div class="flex justify-end gap-3">
          <Button v-if="showCancel" variant="ghost" :disabled="isSubmitting" @click="handleClose">
            {{ cancelText }}
          </Button>
          <Button variant="primary" :loading="isSubmitting" :disabled="isSubmitting || !isValid" @click="handleSubmit">
            {{ submitText }}
          </Button>
        </div>
      </slot>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import Modal from "@/components/ui/Modal.vue";
import Button from "@/components/ui/Button.vue";

interface Props {
  modelValue: boolean;
  title: string;
  size?: "sm" | "md" | "lg" | "xl";
  submitText?: string;
  cancelText?: string;
  showCancel?: boolean;
  closeButton?: boolean;
  closeOnOutsideClick?: boolean;
  closeOnEscape?: boolean;
  validateOnMount?: boolean;
}

interface Emits {
  (e: "update:modelValue", value: boolean): void;
  (e: "submit", formData: Record<string, unknown>): void;
  (e: "cancel" | "close"): void;
}

const props = withDefaults(defineProps<Props>(), {
  size: "md",
  submitText: "Submit",
  cancelText: "Cancel",
  showCancel: true,
  closeButton: true,
  closeOnOutsideClick: true,
  closeOnEscape: true,
  validateOnMount: false
});

const emit = defineEmits<Emits>();

const formData = ref<Record<string, unknown>>({});
const errors = ref<Record<string, string>>({});
const isSubmitting = ref(false);

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value)
});

const isValid = computed(() => {
  return Object.keys(errors.value).length === 0;
});

watch(isOpen, (newValue) => {
  if (newValue && props.validateOnMount) {
    validateForm();
  }
  // Initialize formData when modal opens
  if (newValue) {
    // Ensure formData has expected structure based on common fields
    if (!formData.value.name) {
      formData.value.name = "";
    }
  }
});

function handleClose(): void {
  if (!isSubmitting.value) {
    isOpen.value = false;
    emit("close");
    emit("cancel");
    resetForm();
  }
}

async function handleSubmit(): Promise<void> {
  if (isSubmitting.value || !isValid.value) {
    return;
  }

  validateForm();

  if (!isValid.value) {
    return;
  }

  isSubmitting.value = true;
  try {
    emit("submit", formData.value);
  } finally {
    isSubmitting.value = false;
  }
}

function validateForm(): void {
  errors.value = {};
}

function resetForm(): void {
  formData.value = {};
  errors.value = {};
}

function setFormData(data: Record<string, unknown>): void {
  formData.value = { ...data };
}

function setError(field: string, message: string): void {
  errors.value[field] = message;
}

function clearError(field: string): void {
  errors.value[field] = undefined;
}

function clearAllErrors(): void {
  errors.value = {};
}

defineExpose({
  open: () => {
    isOpen.value = true;
  },
  close: handleClose,
  setFormData,
  getFormData: () => formData.value,
  setError,
  clearError,
  clearAllErrors,
  resetForm,
  validateForm,
  isValid: () => isValid.value
});
</script>
