package field

import (
	"fmt"
	"github.com/shurco/gosign/internal/models"
)

// ValidateConditions validates field conditions for logical errors
func ValidateConditions(fields []models.Field) error {
	fieldMap := make(map[string]models.Field)
	for _, f := range fields {
		fieldMap[f.ID] = f
	}
	
	for _, field := range fields {
		for _, group := range field.ConditionGroups {
			for _, cond := range group.Conditions {
				// Check if referenced field exists
				if _, exists := fieldMap[cond.FieldID]; !exists {
					return fmt.Errorf("field %s references non-existent field %s", 
						field.ID, cond.FieldID)
				}
				
				// Prevent circular dependencies
				if cond.FieldID == field.ID {
					return fmt.Errorf("field %s cannot depend on itself", field.ID)
				}
				
				// Check for valid operator for field type
				targetField := fieldMap[cond.FieldID]
				if err := validateOperatorForType(targetField.Type, cond.Operator); err != nil {
					return fmt.Errorf("invalid operator %s for field type %s: %w",
						cond.Operator, targetField.Type, err)
				}
			}
		}
	}
	
	return nil
}

func validateOperatorForType(fieldType models.FieldType, operator models.ConditionOperator) error {
	// Numeric fields can use comparison operators
	if fieldType == models.FieldTypeNumber {
		validOps := []models.ConditionOperator{
			models.ConditionEquals, models.ConditionNotEquals,
			models.ConditionGreaterThan, models.ConditionLessThan,
			models.ConditionIsEmpty, models.ConditionIsNotEmpty,
		}
		if !contains(validOps, operator) {
			return fmt.Errorf("operator %s not valid for numeric fields", operator)
		}
	}
	
	// Checkbox can only use equals/not equals
	if fieldType == models.FieldTypeCheckbox {
		if operator != models.ConditionEquals && operator != models.ConditionNotEquals {
			return fmt.Errorf("checkbox fields only support equals/not_equals")
		}
	}
	
	// Text fields can use string operators
	if fieldType == models.FieldTypeText {
		validOps := []models.ConditionOperator{
			models.ConditionEquals, models.ConditionNotEquals,
			models.ConditionContains, models.ConditionNotContains,
			models.ConditionIsEmpty, models.ConditionIsNotEmpty,
		}
		if !contains(validOps, operator) {
			return fmt.Errorf("operator %s not valid for text fields", operator)
		}
	}
	
	return nil
}

func contains(slice []models.ConditionOperator, item models.ConditionOperator) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
