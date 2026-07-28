<script lang="ts">
  import { onMount } from "svelte";
  import { apiUrl } from "@/services/api";
  import { t } from "@/i18n/index.svelte";
  import FormControl from "@/components/ui/FormControl.svelte";
  import Input from "@/components/ui/Input.svelte";
  import Select from "@/components/ui/Select.svelte";
  import Button from "@/components/ui/Button.svelte";
  import { fetchWithAuth } from "@/utils/auth";

  let storageSettings = $state({
    provider: "local",
    s3_bucket: "",
    s3_region: "us-east-1",
    s3_access_key: "",
    s3_secret_key: ""
  });

  onMount(async () => {
    await loadSettings();
  });

  async function loadSettings(): Promise<void> {
    try {
      const response = await fetchWithAuth(apiUrl("/settings"));
      if (response.ok) {
        const data = await response.json();
        const settings = data.data || data;
        if (settings.storage) {
          storageSettings = {
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
      const response = await fetchWithAuth(apiUrl("/settings/storage"), {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(storageSettings)
      });

      if (response.ok) {
        alert(t("settings.storageSaved"));
      } else {
        alert(t("settings.storageSaveError"));
      }
    } catch (error) {
      console.error("Failed to save storage settings:", error);
      alert(t("settings.storageSaveError"));
    }
  }
</script>

<div class="space-y-4">
  <FormControl label={t("settings.storageProvider")}>
    <Select bind:value={storageSettings.provider}>
      <option value="local">{t("settings.localFilesystem")}</option>
      <option value="s3">{t("settings.amazonS3")}</option>
      <option value="gcs">{t("settings.googleCloudStorage")}</option>
      <option value="azure">{t("settings.azureBlobStorage")}</option>
    </Select>
  </FormControl>

  {#if storageSettings.provider === "s3"}
    <FormControl label={t("settings.s3Bucket")}>
      <Input bind:value={storageSettings.s3_bucket} type="text" />
    </FormControl>

    <FormControl label={t("settings.region")}>
      <Input bind:value={storageSettings.s3_region} type="text" placeholder="us-east-1" />
    </FormControl>

    <div class="grid grid-cols-2 gap-4">
      <FormControl label={t("settings.accessKeyId")}>
        <Input bind:value={storageSettings.s3_access_key} type="text" />
      </FormControl>

      <FormControl label={t("settings.secretAccessKey")}>
        <Input bind:value={storageSettings.s3_secret_key} type="password" />
      </FormControl>
    </div>
  {/if}

  <div class="flex justify-end pt-4">
    <Button variant="primary" onclick={saveStorage}>{t("common.save")}</Button>
  </div>
</div>
