<template>
  <TextInput
    v-if="isTextType"
    v-model="localValue"
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
    v-model="localValue"
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
    v-model="localValue"
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
      v-model="localValue"
    :type="type"
      :placeholder="placeholder"
      :required="required"
      :readonly="readonly"
      :disabled="disabled"
    :error="error"
    @update:modelValue="handleUpdate"
    @blur="$emit('blur')"
    />
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
  error: ""
});

const emit = defineEmits<Emits>();
const localValue = ref(props.modelValue);

const isTextType = computed(() => {
  return ["text", "number", "phone"].includes(props.type);
});

const isSelectType = computed(() => {
  return ["select", "radio", "checkbox", "multiple"].includes(props.type);
});

watch(
  () => props.modelValue,
  (newValue) => {
    localValue.value = newValue;
  }
);

function handleUpdate(value: string | boolean | string[]): void {
  localValue.value = value;
  emit("update:modelValue", value);
}
</script>
