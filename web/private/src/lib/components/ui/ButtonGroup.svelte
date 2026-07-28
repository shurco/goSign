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

<!-- twing-m toolbar-group: pill container, active segment — blue pill -->
<div class="btn-group" class:disabled>
  {#each normalizedOptions as option (getOptionValue(option))}
    <button
      type="button"
      class="btn-group-item"
      class:active={isSelected(option)}
      {disabled}
      onclick={() => handleClick(option)}
    >
      {#if getOptionIcon(option)}
        <SvgIcon name={getOptionIcon(option)!} width="16" height="16" />
      {/if}
      <span>{getOptionLabel(option)}</span>
    </button>
  {/each}
</div>

<style>
  .btn-group {
    display: flex;
    align-items: stretch;
    gap: var(--space-2);
    background: var(--base-cont-top);
    border: 1px solid var(--base-line-tertiary);
    border-radius: var(--radius-10);
    padding: var(--space-2);
  }
  .btn-group.disabled {
    opacity: 0.6;
  }
  .btn-group-item {
    flex: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--space-6);
    border: none;
    background: transparent;
    font-size: var(--font-size-12);
    font-weight: var(--font-weight-medium);
    color: var(--base-txt-tertiary);
    padding: 6px 10px;
    border-radius: var(--radius-8);
    cursor: pointer;
    font-family: inherit;
    white-space: nowrap;
    transition: var(--transition-colors);
  }
  .btn-group-item:hover:not(:disabled) {
    background: var(--base-surf-page);
  }
  .btn-group-item.active {
    background: var(--base-hlt-invert);
    color: var(--base-txt-btn-flip);
  }
  .btn-group-item:disabled {
    cursor: not-allowed;
  }
</style>
