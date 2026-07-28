<script lang="ts">
  import { getContext } from "svelte";
  import type { Snippet } from "svelte";

  interface Props {
    value: string;
    children?: Snippet;
  }

  let { value, children }: Props = $props();

  const activeTab = getContext<{ value: string }>("activeTab");
  const setActiveTab = getContext<(tabValue: string) => void>("setActiveTab");

  const isActive = $derived(activeTab.value === value);

  const tabClasses = $derived.by(() => {
    const base = "px-4 py-2 rounded-md text-sm font-medium transition-colors";
    const active = isActive ? "bg-white text-gray-900 border border-gray-200" : "text-gray-600 hover:text-gray-900";

    return [base, active].join(" ");
  });

  function handleClick(): void {
    setActiveTab?.(value);
  }
</script>

<button type="button" class={tabClasses} onclick={handleClick}>
  {@render children?.()}
</button>
