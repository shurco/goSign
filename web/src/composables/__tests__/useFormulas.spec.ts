import { describe, it, expect, beforeEach, vi } from 'vitest'
import { ref, nextTick } from 'vue'
import { useFormulas } from '../useFormulas'
import type { Field } from '@/models/template'

describe('useFormulas - Formula Parsing and Calculation', () => {
  let fields: ReturnType<typeof ref<Field[]>>
  let formData: ReturnType<typeof ref<Record<string, any>>>

  beforeEach(() => {
    // Suppress console.error in tests
    vi.spyOn(console, 'error').mockImplementation(() => {})

    fields = ref<Field[]>([
      {
        id: 'field_1',
        type: 'number',
        name: 'Field 1',
        required: false,
      },
      {
        id: 'field_2',
        type: 'number',
        name: 'Field 2',
        required: false,
      },
      {
        id: 'field_3',
        type: 'number',
        name: 'Field 3',
        required: false,
      },
      {
        id: 'calculated_field',
        type: 'number',
        name: 'Calculated Field',
        required: false,
        formula: 'field_1 + field_2',
        calculation_type: 'number',
      },
    ])

    formData = ref<Record<string, any>>({
      field_1: 10,
      field_2: 20,
      field_3: 5,
      calculated_field: 0,
    })
  })

  describe('Basic Arithmetic Operations', () => {
    it('should evaluate addition', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2')
      expect(result).toBe(30)
    })

    it('should evaluate subtraction', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 - field_2')
      expect(result).toBe(-10)
    })

    it('should evaluate multiplication', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 * field_2')
      expect(result).toBe(200)
    })

    it('should evaluate division', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_2 / field_1')
      expect(result).toBe(2)
    })

    it('should handle division by zero', () => {
      formData.value.field_1 = 0
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_2 / field_1')
      expect(result).toBe(Infinity)
    })

    it('should evaluate complex expressions with parentheses', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('(field_1 + field_2) * field_3')
      expect(result).toBe(150) // (10 + 20) * 5 = 150
    })

    it('should handle order of operations', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2 * field_3')
      expect(result).toBe(110) // 10 + (20 * 5) = 110
    })
  })

  describe('SUM Function', () => {
    it('should sum multiple values', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('SUM(field_1, field_2, field_3)')
      expect(result).toBe(35) // 10 + 20 + 5
    })

    it('should sum two values', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('SUM(field_1, field_2)')
      expect(result).toBe(30)
    })

    it('should handle non-numeric values in SUM', () => {
      formData.value.field_1 = 'not a number'
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('SUM(field_1, field_2)')
      expect(result).toBe(20) // 0 + 20
    })

    it('should return 0 for empty SUM', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('SUM()')
      expect(result).toBe(0)
    })
  })

  describe('IF Function', () => {
    it('should return true value when condition is true', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('IF(field_1 > 5, 100, 0)')
      expect(result).toBe(100)
    })

    it('should return false value when condition is false', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('IF(field_1 < 5, 100, 0)')
      expect(result).toBe(0)
    })

    it('should handle equality condition', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('IF(field_1 == 10, 50, 0)')
      expect(result).toBe(50)
    })

    it('should handle nested IF statements', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('IF(field_1 > 5, IF(field_2 > 15, 200, 100), 0)')
      expect(result).toBe(200)
    })
  })

  describe('MAX Function', () => {
    it('should return maximum value', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('MAX(field_1, field_2, field_3)')
      expect(result).toBe(20)
    })

    it('should handle negative values', () => {
      formData.value.field_1 = -10
      formData.value.field_2 = -5
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('MAX(field_1, field_2)')
      expect(result).toBe(-5)
    })

    it('should handle non-numeric values in MAX', () => {
      formData.value.field_1 = 'not a number'
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('MAX(field_1, field_2)')
      expect(result).toBe(20) // max(0, 20)
    })
  })

  describe('MIN Function', () => {
    it('should return minimum value', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('MIN(field_1, field_2, field_3)')
      expect(result).toBe(5)
    })

    it('should handle negative values', () => {
      formData.value.field_1 = -10
      formData.value.field_2 = -5
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('MIN(field_1, field_2)')
      expect(result).toBe(-10)
    })
  })

  describe('ROUND Function', () => {
    it('should round to nearest integer by default', () => {
      formData.value.field_1 = 10.7
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('ROUND(field_1)')
      expect(result).toBe(11)
    })

    it('should round to specified decimal places', () => {
      formData.value.field_1 = 10.567
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('ROUND(field_1, 2)')
      expect(result).toBe(10.57)
    })

    it('should round down when appropriate', () => {
      formData.value.field_1 = 10.4
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('ROUND(field_1)')
      expect(result).toBe(10)
    })

    it('should handle zero decimal places', () => {
      formData.value.field_1 = 10.567
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('ROUND(field_1, 0)')
      expect(result).toBe(11)
    })
  })

  describe('Variable Substitution', () => {
    it('should substitute field values from formData', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2')
      expect(result).toBe(30)
    })

    it('should handle missing field values as zero', () => {
      delete formData.value.field_1
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2')
      expect(result).toBe(20) // 0 + 20
    })

    it('should handle string numbers', () => {
      formData.value.field_1 = '10'
      formData.value.field_2 = '20'
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2')
      expect(result).toBe(30)
    })

    it('should handle non-numeric strings as zero', () => {
      formData.value.field_1 = 'not a number'
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2')
      expect(result).toBe(20) // 0 + 20
    })
  })

  describe('Calculated Values Computed', () => {
    it('should compute values for fields with formulas', () => {
      const { calculatedValues } = useFormulas(fields, formData)
      expect(calculatedValues.value).toHaveProperty('calculated_field')
      expect(calculatedValues.value.calculated_field).toBe(30)
    })

    it('should not compute values for fields without formulas', () => {
      const { calculatedValues } = useFormulas(fields, formData)
      expect(calculatedValues.value).not.toHaveProperty('field_1')
      expect(calculatedValues.value).not.toHaveProperty('field_2')
    })

    it('should update calculated values when formData changes', () => {
      const { calculatedValues } = useFormulas(fields, formData)
      expect(calculatedValues.value.calculated_field).toBe(30)

      formData.value.field_1 = 50
      expect(calculatedValues.value.calculated_field).toBe(70)
    })

    it('should handle multiple calculated fields', () => {
      fields.value.push({
        id: 'calculated_field_2',
        type: 'number',
        name: 'Calculated Field 2',
        required: false,
        formula: 'field_2 * field_3',
        calculation_type: 'number',
      })

      const { calculatedValues } = useFormulas(fields, formData)
      expect(calculatedValues.value.calculated_field).toBe(30)
      expect(calculatedValues.value.calculated_field_2).toBe(100)
    })
  })

  describe('Auto-update formData', () => {
    it('should automatically update formData with calculated values', async () => {
      useFormulas(fields, formData)
      
      // Wait for next tick to allow watch to execute
      await nextTick()
      expect(formData.value.calculated_field).toBe(30)
    })

    it('should update formData when calculated values change', async () => {
      useFormulas(fields, formData)
      
      formData.value.field_1 = 100
      
      await nextTick()
      expect(formData.value.calculated_field).toBe(120)
    })
  })

  describe('Error Handling', () => {
    it('should return null for invalid formula syntax', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + + field_2')
      expect(result).toBeNull()
    })

    it('should return null for missing closing parenthesis', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('SUM(field_1, field_2')
      expect(result).toBeNull()
    })

    it('should return null for undefined function', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('UNKNOWN_FUNCTION(field_1)')
      expect(result).toBeNull()
    })

    it('should handle empty formula string', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('')
      expect(result).toBeNull()
    })

    it('should not throw error for invalid formula in calculatedValues', () => {
      fields.value[3].formula = 'invalid formula syntax'
      
      const { calculatedValues } = useFormulas(fields, formData)
      expect(calculatedValues.value.calculated_field).toBeUndefined()
    })
  })

  describe('Complex Formulas', () => {
    it('should evaluate nested function calls', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('SUM(field_1, field_2) * field_3')
      expect(result).toBe(150) // (10 + 20) * 5
    })

    it('should evaluate formula with multiple functions', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('MAX(field_1, field_2) + MIN(field_1, field_2)')
      expect(result).toBe(30) // 20 + 10
    })

    it('should evaluate conditional formula with arithmetic', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('IF(field_1 > 5, field_1 * 2, field_1 + 5)')
      expect(result).toBe(20) // 10 * 2
    })

    it('should handle percentage calculations', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 * 0.1')
      expect(result).toBe(1) // 10 * 0.1
    })

    it('should handle power operations', () => {
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 ^ 2')
      expect(result).toBe(100) // 10^2
    })
  })

  describe('Edge Cases', () => {
    it('should handle null values in formData', () => {
      formData.value.field_1 = null
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2')
      expect(result).toBe(20) // 0 + 20
    })

    it('should handle undefined values in formData', () => {
      formData.value.field_1 = undefined
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2')
      expect(result).toBe(20) // 0 + 20
    })

    it('should handle empty string values', () => {
      formData.value.field_1 = ''
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2')
      expect(result).toBe(20) // 0 + 20
    })

    it('should handle very large numbers', () => {
      formData.value.field_1 = 1e10
      formData.value.field_2 = 2e10
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2')
      expect(result).toBe(3e10)
    })

    it('should handle very small numbers', () => {
      formData.value.field_1 = 0.0001
      formData.value.field_2 = 0.0002
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2')
      expect(result).toBeCloseTo(0.0003, 5)
    })

    it('should handle negative numbers', () => {
      formData.value.field_1 = -10
      formData.value.field_2 = -20
      const { evaluateFormula } = useFormulas(fields, formData)
      const result = evaluateFormula('field_1 + field_2')
      expect(result).toBe(-30)
    })
  })
})
