<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    variant?: "info" | "success" | "warning" | "error";
    class?: string;
    icon?: Snippet;
    children?: Snippet;
  }

  let { variant = "info", class: className = "", icon, children }: Props = $props();

  const alertClasses = $derived.by(() => {
    const base = "flex items-start gap-4 p-4 rounded-xl border-l-4";

    const variants = {
      info: "bg-[var(--color-info)]/10 border-[var(--color-info)] text-[var(--color-info-content)]",
      success: "bg-[var(--color-success)]/10 border-[var(--color-success)] text-[var(--color-success-content)]",
      warning: "bg-[var(--color-warning)]/10 border-[var(--color-warning)] text-[var(--color-warning-content)]",
      error: "bg-[var(--color-error)]/10 border-[var(--color-error)] text-[var(--color-error-content)]"
    };

    return [base, variants[variant], className].filter(Boolean).join(" ");
  });
</script>

<div class={alertClasses} role="alert">
  {#if icon}
    <div class="flex-shrink-0">
      {@render icon()}
    </div>
  {/if}
  <div class="flex-1">
    {@render children?.()}
  </div>
</div>
