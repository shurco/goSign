<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    title?: string;
    value?: string | number;
    description?: string;
    variant?: "primary" | "success" | "warning" | "error" | "info";
    figure?: Snippet;
    children?: Snippet;
  }

  let { title = "", value = "", description = "", variant = "primary", figure, children }: Props = $props();

  const valueClass = $derived.by(() => {
    const variants = {
      primary: "text-[var(--color-primary)]",
      success: "text-[var(--color-success)]",
      warning: "text-[var(--color-warning)]",
      error: "text-[var(--color-error)]",
      info: "text-[var(--color-info)]"
    };

    return variants[variant];
  });
</script>

<div class="stat-item p-6">
  {#if figure}
    <div class="stat-figure mb-2">
      {@render figure()}
    </div>
  {/if}
  {#if title}
    <div class="stat-title mb-1 text-sm text-gray-600">{title}</div>
  {/if}
  {#if value}
    <div class="stat-value mb-1 text-3xl font-bold {valueClass}">{value}</div>
  {/if}
  {#if description}
    <div class="stat-desc text-sm text-gray-500">{description}</div>
  {/if}
  {@render children?.()}
</div>
