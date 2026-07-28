<script lang="ts">
  import type { Snippet } from "svelte";
  import SettingsSidebar from "@/layouts/SettingsSidebar.svelte";

  let { children }: { children?: Snippet } = $props();
</script>

<SettingsSidebar>{@render children?.()}</SettingsSidebar>
