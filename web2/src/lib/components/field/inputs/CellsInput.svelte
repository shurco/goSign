<script lang="ts">
  import { tick } from "svelte";

  interface Props {
    value?: string;
    cellCount?: number;
    readonly?: boolean;
    disabled?: boolean;
    error?: string;
  }

  let { value = $bindable(""), cellCount = 6, readonly = false, disabled = false, error = "" }: Props = $props();

  let cellInputs = $state<(HTMLInputElement | null)[]>([]);
  let cellValues = $state<string[]>([]);

  // Initialize cell values from value
  function initializeCells(): void {
    const current = value || "";
    cellValues = Array(cellCount)
      .fill("")
      .map((_, i) => current[i] || "");
  }

  $effect(() => {
    initializeCells();
  });

  $effect(() => {
    const external = value;
    if (external !== getCombinedValue()) {
      initializeCells();
    }
  });

  function getCombinedValue(): string {
    return cellValues.join("");
  }

  function handleCellInput(index: number, event: Event): void {
    const input = event.target as HTMLInputElement;
    const char = input.value.trim().slice(-1); // Take only last character

    // Update the cell value
    cellValues[index] = char;

    // Emit combined value
    value = getCombinedValue();

    // Auto-advance to next cell if value entered
    if (char && index < cellCount - 1) {
      tick().then(() => {
        cellInputs[index + 1]?.focus();
      });
    }
  }

  function handleKeyDown(index: number, event: KeyboardEvent): void {
    // Handle backspace
    if (event.key === "Backspace" && !cellValues[index] && index > 0) {
      event.preventDefault();
      cellValues[index - 1] = "";
      cellInputs[index - 1]?.focus();
      value = getCombinedValue();
    }

    // Handle arrow keys
    if (event.key === "ArrowLeft" && index > 0) {
      event.preventDefault();
      cellInputs[index - 1]?.focus();
    }
    if (event.key === "ArrowRight" && index < cellCount - 1) {
      event.preventDefault();
      cellInputs[index + 1]?.focus();
    }

    // Only allow alphanumeric characters
    if (event.key.length === 1 && !/[a-zA-Z0-9]/.test(event.key)) {
      event.preventDefault();
    }
  }

  function handlePaste(index: number, event: ClipboardEvent): void {
    event.preventDefault();
    const pastedText = event.clipboardData?.getData("text") || "";
    const chars = pastedText.slice(0, cellCount - index).split("");

    chars.forEach((char, i) => {
      const cellIndex = index + i;
      if (cellIndex < cellCount && /[a-zA-Z0-9]/.test(char)) {
        cellValues[cellIndex] = char;
      }
    });

    value = getCombinedValue();

    // Focus the next empty cell or last cell
    const nextIndex = Math.min(index + chars.length, cellCount - 1);
    tick().then(() => {
      cellInputs[nextIndex]?.focus();
    });
  }

  function handleFocus(index: number): void {
    // Select all text when focusing
    tick().then(() => {
      cellInputs[index]?.select();
    });
  }
</script>

<div class="field-input-wrapper">
  <div class="flex items-center gap-1">
    {#each Array(cellCount) as _, index (index)}
      <input
        bind:this={cellInputs[index]}
        value={cellValues[index]}
        type="text"
        maxlength="1"
        class="h-12 w-12 rounded border border-gray-300 text-center text-lg font-semibold focus:border-primary focus:ring-2 focus:ring-primary focus:outline-none {error
          ? 'border-error'
          : ''}"
        {disabled}
        {readonly}
        oninput={(e) => handleCellInput(index, e)}
        onkeydown={(e) => handleKeyDown(index, e)}
        onpaste={(e) => handlePaste(index, e)}
        onfocus={() => handleFocus(index)}
      />
    {/each}
  </div>
  {#if error}
    <div class="mt-1 text-sm text-[var(--color-error)]">{error}</div>
  {/if}
</div>
