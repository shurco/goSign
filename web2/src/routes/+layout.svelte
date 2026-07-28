<script lang="ts">
  import { onMount } from "svelte";
  import type { Snippet } from "svelte";
  import * as NProgress from "nprogress";
  import { beforeNavigate, afterNavigate, goto } from "$app/navigation";
  import { setAuthNavigate } from "@/utils/auth";
  import "@/assets/app.css";

  let { children }: { children?: Snippet } = $props();

  // Let auth utilities redirect via the SPA router instead of full reloads
  setAuthNavigate((path) => goto(path));

  beforeNavigate(() => {
    NProgress.start();
  });

  afterNavigate(() => {
    NProgress.done();
  });

  onMount(() => {
    // Register the SVG sprite (client-only virtual module)
    import("virtual:svg-icons-register");
  });
</script>

{@render children?.()}
