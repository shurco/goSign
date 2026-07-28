import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { flushSync } from "svelte";
import { useFormulas } from "@/composables/useFormulas.svelte";
import type { Field } from "@/models/template";

describe("useFormulas - Formula Parsing and Calculation", () => {
  let fields = $state<Field[]>([]);
  let formData = $state<Record<string, any>>({});
  let cleanup: (() => void) | undefined;

  beforeEach(() => {
    // Suppress console.error in tests
    vi.spyOn(console, "error").mockImplementation(() => {});

    fields = [
      {
        id: "field_1",
        type: "number",
        name: "Field 1",
        required: false,
        submitter_id: ""
      },
      {
        id: "field_2",
        type: "number",
        name: "Field 2",
        required: false,
        submitter_id: ""
      },
      {
        id: "field_3",
        type: "number",
        name: "Field 3",
        required: false,
        submitter_id: ""
      },
      {
        id: "calculated_field",
        type: "number",
        name: "Calculated Field",
        required: false,
        formula: "field_1 + field_2",
        submitter_id: ""
      }
    ];

    formData = {
      field_1: 10,
      field_2: 20,
      field_3: 5,
      calculated_field: 0
    };
  });

  afterEach(() => {
    cleanup?.();
    cleanup = undefined;
  });

  describe("Basic Arithmetic Operations", () => {
    it("should evaluate addition", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2");
      expect(result).toBe(30);
    });

    it("should evaluate subtraction", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 - field_2");
      expect(result).toBe(-10);
    });

    it("should evaluate multiplication", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 * field_2");
      expect(result).toBe(200);
    });

    it("should evaluate division", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_2 / field_1");
      expect(result).toBe(2);
    });

    it("should handle division by zero", () => {
      formData.field_1 = 0;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_2 / field_1");
      expect(result).toBe(Infinity);
    });

    it("should evaluate complex expressions with parentheses", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("(field_1 + field_2) * field_3");
      expect(result).toBe(150); // (10 + 20) * 5 = 150
    });

    it("should handle order of operations", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2 * field_3");
      expect(result).toBe(110); // 10 + (20 * 5) = 110
    });
  });

  describe("SUM Function", () => {
    it("should sum multiple values", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("SUM(field_1, field_2, field_3)");
      expect(result).toBe(35); // 10 + 20 + 5
    });

    it("should sum two values", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("SUM(field_1, field_2)");
      expect(result).toBe(30);
    });

    it("should handle non-numeric values in SUM", () => {
      formData.field_1 = "not a number";
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("SUM(field_1, field_2)");
      expect(result).toBe(20); // 0 + 20
    });

    it("should return 0 for empty SUM", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("SUM()");
      expect(result).toBe(0);
    });
  });

  describe("IF Function", () => {
    it("should return true value when condition is true", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("IF(field_1 > 5, 100, 0)");
      expect(result).toBe(100);
    });

    it("should return false value when condition is false", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("IF(field_1 < 5, 100, 0)");
      expect(result).toBe(0);
    });

    it("should handle equality condition", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("IF(field_1 == 10, 50, 0)");
      expect(result).toBe(50);
    });

    it("should handle nested IF statements", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("IF(field_1 > 5, IF(field_2 > 15, 200, 100), 0)");
      expect(result).toBe(200);
    });
  });

  describe("MAX Function", () => {
    it("should return maximum value", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("MAX(field_1, field_2, field_3)");
      expect(result).toBe(20);
    });

    it("should handle negative values", () => {
      formData.field_1 = -10;
      formData.field_2 = -5;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("MAX(field_1, field_2)");
      expect(result).toBe(-5);
    });

    it("should handle non-numeric values in MAX", () => {
      formData.field_1 = "not a number";
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("MAX(field_1, field_2)");
      expect(result).toBe(20); // max(0, 20)
    });
  });

  describe("MIN Function", () => {
    it("should return minimum value", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("MIN(field_1, field_2, field_3)");
      expect(result).toBe(5);
    });

    it("should handle negative values", () => {
      formData.field_1 = -10;
      formData.field_2 = -5;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("MIN(field_1, field_2)");
      expect(result).toBe(-10);
    });
  });

  describe("ROUND Function", () => {
    it("should round to nearest integer by default", () => {
      formData.field_1 = 10.7;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("ROUND(field_1)");
      expect(result).toBe(11);
    });

    it("should round to specified decimal places", () => {
      formData.field_1 = 10.567;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("ROUND(field_1, 2)");
      expect(result).toBe(10.57);
    });

    it("should round down when appropriate", () => {
      formData.field_1 = 10.4;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("ROUND(field_1)");
      expect(result).toBe(10);
    });

    it("should handle zero decimal places", () => {
      formData.field_1 = 10.567;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("ROUND(field_1, 0)");
      expect(result).toBe(11);
    });
  });

  describe("Variable Substitution", () => {
    it("should substitute field values from formData", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2");
      expect(result).toBe(30);
    });

    it("should handle missing field values as zero", () => {
      delete formData.field_1;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2");
      expect(result).toBe(20); // 0 + 20
    });

    it("should handle string numbers", () => {
      formData.field_1 = "10";
      formData.field_2 = "20";
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2");
      expect(result).toBe(30);
    });

    it("should handle non-numeric strings as zero", () => {
      formData.field_1 = "not a number";
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2");
      expect(result).toBe(20); // 0 + 20
    });
  });

  describe("Calculated Values Computed", () => {
    it("should compute values for fields with formulas", () => {
      let formulas: ReturnType<typeof useFormulas>;
      cleanup = $effect.root(() => {
        formulas = useFormulas(
          () => fields,
          () => formData
        );
      });
      flushSync();
      expect(formulas!.calculatedValues).toHaveProperty("calculated_field");
      expect(formulas!.calculatedValues.calculated_field).toBe(30);
    });

    it("should not compute values for fields without formulas", () => {
      let formulas: ReturnType<typeof useFormulas>;
      cleanup = $effect.root(() => {
        formulas = useFormulas(
          () => fields,
          () => formData
        );
      });
      flushSync();
      expect(formulas!.calculatedValues).not.toHaveProperty("field_1");
      expect(formulas!.calculatedValues).not.toHaveProperty("field_2");
    });

    it("should update calculated values when formData changes", () => {
      let formulas: ReturnType<typeof useFormulas>;
      cleanup = $effect.root(() => {
        formulas = useFormulas(
          () => fields,
          () => formData
        );
      });
      flushSync();
      expect(formulas!.calculatedValues.calculated_field).toBe(30);

      formData.field_1 = 50;
      flushSync();
      expect(formulas!.calculatedValues.calculated_field).toBe(70);
    });

    it("should handle multiple calculated fields", () => {
      fields.push({
        id: "calculated_field_2",
        type: "number",
        name: "Calculated Field 2",
        required: false,
        formula: "field_2 * field_3",
        submitter_id: ""
      });

      let formulas: ReturnType<typeof useFormulas>;
      cleanup = $effect.root(() => {
        formulas = useFormulas(
          () => fields,
          () => formData
        );
      });
      flushSync();
      expect(formulas!.calculatedValues.calculated_field).toBe(30);
      expect(formulas!.calculatedValues.calculated_field_2).toBe(100);
    });
  });

  describe("Auto-update formData", () => {
    it("should automatically update formData with calculated values", () => {
      cleanup = $effect.root(() => {
        useFormulas(
          () => fields,
          () => formData
        );
      });
      flushSync();
      expect(formData.calculated_field).toBe(30);
    });

    it("should update formData when calculated values change", () => {
      cleanup = $effect.root(() => {
        useFormulas(
          () => fields,
          () => formData
        );
      });
      flushSync();

      formData.field_1 = 100;

      flushSync();
      expect(formData.calculated_field).toBe(120);
    });
  });

  describe("Error Handling", () => {
    it("should return null for invalid formula syntax", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + + field_2");
      expect(result).toBeNull();
    });

    it("should return null for missing closing parenthesis", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("SUM(field_1, field_2");
      expect(result).toBeNull();
    });

    it("should return null for undefined function", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("UNKNOWN_FUNCTION(field_1)");
      expect(result).toBeNull();
    });

    it("should handle empty formula string", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("");
      expect(result).toBeNull();
    });

    it("should not throw error for invalid formula in calculatedValues", () => {
      fields[3].formula = "invalid formula syntax";

      let formulas: ReturnType<typeof useFormulas>;
      cleanup = $effect.root(() => {
        formulas = useFormulas(
          () => fields,
          () => formData
        );
      });
      flushSync();
      expect(formulas!.calculatedValues.calculated_field).toBeUndefined();
    });
  });

  describe("Complex Formulas", () => {
    it("should evaluate nested function calls", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("SUM(field_1, field_2) * field_3");
      expect(result).toBe(150); // (10 + 20) * 5
    });

    it("should evaluate formula with multiple functions", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("MAX(field_1, field_2) + MIN(field_1, field_2)");
      expect(result).toBe(30); // 20 + 10
    });

    it("should evaluate conditional formula with arithmetic", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("IF(field_1 > 5, field_1 * 2, field_1 + 5)");
      expect(result).toBe(20); // 10 * 2
    });

    it("should handle percentage calculations", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 * 0.1");
      expect(result).toBe(1); // 10 * 0.1
    });

    it("should handle power operations", () => {
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 ^ 2");
      expect(result).toBe(100); // 10^2
    });
  });

  describe("Edge Cases", () => {
    it("should handle null values in formData", () => {
      formData.field_1 = null;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2");
      expect(result).toBe(20); // 0 + 20
    });

    it("should handle undefined values in formData", () => {
      formData.field_1 = undefined;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2");
      expect(result).toBe(20); // 0 + 20
    });

    it("should handle empty string values", () => {
      formData.field_1 = "";
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2");
      expect(result).toBe(20); // 0 + 20
    });

    it("should handle very large numbers", () => {
      formData.field_1 = 1e10;
      formData.field_2 = 2e10;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2");
      expect(result).toBe(3e10);
    });

    it("should handle very small numbers", () => {
      formData.field_1 = 0.0001;
      formData.field_2 = 0.0002;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2");
      expect(result).toBeCloseTo(0.0003, 5);
    });

    it("should handle negative numbers", () => {
      formData.field_1 = -10;
      formData.field_2 = -20;
      let evaluateFormula: ReturnType<typeof useFormulas>["evaluateFormula"];
      cleanup = $effect.root(() => {
        ({ evaluateFormula } = useFormulas(
          () => fields,
          () => formData
        ));
      });
      flushSync();
      const result = evaluateFormula!("field_1 + field_2");
      expect(result).toBe(-30);
    });
  });
});
