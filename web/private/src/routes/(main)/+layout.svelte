<script lang="ts">
  import type { Snippet } from "svelte";
  import Main from "@/layouts/Main.svelte";

  let { children }: { children?: Snippet } = $props();
</script>

<Main>{@render children?.()}</Main>
