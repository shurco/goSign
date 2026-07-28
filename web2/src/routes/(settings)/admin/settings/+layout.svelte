<script lang="ts">
  import type { Snippet } from "svelte";
  import AdminSettings from "@/pages/AdminSettings.svelte";

  let { children }: { children?: Snippet } = $props();
</script>

<AdminSettings>{@render children?.()}</AdminSettings>
