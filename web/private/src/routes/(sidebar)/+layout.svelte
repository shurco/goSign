<script lang="ts">
  import type { Snippet } from "svelte";
  import RailSidebar from "@/components/layout/Sidebar.svelte";

  let { children }: { children?: Snippet } = $props();
</script>

<div class="app-shell">
  <RailSidebar />

  <main class="app-main">
    {@render children?.()}
  </main>
</div>

<style>
  /* Scroll container like the previous layout: pages (e.g. template editor)
     rely on a fixed-height parent (h-full). */
  main.app-main {
    display: block;
    height: 100vh;
    min-height: 100vh;
    overflow-y: auto;
    overflow-x: hidden;
    padding: var(--space-16) var(--space-24) var(--space-20);
  }
</style>
