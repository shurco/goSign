<script lang="ts">
  import type { Snippet } from "svelte";
  import type { BrandingSettings } from "@/models/account";

  interface Props {
    branding: BrandingSettings;
    children?: Snippet;
  }

  let { branding, children }: Props = $props();

  const theme = $derived(branding.signing_page_theme || "default");
</script>

<div class="signing-page theme-{theme}">
  <!-- Logo -->
  {#if branding.logo_url}
    <div class="company-logo mb-6 text-center">
      <img src={branding.logo_url} alt={branding.company_name} class="max-h-16" />
    </div>
  {/if}

  <!-- Title with company name -->
  <h1 class="mb-4 text-center text-2xl font-bold">
    {branding.company_name || "Document Signing"}
  </h1>

  {@render children?.()}

  <!-- Footer -->
  <footer class="mt-8 border-t pt-4 text-center text-sm text-gray-500">
    {#if branding.terms_url || branding.privacy_url}
      <div class="legal-links mb-2">
        {#if branding.terms_url}
          <a href={branding.terms_url} target="_blank" class="mx-2 hover:underline"> Terms of Service </a>
        {/if}
        {#if branding.privacy_url}
          <a href={branding.privacy_url} target="_blank" class="mx-2 hover:underline"> Privacy Policy </a>
        {/if}
      </div>
    {/if}

    {#if branding.show_powered_by}
      <div class="powered-by">Powered by goSign</div>
    {/if}
  </footer>
</div>

<style>
  /* Default theme */
  .theme-default {
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem;
  }

  /* Minimal theme */
  .theme-minimal {
    max-width: 600px;
    margin: 0 auto;
    padding: 1rem;
    background: white;
    box-shadow: none;
  }

  /* Corporate theme */
  .theme-corporate {
    max-width: 1000px;
    margin: 0 auto;
    padding: 3rem;
    background: var(--color-background, #ffffff);
    border-top: 4px solid var(--color-primary, #4f46e5);
  }
</style>
