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

  const checkboxClasses = $derived.by(() => {
    const base =
      "rounded border-gray-300 text-[var(--color-primary)] focus:ring-[var(--color-primary)] focus:ring-2 transition-colors cursor-pointer";

    const sizes = {
      sm: "h-4 w-4",
      md: "h-5 w-5"
    };

    const disabled = rest.disabled ? "opacity-50 cursor-not-allowed" : "";

    return [base, sizes[size], disabled].filter(Boolean).join(" ");
  });
</script>

<input type="checkbox" class={checkboxClasses} checked={isChecked} onchange={onChange} {...rest} />
