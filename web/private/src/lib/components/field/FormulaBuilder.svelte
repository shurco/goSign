<script lang="ts">
  import type { Field } from "@/models/template";
  import { useFormulas } from "@/composables/useFormulas.svelte";
  import { apiPost } from "@/services/api";
  import { t } from "@/i18n/index.svelte";

  interface FieldWithDisplayName extends Field {
    displayName?: string;
  }

  interface Props {
    field: Field;
    availableFields: FieldWithDisplayName[];
    onUpdateFormula?: (formula: string) => void;
  }

  let { field, availableFields, onUpdateFormula }: Props = $props();

  // Intentional initial snapshot: local editable copy, synced back via onUpdateFormula
  // svelte-ignore state_referenced_locally
  let formula = $state(field.preferences?.formula ?? (field as Field & { formula?: string }).formula ?? "");
  let validationError = $state<string | null>(null);

  function formulaToDisplay(formulaStr: string): string {
    let out = formulaStr;
    const escapeRe = (s: string) => s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    const sorted = [...availableFields].sort((a, b) => b.id.length - a.id.length);
    for (const f of sorted) {
      const name = f.displayName ?? f.name ?? f.id;
      const re = new RegExp(escapeRe(f.id), "g");
      out = out.replace(re, `[[${name}]]`);
    }
    return out;
  }

  function displayToFormula(displayStr: string): string {
    const re = /\[\[([^\]]*?)\]\]/g;
    return displayStr.replace(re, (_, name: string) => {
      const f = availableFields.find((x) => (x.displayName ?? x.name ?? x.id) === name);
      return f ? f.id : `[[${name}]]`;
    });
  }

  const displayFormula = $derived(formulaToDisplay(formula));

  function onFormulaDisplayInput(e: Event): void {
    const target = e.target as HTMLTextAreaElement;
    formula = displayToFormula(target.value);
  }

  const availableFunctions = [
    { name: "SUM", syntax: "SUM(field_1, field_2)", description: "Sum of multiple fields" },
    { name: "IF", syntax: "IF(field_1 > 100, field_2, 0)", description: "Conditional value" },
    { name: "MAX", syntax: "MAX(field_1, field_2)", description: "Maximum value" },
    { name: "MIN", syntax: "MIN(field_1, field_2)", description: "Minimum value" },
    { name: "ROUND", syntax: "ROUND(field_1, 2)", description: "Round to decimals" }
  ];

  const examples = [
    { label: "Sum two fields", formula: "field_1 + field_2" },
    { label: "Calculate tax (20%)", formula: "field_1 * 1.2" },
    { label: "Conditional discount", formula: "IF(field_1 > 1000, field_1 * 0.9, field_1)" },
    { label: "Sum with tax", formula: "SUM(field_1, field_2) * 1.2" }
  ];

  const sampleFormData = $derived.by(() => {
    const data: Record<string, number> = {};
    for (const f of availableFields) {
      data[f.id] = 10;
    }
    return data;
  });

  const { evaluateFormula } = useFormulas(
    () => availableFields,
    () => sampleFormData
  );

  const previewResult = $derived.by(() => {
    if (!formula || validationError) {
      return null;
    }
    const result = evaluateFormula(formula);
    return result !== null ? result.toFixed(2) : null;
  });

  export function getFormula(): string {
    return formula;
  }

  export { formula };

  function insertField(fieldId: string): void {
    formula += fieldId;
    validateFormula();
  }

  function applyExample(exampleFormula: string): void {
    formula = exampleFormula;
    validateFormula();
  }

  function exampleDisplayFormula(exampleFormula: string): string {
    return formulaToDisplay(exampleFormula);
  }

  function insertFunction(syntax: string): void {
    formula += syntax;
    validateFormula();
  }

  async function validateFormula(): Promise<void> {
    if (!formula.trim()) {
      validationError = null;
      onUpdateFormula?.("");
      return;
    }

    try {
      const response = await apiPost("/templates/formulas/validate", {
        formula,
        fields: availableFields
      });

      if (response && !response.message?.includes("error")) {
        validationError = null;
        onUpdateFormula?.(formula);
      } else {
        validationError = response.message || "Invalid formula";
      }
    } catch (error: unknown) {
      validationError = error instanceof Error ? error.message : "Invalid formula";
    }
  }

  $effect(() => {
    validateFormula();
  });
</script>

<div class="formula-builder space-y-5">
  <p class="text-sm text-gray-600">
    {t("fields.formula.description") ||
      "Use field IDs and operators to compute a value. Click fields and functions below to insert."}
  </p>

  <div class="formula-editor">
    <label class="mb-1.5 block text-sm font-medium text-gray-700">
      {t("fields.formula.expression") || "Formula"}
    </label>
    <textarea
      value={displayFormula}
      placeholder={t("fields.formula.placeholder")}
      class="formula-input w-full rounded-xl border px-4 py-3 font-mono text-sm leading-relaxed transition-colors focus:ring-2 focus:outline-none {validationError
        ? 'border-red-300 bg-red-50/30 focus:border-red-400 focus:ring-red-200'
        : 'border-gray-300 bg-white focus:border-indigo-400 focus:ring-indigo-200'}"
      rows="4"
      spellcheck="false"
      oninput={onFormulaDisplayInput}></textarea>
    {#if validationError}
      <div class="mt-2 flex items-center gap-2 text-sm text-red-600">
        <span aria-hidden="true">⊗</span>
        <span>{validationError}</span>
      </div>
    {:else if previewResult !== null}
      <div class="mt-2 flex items-center gap-2 text-sm text-emerald-600">
        <span aria-hidden="true">✓</span>
        <span>{t("fields.formula.preview")}: <strong>{previewResult}</strong></span>
      </div>
    {/if}
  </div>

  <section class="rounded-xl border border-gray-200 bg-gray-50/50 p-4">
    <h4 class="mb-2 text-xs font-semibold tracking-wide text-gray-500 uppercase">
      {t("fields.formula.insertField")}
    </h4>
    <div class="flex flex-wrap gap-2">
      {#each availableFields as f (f.id)}
        <button
          type="button"
          class="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm text-gray-700 shadow-sm transition-colors hover:border-indigo-200 hover:bg-indigo-50 hover:text-indigo-700"
          onclick={() => insertField(f.id)}
        >
          {f.displayName ?? f.name ?? f.id}
        </button>
      {/each}
    </div>
    {#if !availableFields.length}
      <p class="text-sm text-gray-500">
        {t("fields.formula.noFields") || "No number/text fields available."}
      </p>
    {/if}
  </section>

  <section class="rounded-xl border border-gray-200 bg-gray-50/50 p-4">
    <h4 class="mb-2 text-xs font-semibold tracking-wide text-gray-500 uppercase">
      {t("fields.formula.functions")}
    </h4>
    <div class="flex flex-wrap gap-2">
      {#each availableFunctions as func (func.name)}
        <button
          type="button"
          class="rounded-lg border border-gray-200 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 shadow-sm transition-colors hover:border-indigo-200 hover:bg-indigo-50 hover:text-indigo-700"
          title={func.description}
          onclick={() => insertFunction(func.syntax)}
        >
          {func.name}
        </button>
      {/each}
    </div>
  </section>

  <section class="rounded-xl border border-gray-200 bg-gray-50/50 p-4">
    <h4 class="mb-2 text-xs font-semibold tracking-wide text-gray-500 uppercase">
      {t("fields.formula.examples")}
    </h4>
    <div class="space-y-1.5">
      {#each examples as example (example.label)}
        <button
          type="button"
          class="flex w-full items-start gap-3 rounded-lg border border-transparent p-2.5 text-left text-sm transition-colors hover:border-gray-200 hover:bg-white"
          onclick={() => applyExample(example.formula)}
        >
          <code class="shrink-0 rounded bg-gray-200 px-2 py-0.5 font-mono text-xs text-indigo-600">
            {exampleDisplayFormula(example.formula)}
          </code>
          <span class="text-gray-600">{example.label}</span>
        </button>
      {/each}
    </div>
  </section>
</div>

<style>
  .formula-input {
    font-family: "Courier New", monospace;
  }
</style>
