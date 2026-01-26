<template>
  <!-- Calculated field (read-only) -->
  <TextInput
    v-if="isCalculated"
    :modelValue="formatCalculated(calculatedValue)"
    type="text"
    :placeholder="placeholder"
    :required="required"
    readonly
    :disabled="disabled"
    :error="error"
    class="calculated-field"
    @update:modelValue="() => {}"
    @blur="$emit('blur')"
  />
  <!-- Regular text input -->
  <TextInput
    v-else-if="isTextType"
    :modelValue="stringValue"
    :type="type"
    :placeholder="placeholder"
    :required="required"
    :readonly="readonly"
    :disabled="disabled"
    :error="error"
    @update:modelValue="handleUpdate"
    @blur="$emit('blur')"
  />
  <DateInput
    v-else-if="type === 'date'"
    :modelValue="stringValue"
    :placeholder="placeholder"
    :required="required"
    :readonly="readonly"
    :disabled="disabled"
    :error="error"
    @update:modelValue="handleUpdate"
    @blur="$emit('blur')"
  />
  <SelectInput
    v-else-if="isSelectType"
    :modelValue="selectModelValue"
    :type="type as any"
    :placeholder="placeholder"
    :required="required"
    :disabled="disabled"
    :options="options"
    :error="error"
    @update:modelValue="handleUpdate"
    @blur="$emit('blur')"
  />
  <FileInput
    v-else-if="type === 'file' || type === 'image'"
    :modelValue="stringValue"
    :type="type"
    :placeholder="placeholder"
    :required="required"
    :readonly="readonly"
    :disabled="disabled"
    :error="error"
    @update:modelValue="handleUpdate"
    @blur="$emit('blur')"
  />
  <SignatureInput
    v-else-if="isSignatureType || type === 'stamp'"
    :modelValue="stringValue"
    :mode="type === 'initials' ? 'initials' : type === 'stamp' ? 'stamp' : 'signature'"
    :format="signatureFormat"
    :placeholder="placeholder"
    :required="required"
    :disabled="disabled || (type === 'stamp' && readonly)"
    :error="error"
    @update:modelValue="handleUpdate"
    @blur="$emit('blur')"
  />
  <CellsInput
    v-else-if="type === 'cells'"
    :modelValue="stringValue"
    :cell-count="cellCount"
    :placeholder="placeholder"
    :required="required"
    :readonly="readonly"
    :disabled="disabled"
    :error="error"
    @update:modelValue="handleUpdate"
    @blur="$emit('blur')"
  />
  <div
    v-else-if="type === 'payment'"
    class="field-input-wrapper"
  >
    <div class="text-lg font-semibold">
      {{ formatPaymentPrice(price, currency) }}
    </div>
    <div class="text-sm text-gray-500 mt-1">
      {{ placeholder || 'Payment required' }}
    </div>
    <div v-if="error" class="mt-1 text-sm text-[var(--color-error)]">{{ error }}</div>
  </div>
  <div v-else class="field-input-wrapper">
    <div class="text-sm text-gray-500">Field type "{{ type }}" not yet implemented</div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import TextInput from "@/components/field/inputs/TextInput.vue";
import DateInput from "@/components/field/inputs/DateInput.vue";
import SelectInput from "@/components/field/inputs/SelectInput.vue";
import FileInput from "@/components/field/inputs/FileInput.vue";
import SignatureInput from "@/components/field/inputs/SignatureInput.vue";
import CellsInput from "@/components/field/inputs/CellsInput.vue";

interface Option {
  id?: string;
  value?: string;
  label?: string;
}

interface Props {
  type: string;
  modelValue?: string | boolean | string[];
  placeholder?: string;
  label?: string;
  required?: boolean;
  readonly?: boolean;
  disabled?: boolean;
  options?: Option[];
  error?: string;
  formula?: string;
  calculationType?: 'number' | 'currency';
  calculatedValue?: number;
  cellCount?: number;
  price?: number;
  currency?: string;
  dateFormat?: string;
  /** Signature field format: '', drawn, typed, drawn_or_typed, drawn_or_upload, upload */
  signatureFormat?: string;
}

interface Emits {
  (e: "update:modelValue", value: string | boolean | string[]): void;
  (e: "blur"): void;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: "",
  placeholder: "",
  label: "",
  required: false,
  readonly: false,
  disabled: false,
  options: () => [],
  error: "",
  cellCount: 6,
  price: 0,
  currency: "USD",
  dateFormat: undefined,
  signatureFormat: ""
});

const emit = defineEmits<Emits>();
const localValue = ref(props.modelValue);

const stringValue = computed(() =>
  typeof localValue.value === "string" ? localValue.value : ""
);

const selectModelValue = computed(() => {
  if (props.type === "checkbox") {
    const v = localValue.value;
    return v === true || v === "true";
  }
  if (props.type === "multiple" || props.type === "multi_select") {
    return Array.isArray(localValue.value) ? localValue.value : [];
  }
  if (props.type === "radio") {
    const v = localValue.value;
    return v != null && v !== "" ? String(v) : "";
  }
  return stringValue.value;
});

const isTextType = computed(() => {
  return ["text", "number", "phone"].includes(props.type);
});

const isSelectType = computed(() => {
  return ["select", "radio", "checkbox", "multiple", "multi_select"].includes(props.type);
});

const isCalculated = computed(() => {
  return !!props.formula;
});

const isSignatureType = computed(() => {
  return ["signature", "initials"].includes(props.type);
});

function formatCalculated(value: number | undefined): string {
  if (value === undefined || value === null) return '';

  if (props.calculationType === 'currency') {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(value);
  }

  return value.toFixed(2);
}

function formatPaymentPrice(price: number | undefined, currency: string | undefined): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency || 'USD'
  }).format(price ?? 0);
}

watch(
  () => props.modelValue,
  (newValue) => {
    localValue.value = newValue;
  }
);

// Watch calculatedValue to update display for calculated fields
// Note: For calculated fields, the value is displayed directly from calculatedValue prop
// No need to update localValue as it's read-only

function handleUpdate(value: string | boolean | string[]): void {
  localValue.value = value;
  emit("update:modelValue", value);
}
</script>
