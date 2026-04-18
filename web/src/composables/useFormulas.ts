import { computed, type Ref, watch } from "vue";
import { Parser } from "expr-eval";
import type { Field } from "@/models/template";

// Rejects formulas containing consecutive binary `+` tokens (e.g. `a + + b`, `a++b`),
// which expr-eval otherwise accepts via unary plus and produces a surprising result.
const DOUBLE_PLUS_RE = /\+\s*\+/;

function createParser(): Parser {
  const parser = new Parser();

  parser.functions.SUM = (...args: number[]) => args.reduce((sum, val) => sum + (Number(val) || 0), 0);

  parser.functions.IF = (condition: boolean, trueVal: number, falseVal: number) => (condition ? trueVal : falseVal);

  parser.functions.MAX = (...args: number[]) => Math.max(...args.map((v) => Number(v) || 0));

  parser.functions.MIN = (...args: number[]) => Math.min(...args.map((v) => Number(v) || 0));

  parser.functions.ROUND = (value: number, decimals = 0) => {
    const multiplier = Math.pow(10, decimals);
    return Math.round(value * multiplier) / multiplier;
  };

  return parser;
}

export function useFormulas(fields: Ref<Field[]>, formData: Ref<Record<string, any>>) {
  const parser = createParser();

  function buildVariables(): Record<string, number> {
    const variables: Record<string, number> = {};
    for (const field of fields.value) {
      variables[field.id] = Number(formData.value[field.id]) || 0;
    }
    return variables;
  }

  function evaluateFormula(formula: string): number | null {
    if (!formula || DOUBLE_PLUS_RE.test(formula)) {
      return null;
    }
    try {
      const num = Number(parser.evaluate(formula, buildVariables()));
      // Preserve Infinity (e.g. division by zero) but reject NaN.
      return Number.isNaN(num) ? null : num;
    } catch (error) {
      console.error("Formula evaluation error:", error);
      return null;
    }
  }

  // Recomputed automatically when any tracked field value changes; we explicitly
  // touch every referenced field inside the computed so Vue registers dependencies.
  const calculatedValues = computed(() => {
    const values: Record<string, number> = {};
    const data = formData.value;
    for (const field of fields.value) {
      void data[field.id];
    }
    for (const field of fields.value) {
      if (!field.formula) {
        continue;
      }
      const result = evaluateFormula(field.formula);
      if (result !== null) {
        values[field.id] = result;
      }
    }
    return values;
  });

  watch(
    calculatedValues,
    (newValues) => {
      for (const [fieldId, value] of Object.entries(newValues)) {
        if (formData.value[fieldId] !== value) {
          formData.value[fieldId] = value;
        }
      }
    },
    { immediate: true }
  );

  return {
    calculatedValues,
    evaluateFormula
  };
}
