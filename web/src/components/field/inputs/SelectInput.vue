<template>
  <div class="field-input-wrapper">
    <Select
      v-if="type === 'select'"
      :model-value="typeof localValue === 'string' ? localValue : String(localValue || '')"
      @update:model-value="(val) => { localValue = String(val); }"
      :required="required"
      :disabled="disabled"
      @blur="handleBlur"
    >
      <option value="">{{ placeholder || "Select..." }}</option>
      <option v-for="option in normalizedOptions" :key="option.id || option.value" :value="option.value || option.id">
        {{ option.label || option.value }}
      </option>
    </Select>

    <div v-else-if="type === 'radio'" class="space-y-2">
      <label
        v-for="option in normalizedOptions"
        :key="option.id || option.value"
        class="flex cursor-pointer items-center gap-2"
      >
        <Radio
          v-model="localValue"
          :name="radioGroupName"
          :value="option.value ?? option.id ?? ''"
          :disabled="disabled"
        />
        <span>{{ option.label || option.value }}</span>
      </label>
    </div>

    <div v-else-if="type === 'checkbox'" class="flex items-center gap-2">
      <Checkbox v-model="localValue" :disabled="disabled" @blur="handleBlur" />
      <span v-if="placeholder">{{ placeholder }}</span>
    </div>

    <div v-else-if="type === 'multiple'" class="space-y-2">
      <label
        v-for="option in normalizedOptions"
        :key="option.id || option.value"
        class="flex cursor-pointer items-center gap-2"
      >
        <Checkbox
          v-model="selectedValues"
          :value="option.value || option.id"
          :disabled="disabled"
        />
        <span>{{ option.label || option.value }}</span>
      </label>
    </div>

    <div v-if="error" class="mt-1 text-sm text-[var(--color-error)]">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, useId } from "vue";
import Select from "@/components/ui/Select.vue";
import Radio from "@/components/ui/Radio.vue";
import Checkbox from "@/components/ui/Checkbox.vue";

interface Option {
  id?: string;
  value?: string;
  label?: string;
}

interface Props {
  modelValue?: string | boolean | string[];
  type: "select" | "radio" | "checkbox" | "multiple";
  placeholder?: string;
  required?: boolean;
  disabled?: boolean;
  /** Options: array of { id?, value?, label? } or strings (normalized to objects) */
  options?: (Option | string)[];
  error?: string;
}

interface Emits {
  (e: "update:modelValue", value: string | boolean | string[]): void;
  (e: "blur"): void;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: "",
  type: "select",
  placeholder: "",
  required: false,
  disabled: false,
  options: () => [],
  error: ""
});

const emit = defineEmits<Emits>();

const radioGroupName = useId();

// Normalize options: accept objects { id?, value?, label? } or strings
const normalizedOptions = computed((): Option[] => {
  const raw = props.options ?? [];
  if (!Array.isArray(raw)) return [];
  return raw.map((item): Option => {
    if (typeof item === "string") {
      return { id: item, value: item, label: item };
    }
    const o = item as Option;
    const value = String(o.value ?? o.id ?? "");
    return {
      id: o.id ?? value,
      value,
      label: o.label ?? value
    };
  });
});

const localValue = ref(props.modelValue);
const selectedValues = ref<string[]>(Array.isArray(props.modelValue) ? props.modelValue : []);

watch(
  () => props.modelValue,
  (newValue) => {
    if (props.type === "radio") {
      localValue.value = newValue != null ? String(newValue) : "";
    } else if (props.type === "multiple") {
      localValue.value = newValue;
      selectedValues.value = Array.isArray(newValue) ? newValue : [];
    } else {
      localValue.value = newValue;
    }
  }
);

watch(localValue, (newValue) => {
  emit("update:modelValue", newValue);
});

watch(selectedValues, (newValue) => {
  if (props.type === "multiple") {
    emit("update:modelValue", newValue);
  }
});

function handleBlur(): void {
  emit("blur");
}
</script>

