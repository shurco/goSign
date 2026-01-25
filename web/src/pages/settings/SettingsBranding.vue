<template>
  <div class="space-y-4">
    <FormControl :label="$t('branding.companyName')">
      <Input v-model="brandingSettings.company_name" type="text" />
    </FormControl>

    <FormControl :label="$t('branding.companyLogo')">
      <FileDropZone
        accept="image/*"
        :selected-label="logoFileName || (brandingSettings.logo_url ? 'Image' : '')"
        @change="handleLogoUpload"
        @clear="clearLogo"
      />
      <img
        v-if="brandingSettings.logo_url"
        :src="brandingSettings.logo_url"
        :alt="$t('branding.companyLogo')"
        class="mt-2 max-h-20 object-contain"
      />
    </FormControl>

    <FormControl :label="$t('branding.primaryColor')">
      <Input v-model="brandingSettings.primary_color" type="color" class="w-32" />
    </FormControl>

    <div class="flex justify-end pt-4">
      <Button variant="primary" @click="saveBranding">{{ $t('common.save') }}</Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import FormControl from "@/components/ui/FormControl.vue";
import Input from "@/components/ui/Input.vue";
import FileDropZone from "@/components/ui/FileDropZone.vue";
import Button from "@/components/ui/Button.vue";
import { fetchWithAuth } from "@/utils/auth";

const { t } = useI18n();

const brandingSettings = ref({
  company_name: "",
  logo_url: "",
  primary_color: "#4F46E5"
});
const logoFileName = ref("");

onMounted(async () => {
  await loadBranding();
});

async function loadBranding(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/branding");
    if (response.ok) {
      const data = await response.json();
      if (data.data && data.data.branding) {
        const branding = data.data.branding;
        brandingSettings.value = {
          company_name: branding.company_name || branding.CompanyName || "",
          logo_url: branding.logo_url || branding.LogoURL || "",
          primary_color: branding.primary_color || branding.PrimaryColor || "#4F46E5"
        };
      }
    }
  } catch (error) {
    if (!window.location.pathname.includes("/auth/") && !window.location.pathname.includes("/signin")) {
      console.error("Failed to load branding:", error);
    }
  }
}

async function saveBranding(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/branding", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        branding: brandingSettings.value
      })
    });

    if (response.ok) {
      await loadBranding();
      alert(t('success.saved'));
    } else {
      alert(t('settings.brandingSaveError'));
    }
  } catch (error) {
    console.error("Failed to save branding settings:", error);
    alert(t('settings.brandingSaveError'));
  }
}

function handleLogoUpload(file: File): void {
  logoFileName.value = file.name;
  const reader = new FileReader();
  reader.onload = (e) => {
    brandingSettings.value.logo_url = (e.target?.result as string) || "";
  };
  reader.readAsDataURL(file);
}

function clearLogo(): void {
  logoFileName.value = "";
  brandingSettings.value.logo_url = "";
}
</script>
