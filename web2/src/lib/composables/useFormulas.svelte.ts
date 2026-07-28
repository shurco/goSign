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

/**
 * Formula engine. `fields` and `formData` are getters backed by the caller's
 * `$state`. Calculated values are written back into the caller's formData
 * object via an effect, mirroring the previous watch-based behavior.
 * Must be called during component initialization (effect context required).
 */
export function useFormulas(fields: () => Field[], formData: () => Record<string, any>) {
  const parser = createParser();

  function buildVariables(): Record<string, number> {
    const variables: Record<string, number> = {};
    const data = formData();
    for (const field of fields()) {
      variables[field.id] = Number(data[field.id]) || 0;
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

  const calculatedValues = $derived.by(() => {
    const values: Record<string, number> = {};
    const data = formData();
    // Touch every field value so the derived tracks all inputs.
    for (const field of fields()) {
      void data[field.id];
    }
    for (const field of fields()) {
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

  $effect(() => {
    const newValues = calculatedValues;
    const data = formData();
    for (const [fieldId, value] of Object.entries(newValues)) {
      if (data[fieldId] !== value) {
        data[fieldId] = value;
      }
    }
  });

  return {
    get calculatedValues() {
      return calculatedValues;
    },
    evaluateFormula
  };
}
