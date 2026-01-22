import { describe, it, expect, beforeEach } from 'vitest'
import { ref } from 'vue'
import { useConditions } from '../useConditions'
import type { Field, FieldCondition, FieldConditionGroup, ConditionOperator } from '@/models/template'

describe('useConditions - Condition Evaluation Logic', () => {
  let fields: ReturnType<typeof ref<Field[]>>
  let formData: ReturnType<typeof ref<Record<string, any>>>

  beforeEach(() => {
    fields = ref<Field[]>([
      {
        id: 'field_1',
        type: 'text',
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
        type: 'text',
        name: 'Field 3',
        required: false,
      },
    ])

    formData = ref<Record<string, any>>({
      field_1: '',
      field_2: 0,
      field_3: '',
    })
  })

  describe('Equals Operator', () => {
    it('should return true when field value equals condition value', () => {
      formData.value.field_1 = 'test'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'equals',
        value: 'test',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return false when field value does not equal condition value', () => {
      formData.value.field_1 = 'test'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'equals',
        value: 'different',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })

    it('should handle numeric equality', () => {
      formData.value.field_2 = 42
      
      const condition: FieldCondition = {
        field_id: 'field_2',
        operator: 'equals',
        value: 42,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })
  })

  describe('Not Equals Operator', () => {
    it('should return true when field value does not equal condition value', () => {
      formData.value.field_1 = 'test'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'not_equals',
        value: 'different',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return false when field value equals condition value', () => {
      formData.value.field_1 = 'test'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'not_equals',
        value: 'test',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })
  })

  describe('Contains Operator', () => {
    it('should return true when field value contains condition value', () => {
      formData.value.field_1 = 'hello world'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'contains',
        value: 'world',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return false when field value does not contain condition value', () => {
      formData.value.field_1 = 'hello world'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'contains',
        value: 'xyz',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })

    it('should handle case-sensitive matching', () => {
      formData.value.field_1 = 'Hello World'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'contains',
        value: 'hello',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })

    it('should convert non-string values to strings', () => {
      formData.value.field_2 = 12345
      
      const condition: FieldCondition = {
        field_id: 'field_2',
        operator: 'contains',
        value: '234',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })
  })

  describe('Not Contains Operator', () => {
    it('should return true when field value does not contain condition value', () => {
      formData.value.field_1 = 'hello world'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'not_contains',
        value: 'xyz',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return false when field value contains condition value', () => {
      formData.value.field_1 = 'hello world'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'not_contains',
        value: 'world',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })
  })

  describe('Greater Than Operator', () => {
    it('should return true when field value is greater than condition value', () => {
      formData.value.field_2 = 10
      
      const condition: FieldCondition = {
        field_id: 'field_2',
        operator: 'greater_than',
        value: 5,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return false when field value is not greater than condition value', () => {
      formData.value.field_2 = 5
      
      const condition: FieldCondition = {
        field_id: 'field_2',
        operator: 'greater_than',
        value: 10,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })

    it('should handle string numbers', () => {
      formData.value.field_1 = '10'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'greater_than',
        value: '5',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return false when values are equal', () => {
      formData.value.field_2 = 5
      
      const condition: FieldCondition = {
        field_id: 'field_2',
        operator: 'greater_than',
        value: 5,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })
  })

  describe('Less Than Operator', () => {
    it('should return true when field value is less than condition value', () => {
      formData.value.field_2 = 5
      
      const condition: FieldCondition = {
        field_id: 'field_2',
        operator: 'less_than',
        value: 10,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return false when field value is not less than condition value', () => {
      formData.value.field_2 = 10
      
      const condition: FieldCondition = {
        field_id: 'field_2',
        operator: 'less_than',
        value: 5,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })

    it('should return false when values are equal', () => {
      formData.value.field_2 = 5
      
      const condition: FieldCondition = {
        field_id: 'field_2',
        operator: 'less_than',
        value: 5,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })
  })

  describe('Is Empty Operator', () => {
    it('should return true when field value is empty string', () => {
      formData.value.field_1 = ''
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'is_empty',
        value: null,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return true when field value is null', () => {
      formData.value.field_1 = null
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'is_empty',
        value: null,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return true when field value is undefined', () => {
      formData.value.field_1 = undefined
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'is_empty',
        value: null,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return true when field value is empty array', () => {
      formData.value.field_1 = []
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'is_empty',
        value: null,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return false when field value is not empty', () => {
      formData.value.field_1 = 'test'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'is_empty',
        value: null,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })

    it('should return false when field value is non-empty array', () => {
      formData.value.field_1 = ['item1', 'item2']
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'is_empty',
        value: null,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })
  })

  describe('Is Not Empty Operator', () => {
    it('should return true when field value is not empty', () => {
      formData.value.field_1 = 'test'
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'is_not_empty',
        value: null,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return true when field value is non-empty array', () => {
      formData.value.field_1 = ['item1']
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'is_not_empty',
        value: null,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should return false when field value is empty string', () => {
      formData.value.field_1 = ''
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'is_not_empty',
        value: null,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })

    it('should return false when field value is empty array', () => {
      formData.value.field_1 = []
      
      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'is_not_empty',
        value: null,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })
  })

  describe('AND Logic in Condition Groups', () => {
    it('should return true when all conditions in AND group are met', () => {
      formData.value.field_1 = 'test'
      formData.value.field_2 = 10

      const group: FieldConditionGroup = {
        logic: 'AND',
        conditions: [
          {
            field_id: 'field_1',
            operator: 'equals',
            value: 'test',
          },
          {
            field_id: 'field_2',
            operator: 'greater_than',
            value: 5,
          },
        ],
        action: 'show',
      }

      const { evaluateGroup } = useConditions(fields, formData)
      expect(evaluateGroup(group)).toBe(true)
    })

    it('should return false when any condition in AND group is not met', () => {
      formData.value.field_1 = 'test'
      formData.value.field_2 = 3

      const group: FieldConditionGroup = {
        logic: 'AND',
        conditions: [
          {
            field_id: 'field_1',
            operator: 'equals',
            value: 'test',
          },
          {
            field_id: 'field_2',
            operator: 'greater_than',
            value: 5,
          },
        ],
        action: 'show',
      }

      const { evaluateGroup } = useConditions(fields, formData)
      expect(evaluateGroup(group)).toBe(false)
    })
  })

  describe('OR Logic in Condition Groups', () => {
    it('should return true when any condition in OR group is met', () => {
      formData.value.field_1 = 'test'
      formData.value.field_2 = 3

      const group: FieldConditionGroup = {
        logic: 'OR',
        conditions: [
          {
            field_id: 'field_1',
            operator: 'equals',
            value: 'test',
          },
          {
            field_id: 'field_2',
            operator: 'greater_than',
            value: 5,
          },
        ],
        action: 'show',
      }

      const { evaluateGroup } = useConditions(fields, formData)
      expect(evaluateGroup(group)).toBe(true)
    })

    it('should return false when no conditions in OR group are met', () => {
      formData.value.field_1 = 'different'
      formData.value.field_2 = 3

      const group: FieldConditionGroup = {
        logic: 'OR',
        conditions: [
          {
            field_id: 'field_1',
            operator: 'equals',
            value: 'test',
          },
          {
            field_id: 'field_2',
            operator: 'greater_than',
            value: 5,
          },
        ],
        action: 'show',
      }

      const { evaluateGroup } = useConditions(fields, formData)
      expect(evaluateGroup(group)).toBe(false)
    })
  })

  describe('Field State Calculation - Show Action', () => {
    it('should show field when condition is met with show action', () => {
      formData.value.field_1 = 'test'

      fields.value[2].condition_groups = [
        {
          logic: 'AND',
          conditions: [
            {
              field_id: 'field_1',
              operator: 'equals',
              value: 'test',
            },
          ],
          action: 'show',
        },
      ]

      const { getFieldState } = useConditions(fields, formData)
      const state = getFieldState(fields.value[2])
      expect(state.visible).toBe(true)
    })

    it('should hide field when condition is not met with show action', () => {
      formData.value.field_1 = 'different'

      fields.value[2].condition_groups = [
        {
          logic: 'AND',
          conditions: [
            {
              field_id: 'field_1',
              operator: 'equals',
              value: 'test',
            },
          ],
          action: 'show',
        },
      ]

      const { getFieldState } = useConditions(fields, formData)
      const state = getFieldState(fields.value[2])
      expect(state.visible).toBe(false)
    })
  })

  describe('Field State Calculation - Hide Action', () => {
    it('should hide field when condition is met with hide action', () => {
      formData.value.field_1 = 'test'

      fields.value[2].condition_groups = [
        {
          logic: 'AND',
          conditions: [
            {
              field_id: 'field_1',
              operator: 'equals',
              value: 'test',
            },
          ],
          action: 'hide',
        },
      ]

      const { getFieldState } = useConditions(fields, formData)
      const state = getFieldState(fields.value[2])
      expect(state.visible).toBe(false)
    })
  })

  describe('Field State Calculation - Require Action', () => {
    it('should make field required when condition is met with require action', () => {
      formData.value.field_1 = 'test'

      fields.value[2].condition_groups = [
        {
          logic: 'AND',
          conditions: [
            {
              field_id: 'field_1',
              operator: 'equals',
              value: 'test',
            },
          ],
          action: 'require',
        },
      ]

      const { getFieldState } = useConditions(fields, formData)
      const state = getFieldState(fields.value[2])
      expect(state.required).toBe(true)
    })

    it('should not make field required when condition is not met', () => {
      formData.value.field_1 = 'different'

      fields.value[2].required = false
      fields.value[2].condition_groups = [
        {
          logic: 'AND',
          conditions: [
            {
              field_id: 'field_1',
              operator: 'equals',
              value: 'test',
            },
          ],
          action: 'require',
        },
      ]

      const { getFieldState } = useConditions(fields, formData)
      const state = getFieldState(fields.value[2])
      expect(state.required).toBe(false)
    })
  })

  describe('Field State Calculation - Disable Action', () => {
    it('should disable field when condition is met with disable action', () => {
      formData.value.field_1 = 'test'

      fields.value[2].condition_groups = [
        {
          logic: 'AND',
          conditions: [
            {
              field_id: 'field_1',
              operator: 'equals',
              value: 'test',
            },
          ],
          action: 'disable',
        },
      ]

      const { getFieldState } = useConditions(fields, formData)
      const state = getFieldState(fields.value[2])
      expect(state.disabled).toBe(true)
    })
  })

  describe('Multiple Condition Groups', () => {
    it('should apply multiple condition groups correctly', () => {
      formData.value.field_1 = 'test'
      formData.value.field_2 = 10

      fields.value[2].condition_groups = [
        {
          logic: 'AND',
          conditions: [
            {
              field_id: 'field_1',
              operator: 'equals',
              value: 'test',
            },
          ],
          action: 'show',
        },
        {
          logic: 'AND',
          conditions: [
            {
              field_id: 'field_2',
              operator: 'greater_than',
              value: 5,
            },
          ],
          action: 'require',
        },
      ]

      const { getFieldState } = useConditions(fields, formData)
      const state = getFieldState(fields.value[2])
      expect(state.visible).toBe(true)
      expect(state.required).toBe(true)
    })
  })

  describe('Field States Computed', () => {
    it('should compute states for all fields', () => {
      formData.value.field_1 = 'test'

      fields.value[2].condition_groups = [
        {
          logic: 'AND',
          conditions: [
            {
              field_id: 'field_1',
              operator: 'equals',
              value: 'test',
            },
          ],
          action: 'show',
        },
      ]

      const { fieldStates } = useConditions(fields, formData)
      expect(fieldStates.value).toHaveProperty('field_1')
      expect(fieldStates.value).toHaveProperty('field_2')
      expect(fieldStates.value).toHaveProperty('field_3')
      expect(fieldStates.value.field_3.visible).toBe(true)
    })

    it('should update states reactively when formData changes', () => {
      formData.value.field_1 = 'test'

      fields.value[2].condition_groups = [
        {
          logic: 'AND',
          conditions: [
            {
              field_id: 'field_1',
              operator: 'equals',
              value: 'test',
            },
          ],
          action: 'show',
        },
      ]

      const { fieldStates } = useConditions(fields, formData)
      expect(fieldStates.value.field_3.visible).toBe(true)

      formData.value.field_1 = 'different'
      expect(fieldStates.value.field_3.visible).toBe(false)
    })
  })

  describe('Edge Cases', () => {
    it('should handle field with no condition groups', () => {
      fields.value[0].condition_groups = undefined

      const { getFieldState } = useConditions(fields, formData)
      const state = getFieldState(fields.value[0])
      expect(state.visible).toBe(true)
      expect(state.required).toBe(false)
      expect(state.disabled).toBe(false)
    })

    it('should handle field with empty condition groups array', () => {
      fields.value[0].condition_groups = []

      const { getFieldState } = useConditions(fields, formData)
      const state = getFieldState(fields.value[0])
      expect(state.visible).toBe(true)
      expect(state.required).toBe(false)
      expect(state.disabled).toBe(false)
    })

    it('should handle missing field in formData', () => {
      delete formData.value.field_1

      const condition: FieldCondition = {
        field_id: 'field_1',
        operator: 'is_empty',
        value: null,
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(true)
    })

    it('should handle unknown operator gracefully', () => {
      formData.value.field_1 = 'test'

      const condition = {
        field_id: 'field_1',
        operator: 'unknown_operator' as ConditionOperator,
        value: 'test',
      }

      const { evaluateCondition } = useConditions(fields, formData)
      expect(evaluateCondition(condition)).toBe(false)
    })
  })
})
