<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";

  interface Props extends Omit<HTMLInputAttributes, "checked" | "value" | "size" | "type" | "name"> {
    value?: string;
    optionValue?: string;
    name?: string;
    size?: "sm" | "md";
  }

  let { value = $bindable(""), optionValue = "", name = "", size = "md", ...rest }: Props = $props();

  const isChecked = $derived.by(() => {
    const current = value != null ? String(value) : "";
    const opt = optionValue != null ? String(optionValue) : "";
    return current === opt;
  });

  function onChange(e: Event): void {
    const target = e.target as HTMLInputElement;
    if (target.checked) {
      value = optionValue != null ? String(optionValue) : "";
    }
  }
</script>

<input
  type="radio"
  class="ui-radio ui-radio-{size}"
  {name}
  checked={isChecked}
  value={optionValue}
  onchange={onChange}
  {...rest}
/>

<style>
  .ui-radio {
    accent-color: var(--base-hlt-invert);
    cursor: pointer;
    transition: var(--transition-colors);
  }
  .ui-radio:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .ui-radio:focus-visible {
    outline: none;
    box-shadow: var(--shadow-focus);
  }
  .ui-radio-sm {
    width: 14px;
    height: 14px;
  }
  .ui-radio-md {
    width: 16px;
    height: 16px;
  }
</style>
