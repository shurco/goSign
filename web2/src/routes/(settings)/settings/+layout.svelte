<script lang="ts">
  import type { Snippet } from "svelte";
  import Settings from "@/pages/Settings.svelte";

  let { children }: { children?: Snippet } = $props();
</script>

<Settings>{@render children?.()}</Settings>
