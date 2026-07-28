<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";

  interface Props extends Omit<HTMLInputAttributes, "size" | "type"> {
    size?: "xs" | "sm" | "md" | "lg";
  }

  let { size = "md", ...rest }: Props = $props();

  const switchClasses = $derived.by(() => {
    const base = "toggle";

    const sizes = {
      xs: "toggle-xs",
      sm: "toggle-sm",
      md: "toggle-md",
      lg: "toggle-lg"
    };

    return [base, sizes[size]].filter(Boolean).join(" ");
  });
</script>

<input type="checkbox" class={switchClasses} {...rest} />
