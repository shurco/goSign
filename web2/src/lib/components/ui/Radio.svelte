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

  const radioClasses = $derived.by(() => {
    const base =
      "border-gray-300 text-[var(--color-primary)] focus:ring-[var(--color-primary)] focus:ring-2 transition-colors cursor-pointer";

    const sizes = {
      sm: "h-4 w-4",
      md: "h-5 w-5"
    };

    const disabled = "$attrs.disabled" in ["", true] ? "opacity-50 cursor-not-allowed" : "";

    return [base, sizes[size], disabled].filter(Boolean).join(" ");
  });
</script>

<input type="radio" class={radioClasses} {name} checked={isChecked} value={optionValue} onchange={onChange} {...rest} />
