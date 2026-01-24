<template>
  <div class="field-input-wrapper">
    <div class="flex items-center gap-1">
      <input
        v-for="(_, index) in cellCount"
        :key="index"
        ref="cellInputs"
        v-model="cellValues[index]"
        type="text"
        maxlength="1"
        class="h-12 w-12 rounded border border-gray-300 text-center text-lg font-semibold focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary"
        :class="{ 'border-error': error }"
        :disabled="disabled"
        :readonly="readonly"
        @input="handleCellInput(index, $event)"
        @keydown="handleKeyDown(index, $event)"
        @paste="handlePaste(index, $event)"
        @focus="handleFocus(index)"
      />
    </div>
    <div v-if="error" class="mt-1 text-sm text-[var(--color-error)]">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";

interface Props {
  modelValue?: string;
  cellCount?: number;
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
  cellCount: 6,
  placeholder: "",
  required: false,
  readonly: false,
  disabled: false,
  error: ""
});

const emit = defineEmits<Emits>();
const cellInputs = ref<(HTMLInputElement | null)[]>([]);
const cellValues = ref<string[]>([]);

// Initialize cell values from modelValue
function initializeCells(): void {
  const value = props.modelValue || "";
  cellValues.value = Array(props.cellCount)
    .fill("")
    .map((_, i) => value[i] || "");
}

// Initialize on mount and when cellCount changes
watch(
  () => props.cellCount,
  () => {
    initializeCells();
  },
  { immediate: true }
);

watch(
  () => props.modelValue,
  (newValue) => {
    if (newValue !== getCombinedValue()) {
      initializeCells();
    }
  }
);

function getCombinedValue(): string {
  return cellValues.value.join("");
}

function handleCellInput(index: number, event: Event): void {
  const input = event.target as HTMLInputElement;
  const value = input.value.trim().slice(-1); // Take only last character
  
  // Update the cell value
  cellValues.value[index] = value;
  
  // Emit combined value
  emit("update:modelValue", getCombinedValue());
  
  // Auto-advance to next cell if value entered
  if (value && index < props.cellCount - 1) {
    nextTick(() => {
      cellInputs.value[index + 1]?.focus();
    });
  }
}

function handleKeyDown(index: number, event: KeyboardEvent): void {
  // Handle backspace
  if (event.key === "Backspace" && !cellValues.value[index] && index > 0) {
    event.preventDefault();
    cellValues.value[index - 1] = "";
    cellInputs.value[index - 1]?.focus();
    emit("update:modelValue", getCombinedValue());
  }
  
  // Handle arrow keys
  if (event.key === "ArrowLeft" && index > 0) {
    event.preventDefault();
    cellInputs.value[index - 1]?.focus();
  }
  if (event.key === "ArrowRight" && index < props.cellCount - 1) {
    event.preventDefault();
    cellInputs.value[index + 1]?.focus();
  }
  
  // Only allow alphanumeric characters
  if (event.key.length === 1 && !/[a-zA-Z0-9]/.test(event.key)) {
    event.preventDefault();
  }
}

function handlePaste(index: number, event: ClipboardEvent): void {
  event.preventDefault();
  const pastedText = event.clipboardData?.getData("text") || "";
  const chars = pastedText.slice(0, props.cellCount - index).split("");
  
  chars.forEach((char, i) => {
    const cellIndex = index + i;
    if (cellIndex < props.cellCount && /[a-zA-Z0-9]/.test(char)) {
      cellValues.value[cellIndex] = char;
    }
  });
  
  emit("update:modelValue", getCombinedValue());
  
  // Focus the next empty cell or last cell
  const nextIndex = Math.min(index + chars.length, props.cellCount - 1);
  nextTick(() => {
    cellInputs.value[nextIndex]?.focus();
  });
}

function handleFocus(index: number): void {
  // Select all text when focusing
  nextTick(() => {
    cellInputs.value[index]?.select();
  });
}
</script>
