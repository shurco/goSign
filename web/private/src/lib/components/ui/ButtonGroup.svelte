<script lang="ts">
  import SvgIcon from "@/components/SvgIcon.svelte";

  export interface ButtonGroupOption {
    value: string | number;
    label: string;
    icon?: string;
  }

  type NormalizedOption = ButtonGroupOption;

  interface Props {
    value: string | number;
    options: ButtonGroupOption[] | string[] | Array<{ value: string | number; label: string; icon?: string }>;
    disabled?: boolean;
  }

  let { value = $bindable(), options, disabled = false }: Props = $props();

  // Normalize options to handle different input formats
  const normalizedOptions = $derived.by(() => {
    return options.map((option) => {
      if (typeof option === "string") {
        return { value: option, label: option, icon: undefined };
      }
      return option;
    });
  });

  function getOptionValue(option: NormalizedOption | string): string | number {
    if (typeof option === "string") {
      return option;
    }
    return option.value;
  }

  function getOptionLabel(option: NormalizedOption | string): string {
    if (typeof option === "string") {
      return option;
    }
    return option.label || String(option.value);
  }

  function getOptionIcon(option: NormalizedOption | string): string | undefined {
    if (typeof option === "string") {
      return undefined;
    }
    return option.icon;
  }

  function isSelected(option: NormalizedOption | string): boolean {
    return getOptionValue(option) === value;
  }

  function handleClick(option: NormalizedOption | string): void {
    if (disabled) {
      return;
    }
    value = getOptionValue(option);
  }
</script>

<div class="flex gap-2">
  {#each normalizedOptions as option (getOptionValue(option))}
    <button
      type="button"
      class="flex-1 rounded-md border px-3 py-2 text-sm font-medium transition-colors {isSelected(option)
        ? 'border-blue-500 bg-blue-50 text-blue-700'
        : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'}"
      {disabled}
      onclick={() => handleClick(option)}
    >
      <div class="flex items-center justify-center gap-2">
        {#if getOptionIcon(option)}
          <SvgIcon name={getOptionIcon(option)!} width="16" height="16" />
        {/if}
        <span>{getOptionLabel(option)}</span>
      </div>
    </button>
  {/each}
</div>
