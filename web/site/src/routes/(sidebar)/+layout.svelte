<script lang="ts">
  import type { Snippet } from "svelte";
  import Sidebar from "@/layouts/Sidebar.svelte";

  let { children }: { children?: Snippet } = $props();
</script>

<Sidebar>{@render children?.()}</Sidebar>
