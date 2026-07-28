package formula

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/expr-lang/expr"

	"github.com/shurco/gosign/internal/models"
)

// ValidateFormula validates formula syntax and field references
func ValidateFormula(formula string, fields []models.Field) error {
	if formula == "" {
		return nil
	}

	fieldMap := make(map[string]bool)
	for _, f := range fields {
		fieldMap[f.ID] = true
	}

	fieldRefs := extractFieldReferences(formula)
	for _, ref := range fieldRefs {
		if !fieldMap[ref] {
			return fmt.Errorf("formula references non-existent field: %s", ref)
		}
	}

	// Rewrite formula: UUIDs contain hyphens and expr parses them as minus operator.
	// Replace each UUID with a valid identifier (e.g. __f0__) before compiling.
	rewritten, placeholderToID := rewriteFormulaWithPlaceholders(formula)
	env := buildFormulaEnvWithPlaceholders(rewritten, placeholderToID)
	_, err := expr.Compile(rewritten, expr.Env(env))
	if err != nil {
		return fmt.Errorf("invalid formula syntax: %w", err)
	}

	return nil
}

// UUID pattern so UUIDs are treated as single field references (not split by hyphen).
var uuidPattern = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

// extractFieldReferences extracts field IDs from formula (full UUIDs and other identifiers).
func extractFieldReferences(formula string) []string {
	builtins := map[string]bool{
		"SUM": true, "IF": true, "MAX": true, "MIN": true, "ROUND": true,
		"and": true, "or": true, "not": true,
		"true": true, "false": true,
	}
	seen := make(map[string]bool)
	var unique []string

	// 1) Extract full UUIDs first (otherwise hyphen would break them into invalid refs like "e5f3").
	for _, m := range uuidPattern.FindAllString(formula, -1) {
		if !seen[m] {
			seen[m] = true
			unique = append(unique, m)
		}
	}
	// 2) Mask UUIDs so the identifier regex doesn't match UUID segments.
	formulaMasked := uuidPattern.ReplaceAllString(formula, " ")

	// 3) Extract other identifiers (e.g. field_1, custom names).
	re := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`)
	for _, match := range re.FindAllStringSubmatch(formulaMasked, -1) {
		fieldID := match[1]
		if builtins[fieldID] || seen[fieldID] {
			continue
		}
		if _, err := strconv.ParseFloat(fieldID, 64); err == nil {
			continue
		}
		seen[fieldID] = true
		unique = append(unique, fieldID)
	}

	return unique
}

// rewriteFormulaWithPlaceholders replaces each UUID in formula with a valid identifier
// (expr does not accept hyphens in identifiers; they are parsed as minus).
// Returns rewritten formula and map placeholder -> field ID for building env.
func rewriteFormulaWithPlaceholders(formula string) (string, map[string]string) {
	// Collect UUIDs in formula in order of first occurrence
	uuids := uuidPattern.FindAllString(formula, -1)
	seen := make(map[string]bool)
	var order []string
	for _, u := range uuids {
		if !seen[u] {
			seen[u] = true
			order = append(order, u)
		}
	}
	placeholderToID := make(map[string]string)
	rewritten := formula
	for i, id := range order {
		placeholder := fmt.Sprintf("__f%d__", i)
		placeholderToID[placeholder] = id
		// Replace all occurrences of this UUID (use regex to match full UUID)
		re := regexp.MustCompile(regexp.QuoteMeta(id))
		rewritten = re.ReplaceAllString(rewritten, placeholder)
	}
	return rewritten, placeholderToID
}

// buildFormulaEnvWithPlaceholders builds env for expr with zero values so compile succeeds:
// placeholders (rewritten UUIDs) and any remaining identifiers (e.g. field_1) map to 0.
func buildFormulaEnvWithPlaceholders(rewritten string, placeholderToID map[string]string) map[string]any {
	env := make(map[string]any)
	for placeholder := range placeholderToID {
		env[placeholder] = float64(0)
	}
	for _, id := range extractIdentifiers(rewritten) {
		if _, ok := env[id]; ok {
			continue
		}
		env[id] = float64(0)
	}
	addBuiltinFunctions(env)
	return env
}

// extractIdentifiers returns non-builtin identifiers from a formula (no UUIDs; for use on rewritten formula).
func extractIdentifiers(formula string) []string {
	builtins := map[string]bool{
		"SUM": true, "IF": true, "MAX": true, "MIN": true, "ROUND": true,
		"and": true, "or": true, "not": true, "true": true, "false": true,
	}
	re := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`)
	matches := re.FindAllStringSubmatch(formula, -1)
	seen := make(map[string]bool)
	var out []string
	for _, m := range matches {
		id := m[1]
		if builtins[id] || seen[id] {
			continue
		}
		if _, err := strconv.ParseFloat(id, 64); err == nil {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func addBuiltinFunctions(env map[string]any) {
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
		return math.Round(value*multiplier) / multiplier
	}
}
