<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";

  interface Props extends Omit<HTMLInputAttributes, "checked" | "value" | "size" | "type"> {
    checked?: boolean | string[];
    value?: string;
    size?: "sm" | "md";
  }

  let { checked = $bindable(false), value, size = "md", ...rest }: Props = $props();

  const isChecked = $derived.by(() => {
    if (value !== undefined && value !== "") {
      const arr = Array.isArray(checked) ? checked : [];
      return arr.includes(value);
    }
    return checked === true;
  });

  function onChange(e: Event): void {
    const target = e.target as HTMLInputElement;
    if (value !== undefined && value !== "") {
      const arr = Array.isArray(checked) ? [...checked] : [];
      const i = arr.indexOf(value);
      if (target.checked) {
        if (i === -1) {
          arr.push(value);
        }
      } else {
        if (i !== -1) {
          arr.splice(i, 1);
        }
      }
      checked = arr;
    } else {
      checked = target.checked;
    }
  }
</script>

<input type="checkbox" class="ui-checkbox ui-checkbox-{size}" checked={isChecked} onchange={onChange} {...rest} />

<style>
  .ui-checkbox {
    accent-color: var(--base-hlt-invert);
    border-radius: var(--radius-4);
    cursor: pointer;
    transition: var(--transition-colors);
  }
  .ui-checkbox:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .ui-checkbox:focus-visible {
    outline: none;
    box-shadow: var(--shadow-focus);
  }
  .ui-checkbox-sm {
    width: 14px;
    height: 14px;
  }
  .ui-checkbox-md {
    width: 16px;
    height: 16px;
  }
</style>
