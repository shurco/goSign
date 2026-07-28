<script lang="ts">
  import type { Snippet } from "svelte";
  import Blank from "@/layouts/Blank.svelte";

  let { children }: { children?: Snippet } = $props();
</script>

<Blank>{@render children?.()}</Blank>
