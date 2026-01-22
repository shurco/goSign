<template>
  <div class="formula-builder">
    <h3 class="mb-4 text-lg font-semibold">{{ $t('fields.formula.title') }}</h3>
    
    <!-- Formula input -->
    <div class="formula-editor mb-4">
      <textarea
        v-model="formula"
        :placeholder="$t('fields.formula.placeholder')"
        class="formula-input w-full rounded border border-gray-300 p-2 font-mono text-sm"
        rows="3"
        @input="validateFormula"
      />
      
      <!-- Validation error -->
      <div v-if="validationError" class="mt-2 text-sm text-red-600">
        {{ validationError }}
      </div>
      
      <!-- Preview result -->
      <div v-if="previewResult !== null && !validationError" class="mt-2 text-sm text-green-600">
        {{ $t('fields.formula.preview') }}: {{ previewResult }}
      </div>
    </div>
    
    <!-- Field reference buttons -->
    <div class="mb-4">
      <h4 class="mb-2 text-sm font-medium">{{ $t('fields.formula.insertField') }}</h4>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="field in availableFields"
          :key="field.id"
          class="rounded-md border border-gray-300 bg-white px-2 py-1 text-xs hover:bg-gray-50"
          @click="insertField(field.id)"
        >
          {{ field.name }}
        </button>
      </div>
    </div>
    
    <!-- Function buttons -->
    <div class="mb-4">
      <h4 class="mb-2 text-sm font-medium">{{ $t('fields.formula.functions') }}</h4>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="func in availableFunctions"
          :key="func.name"
          class="rounded-md border border-gray-300 bg-white px-2 py-1 text-xs hover:bg-gray-50"
          @click="insertFunction(func.syntax)"
          :title="func.description"
        >
          {{ func.name }}
        </button>
      </div>
    </div>
    
    <!-- Examples -->
    <div>
      <h4 class="mb-2 text-sm font-medium">{{ $t('fields.formula.examples') }}</h4>
      <div class="space-y-1">
        <div
          v-for="example in examples"
          :key="example.label"
          class="cursor-pointer rounded p-2 text-sm hover:bg-gray-100"
          @click="formula = example.formula"
        >
          <code class="text-blue-600">{{ example.formula }}</code>
          <span class="ml-2 text-gray-600">- {{ example.label }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useFormulas } from '@/composables/useFormulas'
import type { Field } from '@/models/template'
import { apiPost } from '@/services/api'

interface Props {
  field: Field
  availableFields: Field[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:formula': [formula: string]
}>()

const formula = ref(props.field.formula || '')
const validationError = ref<string | null>(null)

const availableFunctions = [
  { name: 'SUM', syntax: 'SUM(field_1, field_2)', description: 'Sum of multiple fields' },
  { name: 'IF', syntax: 'IF(field_1 > 100, field_2, 0)', description: 'Conditional value' },
  { name: 'MAX', syntax: 'MAX(field_1, field_2)', description: 'Maximum value' },
  { name: 'MIN', syntax: 'MIN(field_1, field_2)', description: 'Minimum value' },
  { name: 'ROUND', syntax: 'ROUND(field_1, 2)', description: 'Round to decimals' }
]

const examples = [
  { label: 'Sum two fields', formula: 'field_1 + field_2' },
  { label: 'Calculate tax (20%)', formula: 'field_1 * 1.2' },
  { label: 'Conditional discount', formula: 'IF(field_1 > 1000, field_1 * 0.9, field_1)' },
  { label: 'Sum with tax', formula: 'SUM(field_1, field_2) * 1.2' }
]

// Sample data for preview
const sampleFormData = computed(() => {
  const data: Record<string, any> = {}
  for (const field of props.availableFields) {
    data[field.id] = 10 // Sample value
  }
  return data
})

const { evaluateFormula } = useFormulas(
  computed(() => props.availableFields),
  computed(() => sampleFormData.value)
)

const previewResult = computed(() => {
  if (!formula.value || validationError.value) {
    return null
  }
  const result = evaluateFormula(formula.value)
  return result !== null ? result.toFixed(2) : null
})

function insertField(fieldId: string) {
  formula.value += fieldId
  validateFormula()
}

function insertFunction(syntax: string) {
  formula.value += syntax
  validateFormula()
}

async function validateFormula() {
  if (!formula.value.trim()) {
    validationError.value = null
    emit('update:formula', '')
    return
  }
  
  try {
    const response = await apiPost('/api/v1/templates/formulas/validate', {
      formula: formula.value,
      fields: props.availableFields
    })
    
    // API returns { data: ..., message: ... }
    if (response && !response.message?.includes('error')) {
      validationError.value = null
      emit('update:formula', formula.value)
    } else {
      validationError.value = response.message || 'Invalid formula'
    }
  } catch (error: any) {
    validationError.value = error.message || 'Invalid formula'
  }
}

// Watch formula changes
watch(formula, () => {
  validateFormula()
}, { immediate: true })

// Expose method to get current formula
defineExpose({
  getFormula: () => formula.value,
  formula
})
</script>

<style scoped>
.formula-input {
  font-family: 'Courier New', monospace;
}
</style>
