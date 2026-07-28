<script lang="ts">
  import type { Field } from "@/models/template";

  interface Props {
    fields: Field[];
    filledFieldIds: string[];
    currentFieldId: string | null;
    getFieldLabel?: (field: Field) => string;
    onFieldSelect?: (fieldId: string) => void;
  }

  let { fields, filledFieldIds, currentFieldId, getFieldLabel, onFieldSelect }: Props = $props();

  function dotClasses(field: Field): string {
    const filled = filledFieldIds.includes(field.id);
    const current = currentFieldId === field.id;
    if (current) {
      return "bg-primary";
    }
    if (filled) {
      return "bg-success";
    }
    return "bg-neutral-400 hover:bg-neutral-500";
  }

  function fieldLabel(field: Field, index: number): string {
    return getFieldLabel ? getFieldLabel(field) : field.label || field.name || `Field ${index + 1}`;
  }

  function onSelect(fieldId: string): void {
    onFieldSelect?.(fieldId);
  }
</script>

<div class="flex flex-wrap items-center justify-center gap-1 overflow-x-auto">
  {#each fields as field, index (field.id)}
    <button
      type="button"
      class="field-dot relative h-2 w-2 shrink-0 cursor-pointer rounded-full focus:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-1 {dotClasses(
        field
      )}"
      title={fieldLabel(field, index)}
      aria-label={fieldLabel(field, index)}
      aria-current={currentFieldId === field.id ? "true" : undefined}
      onclick={() => onSelect(field.id)}
    ></button>
  {/each}
</div>
