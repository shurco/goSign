import { computed, type Ref, watch, nextTick } from 'vue'
import { Parser } from 'expr-eval'
import type { Field } from '@/models/template'

export function useFormulas(
  fields: Ref<Field[]>,
  formData: Ref<Record<string, any>>
) {
  const parser = new Parser()
  
  // Add custom functions
  parser.functions.SUM = (...args: number[]) => {
    return args.reduce((sum, val) => sum + (Number(val) || 0), 0)
  }
  
  parser.functions.IF = (condition: boolean, trueVal: number, falseVal: number) => {
    return condition ? trueVal : falseVal
  }
  
  parser.functions.MAX = (...args: number[]) => {
    return Math.max(...args.map(v => Number(v) || 0))
  }
  
  parser.functions.MIN = (...args: number[]) => {
    return Math.min(...args.map(v => Number(v) || 0))
  }
  
  parser.functions.ROUND = (value: number, decimals: number = 0) => {
    const multiplier = Math.pow(10, decimals)
    return Math.round(value * multiplier) / multiplier
  }
  
  // Evaluate single formula
  function evaluateFormula(formula: string): number | null {
    try {
      // Prepare variables from formData
      const variables: Record<string, number> = {}
      
      for (const field of fields.value) {
        const value = formData.value[field.id]
        variables[field.id] = Number(value) || 0
      }
      
      const result = parser.evaluate(formula, variables)
      return Number(result)
    } catch (error) {
      console.error('Formula evaluation error:', error)
      return null
    }
  }
  
  // Calculated field values - computed will automatically recalculate when formData changes
  // Access formData.value inside computed to ensure reactivity
  const calculatedValues = computed(() => {
    // Access formData.value to track changes
    const _ = formData.value // Force dependency tracking
    const values: Record<string, number> = {}
    
    for (const field of fields.value) {
      if (field.formula) {
        const result = evaluateFormula(field.formula)
        if (result !== null) {
          values[field.id] = result
        }
      }
    }
    
    return values
  })
  
  // Auto-update formData with calculated values
  // Use nextTick to prevent infinite loops
  watch(calculatedValues, async (newValues, oldValues) => {
    await nextTick()
    
    for (const [fieldId, value] of Object.entries(newValues)) {
      if (value !== undefined && value !== null) {
        const oldValue = oldValues?.[fieldId]
        const currentValue = formData.value[fieldId]
        // Only update if value actually changed and is different from current
        // Also check if it's a calculated field (has formula)
        const field = fields.value.find(f => f.id === fieldId)
        if (field?.formula && oldValue !== value && currentValue !== value) {
          formData.value[fieldId] = value
        }
      }
    }
  }, { immediate: true, deep: true })
  
  return {
    calculatedValues,
    evaluateFormula
  }
}
