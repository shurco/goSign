<script lang="ts">
  import Select from "@/components/ui/Select.svelte";
  import Radio from "@/components/ui/Radio.svelte";
  import Checkbox from "@/components/ui/Checkbox.svelte";

  interface Option {
    id?: string;
    value?: string;
    label?: string;
  }

  interface Props {
    value?: string | boolean | string[];
    type: "select" | "radio" | "checkbox" | "multiple";
    placeholder?: string;
    required?: boolean;
    disabled?: boolean;
    /** Options: array of { id?, value?, label? } or strings (normalized to objects) */
    options?: (Option | string)[];
    error?: string;
    onBlur?: () => void;
  }

  let {
    value = $bindable<string | boolean | string[]>(""),
    type = "select",
    placeholder = "",
    required = false,
    disabled = false,
    options = [],
    error = "",
    onBlur
  }: Props = $props();

  const radioGroupName = `radio-${Math.random().toString(36).slice(2, 9)}`;

  // Normalize options: accept objects { id?, value?, label? } or strings
  const normalizedOptions = $derived.by((): Option[] => {
    const raw = options ?? [];
    if (!Array.isArray(raw)) {
      return [];
    }
    return raw.map((item): Option => {
      if (typeof item === "string") {
        return { id: item, value: item, label: item };
      }
      const o = item as Option;
      const optionValue = String(o.value ?? o.id ?? "");
      return {
        id: o.id ?? optionValue,
        value: optionValue,
        label: o.label ?? optionValue
      };
    });
  });

  function handleBlur(): void {
    onBlur?.();
  }
</script>

<div class="field-input-wrapper">
  {#if type === "select"}
    <Select bind:value={value as string} {required} {disabled} onblur={handleBlur}>
      <option value="">{placeholder || "Select..."}</option>
      {#each normalizedOptions as option (option.id || option.value)}
        <option value={option.value || option.id}>{option.label || option.value}</option>
      {/each}
    </Select>
  {:else if type === "radio"}
    <div class="space-y-2">
      {#each normalizedOptions as option (option.id || option.value)}
        <label class="flex cursor-pointer items-center gap-2">
          <Radio
            bind:value={value as string}
            name={radioGroupName}
            optionValue={option.value ?? option.id ?? ""}
            {disabled}
          />
          <span>{option.label || option.value}</span>
        </label>
      {/each}
    </div>
  {:else if type === "checkbox"}
    <div class="flex items-center gap-2">
      <Checkbox bind:checked={value as boolean} {disabled} onblur={handleBlur} />
      {#if placeholder}
        <span>{placeholder}</span>
      {/if}
    </div>
  {:else if type === "multiple"}
    <div class="space-y-2">
      {#each normalizedOptions as option (option.id || option.value)}
        <label class="flex cursor-pointer items-center gap-2">
          <Checkbox bind:checked={value as string[]} value={option.value || option.id} {disabled} />
          <span>{option.label || option.value}</span>
        </label>
      {/each}
    </div>
  {/if}

  {#if error}
    <div class="mt-1 text-sm text-[var(--color-error)]">{error}</div>
  {/if}
</div>
