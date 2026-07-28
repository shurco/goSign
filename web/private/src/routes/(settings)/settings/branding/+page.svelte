<script lang="ts">
  import { onMount } from "svelte";
  import { apiGet, apiPut } from "@/services/api";
  import { t } from "@/i18n/index.svelte";
  import FormControl from "@/components/ui/FormControl.svelte";
  import Input from "@/components/ui/Input.svelte";
  import FileDropZone from "@/components/ui/FileDropZone.svelte";
  import Button from "@/components/ui/Button.svelte";

  let brandingSettings = $state({
    company_name: "",
    logo_url: "",
    primary_color: "#4F46E5"
  });
  let logoFileName = $state("");

  onMount(async () => {
    await loadBranding();
  });

  async function loadBranding(): Promise<void> {
    try {
      const data = await apiGet("/settings/branding");
      if (data.data && data.data.branding) {
        const branding = data.data.branding;
        brandingSettings = {
          company_name: branding.company_name || branding.CompanyName || "",
          logo_url: branding.logo_url || branding.LogoURL || "",
          primary_color: branding.primary_color || branding.PrimaryColor || "#4F46E5"
        };
      }
    } catch (error) {
      if (!window.location.pathname.includes("/auth/") && !window.location.pathname.includes("/signin")) {
        console.error("Failed to load branding:", error);
      }
    }
  }

  async function saveBranding(): Promise<void> {
    try {
      await apiPut("/settings/branding", { branding: brandingSettings });
      await loadBranding();
      alert(t("success.saved"));
    } catch (error) {
      console.error("Failed to save branding settings:", error);
      alert(t("settings.brandingSaveError"));
    }
  }

  function handleLogoUpload(file: File): void {
    logoFileName = file.name;
    const reader = new FileReader();
    reader.onload = (e) => {
      brandingSettings.logo_url = (e.target?.result as string) || "";
    };
    reader.readAsDataURL(file);
  }

  function clearLogo(): void {
    logoFileName = "";
    brandingSettings.logo_url = "";
  }
</script>

<div class="space-y-4">
  <FormControl label={t("branding.companyName")}>
    <Input bind:value={brandingSettings.company_name} type="text" />
  </FormControl>

  <FormControl label={t("branding.companyLogo")}>
    <FileDropZone
      accept="image/*"
      selectedLabel={logoFileName || (brandingSettings.logo_url ? "Image" : "")}
      onChange={handleLogoUpload}
      onClear={clearLogo}
    />
    {#if brandingSettings.logo_url}
      <img src={brandingSettings.logo_url} alt={t("branding.companyLogo")} class="mt-2 max-h-20 object-contain" />
    {/if}
  </FormControl>

  <FormControl label={t("branding.primaryColor")}>
    <Input bind:value={brandingSettings.primary_color} type="color" class="w-32" />
  </FormControl>

  <div class="flex justify-end pt-4">
    <Button variant="primary" onclick={saveBranding}>{t("common.save")}</Button>
  </div>
</div>
