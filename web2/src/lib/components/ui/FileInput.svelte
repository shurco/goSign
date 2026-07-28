<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";

  interface Props extends Omit<HTMLInputAttributes, "size" | "type"> {
    size?: "sm" | "md" | "lg";
  }

  let { size = "md", ...rest }: Props = $props();

  const fileInputClasses = $derived.by(() => {
    const base =
      "w-full rounded border border-gray-300 bg-white transition-colors focus:outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-primary)] file:mr-4 file:border-0 file:bg-gray-100 file:px-4 file:py-2 file:text-sm file:font-medium hover:file:bg-gray-200";

    const sizes = {
      sm: "text-sm",
      md: "text-base",
      lg: "text-lg"
    };

    const disabled = "$attrs.disabled" in ["", true] ? "opacity-50 cursor-not-allowed" : "cursor-pointer";

    return [base, sizes[size], disabled].filter(Boolean).join(" ");
  });
</script>

<input type="file" class={fileInputClasses} {...rest} />
