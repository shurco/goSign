<template>
  <div class="space-y-4">
    <FormControl :label="$t('settings.storageProvider')">
      <Select v-model="storageSettings.provider">
        <option value="local">{{ $t('settings.localFilesystem') }}</option>
        <option value="s3">{{ $t('settings.amazonS3') }}</option>
        <option value="gcs">{{ $t('settings.googleCloudStorage') }}</option>
        <option value="azure">{{ $t('settings.azureBlobStorage') }}</option>
      </Select>
    </FormControl>

    <template v-if="storageSettings.provider === 's3'">
      <FormControl :label="$t('settings.s3Bucket')">
        <Input v-model="storageSettings.s3_bucket" type="text" />
      </FormControl>

      <FormControl :label="$t('settings.region')">
        <Input v-model="storageSettings.s3_region" type="text" placeholder="us-east-1" />
      </FormControl>

      <div class="grid grid-cols-2 gap-4">
        <FormControl :label="$t('settings.accessKeyId')">
          <Input v-model="storageSettings.s3_access_key" type="text" />
        </FormControl>

        <FormControl :label="$t('settings.secretAccessKey')">
          <Input v-model="storageSettings.s3_secret_key" type="password" />
        </FormControl>
      </div>
    </template>

    <div class="flex justify-end pt-4">
      <Button variant="primary" @click="saveStorage">{{ $t('common.save') }}</Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import FormControl from "@/components/ui/FormControl.vue";
import Input from "@/components/ui/Input.vue";
import Select from "@/components/ui/Select.vue";
import Button from "@/components/ui/Button.vue";
import { fetchWithAuth } from "@/utils/auth";

const { t } = useI18n();

const storageSettings = ref({
  provider: "local",
  s3_bucket: "",
  s3_region: "us-east-1",
  s3_access_key: "",
  s3_secret_key: ""
});

onMounted(async () => {
  await loadSettings();
});

async function loadSettings(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/settings");
    if (response.ok) {
      const data = await response.json();
      const settings = data.data || data;
      if (settings.storage) {
        storageSettings.value = {
          provider: settings.storage.provider || "local",
          s3_bucket: settings.storage.bucket || "",
          s3_region: settings.storage.region || "us-east-1",
          s3_access_key: "",
          s3_secret_key: ""
        };
      }
    }
  } catch (error) {
    if (!window.location.pathname.includes("/auth/") && !window.location.pathname.includes("/signin")) {
      console.error("Failed to load settings:", error);
    }
  }
}

async function saveStorage(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/settings/storage", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(storageSettings.value)
    });

    if (response.ok) {
      alert(t('settings.storageSaved'));
    } else {
      alert(t('settings.storageSaveError'));
    }
  } catch (error) {
    console.error("Failed to save storage settings:", error);
    alert(t('settings.storageSaveError'));
  }
}
</script>
