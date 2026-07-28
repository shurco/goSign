<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLSelectAttributes } from "svelte/elements";

  interface Props extends Omit<HTMLSelectAttributes, "size" | "value"> {
    value?: string | number;
    error?: boolean;
    size?: "sm" | "md" | "lg";
    children?: Snippet;
  }

  let { value = $bindable(), error = false, size = "md", children, ...rest }: Props = $props();

  // Arrow icon style (same as form.css)
  const arrowStyle = {
    backgroundImage: `url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e")`,
    backgroundPosition: "right 0.5rem center",
    backgroundRepeat: "no-repeat",
    backgroundSize: "1.5em 1.5em"
  };

  const selectClasses = $derived.by(() => {
    const base =
      "w-full rounded border bg-white transition-colors focus:outline-none focus:ring-2 cursor-pointer appearance-none pr-10";

    const borderColor = error
      ? "border-[var(--color-error)] focus:ring-[var(--color-error)]"
      : "border-gray-300 focus:border-[var(--color-primary)] focus:ring-[var(--color-primary)]";

    const sizes = {
      sm: "px-2 py-1 text-sm h-8",
      md: "px-3 py-2 text-base h-12",
      lg: "px-4 py-3 text-lg h-14"
    };

    const disabled = "$attrs.disabled" in ["", true] ? "opacity-50 cursor-not-allowed bg-gray-100" : "";

    return [base, borderColor, sizes[size], disabled].filter(Boolean).join(" ");
  });
</script>

<select
  class={selectClasses}
  style:background-image={arrowStyle.backgroundImage}
  style:background-position={arrowStyle.backgroundPosition}
  style:background-repeat={arrowStyle.backgroundRepeat}
  style:background-size={arrowStyle.backgroundSize}
  {value}
  onchange={(e) => (value = e.currentTarget.value)}
  {...rest}
>
  {@render children?.()}
</select>
