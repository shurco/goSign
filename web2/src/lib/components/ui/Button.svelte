<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLButtonAttributes } from "svelte/elements";
  import LoadingSpinner from "./LoadingSpinner.svelte";

  interface Props extends HTMLButtonAttributes {
    variant?: "primary" | "ghost" | "outline" | "success" | "warning" | "error" | "info";
    size?: "sm" | "md" | "lg";
    loading?: boolean;
    disabled?: boolean;
    circle?: boolean;
    children?: Snippet;
  }

  let {
    variant = "primary",
    size = "md",
    loading = false,
    disabled = false,
    circle = false,
    children,
    ...rest
  }: Props = $props();

  const buttonClasses = $derived.by(() => {
    const base =
      "inline-flex items-center justify-center gap-2 font-medium cursor-pointer select-none text-center transition-colors focus:outline-none no-underline normal-case";

    const variants = {
      primary: "rounded-md border border-blue-500 bg-blue-50 text-blue-700 hover:bg-blue-100",
      ghost: "rounded-md border border-gray-300 bg-white text-gray-700 hover:bg-gray-50",
      outline: "rounded-md border border-gray-300 bg-transparent text-gray-700 hover:bg-gray-50",
      success: "rounded-md border border-green-500 bg-green-50 text-green-700 hover:bg-green-100",
      warning: "rounded-md border border-yellow-500 bg-yellow-50 text-yellow-700 hover:bg-yellow-100",
      error: "rounded-md border border-red-500 bg-red-50 text-red-700 hover:bg-red-100",
      info: "rounded-md border border-blue-500 bg-blue-50 text-blue-700 hover:bg-blue-100"
    };

    const sizes = circle
      ? {
          sm: "h-8 w-8 rounded-full",
          md: "h-12 w-12 rounded-full",
          lg: "h-16 w-16 rounded-full"
        }
      : {
          sm: "px-3 py-2 text-xs rounded-md",
          md: "px-4 py-2 text-sm rounded-md",
          lg: "px-6 py-3 text-base rounded-md"
        };

    const disabledClasses = disabled || loading ? "opacity-60 cursor-not-allowed pointer-events-none" : "";

    return [base, variants[variant], sizes[size], disabledClasses].filter(Boolean).join(" ");
  });

  const spinnerSize = $derived(({ sm: "sm", md: "sm", lg: "md" } as const)[size]);
</script>

<button class={buttonClasses} disabled={disabled || loading} {...rest}>
  {#if loading}
    <LoadingSpinner size={spinnerSize} class="mr-2" />
  {/if}
  {@render children?.()}
</button>
