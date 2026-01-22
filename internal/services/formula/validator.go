package formula

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/antonmedv/expr"
	"github.com/shurco/gosign/internal/models"
)

// ValidateFormula validates formula syntax and field references
func ValidateFormula(formula string, fields []models.Field) error {
	if formula == "" {
		return nil
	}

	// Create field map for validation
	fieldMap := make(map[string]bool)
	for _, f := range fields {
		fieldMap[f.ID] = true
	}

	// Extract field references (e.g., field_1, field_2)
	fieldRefs := extractFieldReferences(formula)

	// Check all referenced fields exist
	for _, ref := range fieldRefs {
		if !fieldMap[ref] {
			return fmt.Errorf("formula references non-existent field: %s", ref)
		}
	}

	// Try to compile formula
	env := buildFormulaEnv(fields)
	_, err := expr.Compile(formula, expr.Env(env))
	if err != nil {
		return fmt.Errorf("invalid formula syntax: %w", err)
	}

	return nil
}

// extractFieldReferences extracts field IDs from formula
func extractFieldReferences(formula string) []string {
	// Match field IDs - can be any identifier pattern
	// First try to match field_XXX pattern, then any identifier
	re := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`)
	matches := re.FindAllStringSubmatch(formula, -1)

	// Filter out built-in functions and operators
	builtins := map[string]bool{
		"SUM": true, "IF": true, "MAX": true, "MIN": true, "ROUND": true,
		"and": true, "or": true, "not": true,
		"true": true, "false": true,
	}

	seen := make(map[string]bool)
	var unique []string
	for _, match := range matches {
		fieldID := match[1]
		// Skip if it's a builtin, number, or already seen
		if builtins[fieldID] || seen[fieldID] {
			continue
		}
		// Check if it's a number
		if _, err := strconv.ParseFloat(fieldID, 64); err == nil {
			continue
		}
		seen[fieldID] = true
		unique = append(unique, fieldID)
	}

	return unique
}

// buildFormulaEnv creates environment map for expr evaluation
func buildFormulaEnv(fields []models.Field) map[string]interface{} {
	env := make(map[string]interface{})

	for _, field := range fields {
		// Initialize all fields as float64 for calculations
		env[field.ID] = float64(0)
	}

	// Add built-in functions
	env["SUM"] = func(args ...float64) float64 {
		sum := float64(0)
		for _, v := range args {
			sum += v
		}
		return sum
	}

	env["IF"] = func(condition bool, trueVal, falseVal float64) float64 {
		if condition {
			return trueVal
		}
		return falseVal
	}

	env["MAX"] = func(args ...float64) float64 {
		if len(args) == 0 {
			return 0
		}
		max := args[0]
		for _, v := range args[1:] {
			if v > max {
				max = v
			}
		}
		return max
	}

	env["MIN"] = func(args ...float64) float64 {
		if len(args) == 0 {
			return 0
		}
		min := args[0]
		for _, v := range args[1:] {
			if v < min {
				min = v
			}
		}
		return min
	}

	env["ROUND"] = func(value float64, decimals int) float64 {
		multiplier := math.Pow(10, float64(decimals))
		return math.Round(value * multiplier) / multiplier
	}

	return env
}

// EvaluateFormula evaluates formula with field values
func EvaluateFormula(formula string, fieldValues map[string]interface{}, fields []models.Field) (float64, error) {
	if formula == "" {
		return 0, nil
	}

	env := buildFormulaEnv(fields)

	// Update env with actual values
	for fieldID, value := range fieldValues {
		// Convert to float64
		switch v := value.(type) {
		case float64:
			env[fieldID] = v
		case int:
			env[fieldID] = float64(v)
		case int64:
			env[fieldID] = float64(v)
		case string:
			// Try to parse string as number
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				env[fieldID] = parsed
			} else {
				env[fieldID] = 0
			}
		default:
			env[fieldID] = 0
		}
	}

	// Compile and evaluate
	program, err := expr.Compile(formula, expr.Env(env))
	if err != nil {
		return 0, fmt.Errorf("failed to compile formula: %w", err)
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return 0, fmt.Errorf("failed to evaluate formula: %w", err)
	}

	// Convert result to float64
	result, ok := output.(float64)
	if !ok {
		// Try to convert from int
		if intResult, ok := output.(int); ok {
			return float64(intResult), nil
		}
		return 0, fmt.Errorf("formula result is not a number: %v", output)
	}

	return result, nil
}
