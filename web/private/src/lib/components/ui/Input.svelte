<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";

  interface Props extends Omit<HTMLInputAttributes, "size" | "type" | "value"> {
    value?: string | number;
    type?: "text" | "email" | "password" | "number" | "date" | "tel" | "url" | "search" | "color";
    error?: boolean;
    size?: "sm" | "md" | "lg";
  }

  let { value = $bindable(""), type = "text", error = false, size = "md", ...rest }: Props = $props();

  const inputClasses = $derived.by(() => {
    const base =
      "w-full rounded-md border border-gray-300 shadow-sm bg-[var(--color-base-100)] text-[var(--color-base-content)] transition-all duration-200 focus:outline-none focus:outline-offset-2 focus:ring-2 focus:ring-[var(--color-primary)]";

    const borderColor = error
      ? "border-[var(--color-error)] focus:border-[var(--color-error)] focus:ring-[var(--color-error)]"
      : "hover:border-[var(--color-base-content)]/20 focus:border-[var(--color-primary)]";

    const sizes = {
      sm: "h-8 px-3 text-xs min-h-[2rem]",
      md: "h-12 px-4 text-sm min-h-[3rem]",
      lg: "h-16 px-4 text-base min-h-[4rem]"
    };

    return [base, borderColor, sizes[size]].filter(Boolean).join(" ");
  });
</script>

<!-- Manual value sync: Svelte forbids bind:value together with a dynamic `type` -->
<input class={inputClasses} {type} {value} oninput={(e) => (value = e.currentTarget.value)} {...rest} />
