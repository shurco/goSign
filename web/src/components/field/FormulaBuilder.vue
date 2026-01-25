<template>
  <div class="formula-builder space-y-5">
    <p class="text-sm text-gray-600">
      {{ $t('fields.formula.description') || 'Use field IDs and operators to compute a value. Click fields and functions below to insert.' }}
    </p>

    <!-- Formula editor -->
    <div class="formula-editor">
      <label class="mb-1.5 block text-sm font-medium text-gray-700">
        {{ $t('fields.formula.expression') || 'Formula' }}
      </label>
      <textarea
        :value="displayFormula"
        :placeholder="$t('fields.formula.placeholder')"
        :class="[
          'formula-input w-full rounded-xl border px-4 py-3 font-mono text-sm leading-relaxed transition-colors focus:outline-none focus:ring-2',
          validationError
            ? 'border-red-300 bg-red-50/30 focus:border-red-400 focus:ring-red-200'
            : 'border-gray-300 bg-white focus:border-indigo-400 focus:ring-indigo-200'
        ]"
        rows="4"
        spellcheck="false"
        @input="onFormulaDisplayInput"
      />
      <div v-if="validationError" class="mt-2 flex items-center gap-2 text-sm text-red-600">
        <span aria-hidden="true">⊗</span>
        <span>{{ validationError }}</span>
      </div>
      <div
        v-else-if="previewResult !== null"
        class="mt-2 flex items-center gap-2 text-sm text-emerald-600"
      >
        <span aria-hidden="true">✓</span>
        <span>{{ $t('fields.formula.preview') }}: <strong>{{ previewResult }}</strong></span>
      </div>
    </div>

    <!-- Insert field -->
    <section class="rounded-xl border border-gray-200 bg-gray-50/50 p-4">
      <h4 class="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
        {{ $t('fields.formula.insertField') }}
      </h4>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="f in availableFields"
          :key="f.id"
          type="button"
          class="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-700 shadow-sm transition-colors hover:border-indigo-200 hover:bg-indigo-50 hover:text-indigo-700"
          @click="insertField(f.id)"
        >
          {{ f.displayName ?? f.name ?? f.id }}
        </button>
      </div>
      <p v-if="!availableFields.length" class="text-sm text-gray-500">
        {{ $t('fields.formula.noFields') || 'No number/text fields available.' }}
      </p>
    </section>

    <!-- Functions -->
    <section class="rounded-xl border border-gray-200 bg-gray-50/50 p-4">
      <h4 class="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
        {{ $t('fields.formula.functions') }}
      </h4>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="func in availableFunctions"
          :key="func.name"
          type="button"
          class="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 shadow-sm transition-colors hover:border-indigo-200 hover:bg-indigo-50 hover:text-indigo-700"
          :title="func.description"
          @click="insertFunction(func.syntax)"
        >
          {{ func.name }}
        </button>
      </div>
    </section>

    <!-- Examples -->
    <section class="rounded-xl border border-gray-200 bg-gray-50/50 p-4">
      <h4 class="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
        {{ $t('fields.formula.examples') }}
      </h4>
      <div class="space-y-1.5">
        <button
          v-for="example in examples"
          :key="example.label"
          type="button"
          class="flex w-full items-start gap-3 rounded-lg border border-transparent p-2.5 text-left text-sm transition-colors hover:border-gray-200 hover:bg-white"
          @click="applyExample(example.formula)"
        >
          <code class="shrink-0 rounded bg-gray-200 px-2 py-0.5 font-mono text-xs text-indigo-600">
            {{ exampleDisplayFormula(example.formula) }}
          </code>
          <span class="text-gray-600">{{ example.label }}</span>
        </button>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useFormulas } from '@/composables/useFormulas'
import type { Field } from '@/models/template'
import { apiPost } from '@/services/api'

interface FieldWithDisplayName extends Field {
  displayName?: string
}

interface Props {
  field: Field
  availableFields: FieldWithDisplayName[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:formula': [formula: string]
}>()

const formula = ref(props.field.preferences?.formula ?? (props.field as any).formula ?? '')
const validationError = ref<string | null>(null)

// Convert formula (field IDs) to display string ([[field name]])
function formulaToDisplay(formulaStr: string): string {
  let out = formulaStr
  const byId = props.availableFields
  const escapeRe = (s: string) => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const sorted = [...byId].sort((a, b) => b.id.length - a.id.length)
  for (const f of sorted) {
    const name = f.displayName ?? f.name ?? f.id
    const re = new RegExp(escapeRe(f.id), 'g')
    out = out.replace(re, `[[${name}]]`)
  }
  return out
}

// Parse display string ([[field name]]) back to formula (field IDs)
function displayToFormula(displayStr: string): string {
  const re = /\[\[([^\]]*?)\]\]/g
  return displayStr.replace(re, (_, name) => {
    const f = props.availableFields.find(
      (x) => (x.displayName ?? x.name ?? x.id) === name
    )
    return f ? f.id : `[[${name}]]`
  })
}

const displayFormula = computed(() => formulaToDisplay(formula.value))

function onFormulaDisplayInput(e: Event) {
  const target = e.target as HTMLTextAreaElement
  formula.value = displayToFormula(target.value)
}

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

function applyExample(exampleFormula: string) {
  formula.value = exampleFormula
  validateFormula()
}

function exampleDisplayFormula(exampleFormula: string): string {
  return formulaToDisplay(exampleFormula)
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
