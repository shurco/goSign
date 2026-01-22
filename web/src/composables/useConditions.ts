import { computed, type Ref, watch } from 'vue'
import type { Field, FieldCondition, FieldConditionGroup } from '@/models/template'

export interface FieldState {
  visible: boolean
  required: boolean
  disabled: boolean
}

export function useConditions(
  fields: Ref<Field[]>,
  formData: Ref<Record<string, any>>
) {
  // Evaluate single condition
  function evaluateCondition(condition: FieldCondition): boolean {
    const fieldValue = formData.value[condition.field_id]
    
    // Handle undefined/null values
    if (fieldValue === undefined || fieldValue === null) {
      if (condition.operator === 'is_empty') return true
      if (condition.operator === 'is_not_empty') return false
      // For other operators, treat undefined as empty string
      const emptyValue = ''
      switch (condition.operator) {
        case 'equals':
          return emptyValue === condition.value
        case 'not_equals':
          return emptyValue !== condition.value
        default:
          return false
      }
    }
    
    switch (condition.operator) {
      case 'equals':
        return fieldValue === condition.value
      
      case 'not_equals':
        return fieldValue !== condition.value
      
      case 'contains':
        return String(fieldValue).includes(String(condition.value))
      
      case 'not_contains':
        return !String(fieldValue).includes(String(condition.value))
      
      case 'greater_than':
        return Number(fieldValue) > Number(condition.value)
      
      case 'less_than':
        return Number(fieldValue) < Number(condition.value)
      
      case 'is_empty':
        return !fieldValue || fieldValue === '' || 
               (Array.isArray(fieldValue) && fieldValue.length === 0)
      
      case 'is_not_empty':
        return !!fieldValue && fieldValue !== '' &&
               !(Array.isArray(fieldValue) && fieldValue.length === 0)
      
      default:
        return false
    }
  }
  
  // Evaluate condition group (with AND/OR logic)
  function evaluateGroup(group: FieldConditionGroup): boolean {
    if (group.logic === 'AND') {
      return group.conditions.every(cond => evaluateCondition(cond))
    } else {
      return group.conditions.some(cond => evaluateCondition(cond))
    }
  }
  
  // Calculate field state based on all conditions
  function getFieldState(field: Field): FieldState {
    const state: FieldState = {
      visible: true,
      required: field.required || false,
      disabled: false
    }
    
    if (!field.condition_groups || field.condition_groups.length === 0) {
      return state
    }
    
    // Evaluate all condition groups
    for (const group of field.condition_groups) {
      const conditionMet = evaluateGroup(group)
      
      if (conditionMet) {
        switch (group.action) {
          case 'show':
            state.visible = true
            break
          case 'hide':
            state.visible = false
            break
          case 'require':
            state.required = true
            break
          case 'disable':
            state.disabled = true
            break
        }
      } else {
        // Inverse logic when condition not met
        if (group.action === 'show') {
          state.visible = false
        }
      }
    }
    
    return state
  }
  
  // Computed map of field states
  // Access formData.value inside computed to ensure reactivity
  const fieldStates = computed(() => {
    // Access formData.value to track changes
    const _ = formData.value // Force dependency tracking
    const states: Record<string, FieldState> = {}
    
    for (const field of fields.value) {
      states[field.id] = getFieldState(field)
    }
    
    return states
  })
  
  return {
    fieldStates,
    evaluateCondition,
    evaluateGroup,
    getFieldState
  }
}
