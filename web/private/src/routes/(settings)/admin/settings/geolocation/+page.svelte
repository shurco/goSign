<script lang="ts">
  import { onMount } from "svelte";
  import { t } from "@/i18n/index.svelte";
  import { apiDelete, apiGet, apiPost, apiPut } from "@/services/api";
  import FormControl from "@/components/ui/FormControl.svelte";
  import Input from "@/components/ui/Input.svelte";
  import Button from "@/components/ui/Button.svelte";
  import LoadingSpinner from "@/components/ui/LoadingSpinner.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

  type ActionState = "idle" | "saving" | "saved" | "error";

  let saveUrlState = $state<ActionState>("idle");
  let saveMaxMindState = $state<ActionState>("idle");
  let manualUpdateState = $state<ActionState>("idle");

  const savingUrl = $derived(saveUrlState === "saving");
  const savingMaxMind = $derived(saveMaxMindState === "saving");
  const manualUpdating = $derived(manualUpdateState === "saving");

  let saveUrlError = $state("");
  let saveMaxMindError = $state("");
  let manualUpdateError = $state("");
  let deleteKeyError = $state("");

  let downloadUrl = $state("");
  let maxmindLicenseKeyForDownload = $state("");

  let maxmindLicenseKeySet = $state(false);
  let maxmindLicenseKeyMasked = $state("");
  let savedDownloadUrl = $state("");
  let lastUpdatedAt = $state<string>("");
  let lastUpdatedSource = $state<string>("");

  async function loadSettings() {
    try {
      const response = await apiGet("/settings");
      if (response.data?.geolocation) {
        maxmindLicenseKeySet = response.data.geolocation.maxmind_license_key_set === true;
        maxmindLicenseKeyMasked = response.data.geolocation.maxmind_license_key_masked || "";
        savedDownloadUrl = response.data.geolocation.download_url || "";
        lastUpdatedAt = response.data.geolocation.last_updated_at || "";
        lastUpdatedSource = response.data.geolocation.last_updated_source || "";

        // Pre-fill URL if saved
        if (savedDownloadUrl && !downloadUrl) {
          downloadUrl = savedDownloadUrl;
        }
      }
    } catch (error) {
      console.error("Failed to load geolocation settings:", error);
    }
  }

  const lastUpdatedLabel = $derived.by(() => {
    if (!lastUpdatedAt) {
      return t("settings.notUpdatedYet");
    }
    const d = new Date(lastUpdatedAt);
    if (Number.isNaN(d.getTime())) {
      return lastUpdatedAt;
    }
    return d.toLocaleString();
  });

  const lastUpdatedSourceLabel = $derived.by(() => {
    if (!lastUpdatedSource) {
      return t("settings.unknownSource");
    }
    if (lastUpdatedSource === "maxmind") {
      return t("settings.sourceMaxMind");
    }
    if (lastUpdatedSource === "url") {
      return t("settings.sourceUrl");
    }
    return t("settings.unknownSource");
  });

  let deleteKeyState = $state<ActionState>("idle");
  const deletingKey = $derived(deleteKeyState === "saving");

  async function deleteMaxMindKey() {
    deleteKeyError = "";
    if (!maxmindLicenseKeySet) {
      return;
    }
    if (deletingKey) {
      return;
    }

    deleteKeyState = "saving";
    try {
      await apiDelete("/settings/geolocation/maxmind-key");
      // Refresh state from backend
      await loadSettings();
      deleteKeyState = "saved";
      window.setTimeout(() => (deleteKeyState = "idle"), 1500);
    } catch (error) {
      console.error("Failed to delete MaxMind key:", error);
      const msg =
        error &&
        typeof error === "object" &&
        "message" in error &&
        typeof (error as { message: string }).message === "string"
          ? String((error as { message: string }).message)
          : t("settings.failedToSaveSettings");
      deleteKeyError = msg;
      deleteKeyState = "error";
      window.setTimeout(() => (deleteKeyState = "idle"), 1500);
    }
  }

  async function forceDownloadFromUrl(urlOverride: string) {
    const urlToUse = urlOverride || downloadUrl || savedDownloadUrl;
    if (!urlToUse) {
      manualUpdateError = t("settings.pleaseEnterDownloadUrl");
      return false;
    }

    try {
      const response = await apiPost("/settings/geolocation/download", { method: "url", url: urlToUse, force: true });

      if (response.data?.status === "success" || response.data?.status === "skipped") {
        // No popups; just return success (button will show state)
        return true;
      } else {
        throw new Error(response.message || "Failed to download database");
      }
    } catch (error) {
      console.error("Failed to download database:", error);
      manualUpdateError = error instanceof Error ? error.message : t("settings.failedToDownloadDatabase");
      return false;
    }
  }

  async function forceDownloadFromMaxMind() {
    if (!maxmindLicenseKeySet) {
      manualUpdateError = t("settings.pleaseEnterMaxMindLicenseKey");
      return false;
    }

    try {
      const response = await apiPost("/settings/geolocation/download", {
        method: "maxmind",
        force: true
      });

      if (response.data?.status === "success" || response.data?.status === "skipped") {
        return true;
      } else {
        throw new Error(response.message || "Failed to download database from MaxMind");
      }
    } catch (error) {
      console.error("Failed to download database from MaxMind:", error);
      manualUpdateError = error instanceof Error ? error.message : t("settings.failedToDownloadDatabase");
      return false;
    }
  }

  async function saveUrlMethod() {
    saveUrlError = "";
    saveUrlState = "idle";

    const url = downloadUrl.trim();
    if (!url) {
      saveUrlError = t("settings.pleaseEnterDownloadUrl");
      saveUrlState = "error";
      return;
    }

    saveUrlState = "saving";
    try {
      await apiPut("/settings/geolocation", { download_url: url, download_method: "url" });
      savedDownloadUrl = url;
      saveUrlState = "saved";
      await loadSettings();
    } catch (error) {
      console.error("Failed to save URL settings:", error);
      const msg =
        error &&
        typeof error === "object" &&
        "message" in error &&
        typeof (error as { message: string }).message === "string"
          ? String((error as { message: string }).message)
          : t("settings.failedToSaveSettings");
      saveUrlError = msg;
      saveUrlState = "error";
    } finally {
      if (saveUrlState === "saved") {
        window.setTimeout(() => {
          saveUrlState = "idle";
        }, 1500);
      }
    }
  }

  async function saveMaxMindMethod() {
    saveMaxMindError = "";
    saveMaxMindState = "idle";

    const licenseKey = maxmindLicenseKeyForDownload.trim();

    if (!maxmindLicenseKeySet && !licenseKey) {
      saveMaxMindError = t("settings.pleaseEnterMaxMindLicenseKey");
      saveMaxMindState = "error";
      return;
    }

    saveMaxMindState = "saving";
    try {
      const payload: Record<string, unknown> = { download_method: "maxmind" };
      if (licenseKey) {
        payload.maxmind_license_key = licenseKey;
      }

      await apiPut("/settings/geolocation", payload);

      if (licenseKey) {
        maxmindLicenseKeySet = true;
        maxmindLicenseKeyForDownload = "";
      }

      saveMaxMindState = "saved";
      await loadSettings();
    } catch (error) {
      console.error("Failed to save MaxMind settings:", error);
      const msg =
        error &&
        typeof error === "object" &&
        "message" in error &&
        typeof (error as { message: string }).message === "string"
          ? String((error as { message: string }).message)
          : t("settings.failedToSaveSettings");
      saveMaxMindError = msg;
      saveMaxMindState = "error";
    } finally {
      if (saveMaxMindState === "saved") {
        window.setTimeout(() => {
          saveMaxMindState = "idle";
        }, 1500);
      }
    }
  }

  async function manualUpdate() {
    manualUpdateError = "";
    if (manualUpdating) {
      return;
    }
    manualUpdateState = "saving";

    if (maxmindLicenseKeySet) {
      const ok = await forceDownloadFromMaxMind();
      if (ok) {
        await loadSettings();
      }
      manualUpdateState = ok ? "saved" : "error";
      window.setTimeout(() => (manualUpdateState = "idle"), 1500);
      return;
    }

    if (savedDownloadUrl) {
      const ok = await forceDownloadFromUrl(savedDownloadUrl);
      if (ok) {
        await loadSettings();
      }
      manualUpdateState = ok ? "saved" : "error";
      window.setTimeout(() => (manualUpdateState = "idle"), 1500);
      return;
    }

    manualUpdateError = t("settings.noMethodConfigured");
    manualUpdateState = "error";
    window.setTimeout(() => (manualUpdateState = "idle"), 1500);
  }

  const hasSavedSettings = $derived(maxmindLicenseKeySet || savedDownloadUrl !== "");

  onMount(() => {
    loadSettings();
  });
</script>

<div class="space-y-6">
  <div class="space-y-6">
    <!-- Download options -->
    <div class="grid gap-4 md:grid-cols-2">
      <!-- Option 1: Download from MaxMind (First Priority) -->
      <div class="rounded-lg border-2 border-gray-200 bg-white p-5 transition-all hover:border-gray-300">
        <div class="mb-3 flex items-center gap-2">
          <svg class="h-5 w-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
            />
          </svg>
          <h5 class="font-semibold text-gray-900">{t("settings.downloadFromMaxMind")}</h5>
        </div>
        <p class="mb-4 text-sm text-gray-600">
          {t("settings.downloadFromMaxMindDescription")}
        </p>

        <!-- Saved key (more visible) -->
        {#if maxmindLicenseKeyMasked}
          <div class="mb-4 rounded-lg border border-gray-200 bg-gray-50 px-4 py-3">
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="text-xs font-medium text-gray-600">{t("settings.savedKeyLabel")}</div>
                <div class="mt-1 font-mono text-base font-semibold text-gray-900">
                  {maxmindLicenseKeyMasked}
                </div>
              </div>
              <button
                class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
                disabled={deletingKey || !maxmindLicenseKeySet}
                title={t("settings.deleteKey")}
                aria-label={t("settings.deleteKey")}
                type="button"
                onclick={(e) => {
                  e.stopPropagation();
                  deleteMaxMindKey();
                }}
              >
                {#if deleteKeyState === "saving"}
                  <LoadingSpinner size="md" />
                {:else}
                  <SvgIcon name="trash-x" class="h-5 w-5 stroke-[2]" />
                {/if}
              </button>
            </div>
            {#if deleteKeyError}
              <div class="mt-2 text-xs text-red-600">
                {deleteKeyError}
              </div>
            {/if}
          </div>
        {/if}

        <div class="space-y-3">
          <FormControl label={t("settings.maxmindLicenseKey")}>
            <Input
              bind:value={maxmindLicenseKeyForDownload}
              type="password"
              placeholder={maxmindLicenseKeySet
                ? t("settings.useSavedKeyOrEnterNew")
                : t("settings.maxmindLicenseKeyPlaceholder")}
              class="w-full"
            />
            {#if maxmindLicenseKeySet}
              <div class="mt-1 text-xs text-gray-500">
                {t("settings.savedKeyWillBeUsedIfEmpty")}
              </div>
            {/if}
            {#if saveMaxMindError}
              <div class="mt-1 text-xs text-red-600">
                {saveMaxMindError}
              </div>
            {/if}
          </FormControl>

          <div class="flex">
            <Button
              variant="ghost"
              disabled={savingMaxMind || (!maxmindLicenseKeySet && !maxmindLicenseKeyForDownload)}
              class="w-full"
              onclick={saveMaxMindMethod}
            >
              {#if saveMaxMindState === "saving"}
                {t("common.saving")}...
              {:else if saveMaxMindState === "saved"}
                {t("common.saved")}
              {:else if saveMaxMindState === "error"}
                {t("common.failed")}
              {:else}
                {t("common.save")}
              {/if}
            </Button>
          </div>
        </div>
      </div>

      <!-- Option 2: Download from URL (Fallback) -->
      <div class="rounded-lg border-2 border-gray-200 bg-white p-5 transition-all hover:border-gray-300">
        <div class="mb-3 flex items-center gap-2">
          <svg class="h-5 w-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
            />
          </svg>
          <h5 class="font-semibold text-gray-900">{t("settings.downloadFromUrl")}</h5>
        </div>
        <p class="mb-4 text-sm text-gray-600">
          {t("settings.downloadUrlDescription")}
        </p>
        <div class="space-y-3">
          <FormControl label={t("settings.downloadUrl")}>
            <Input
              bind:value={downloadUrl}
              type="url"
              placeholder={t("settings.downloadUrlPlaceholder")}
              class="w-full"
            />
            {#if saveUrlError}
              <div class="mt-1 text-xs text-red-600">
                {saveUrlError}
              </div>
            {/if}
          </FormControl>
          <div class="flex gap-2">
            <Button variant="ghost" disabled={savingUrl || !downloadUrl} class="w-full" onclick={saveUrlMethod}>
              {#if saveUrlState === "saving"}
                {t("common.saving")}...
              {:else if saveUrlState === "saved"}
                {t("common.saved")}
              {:else if saveUrlState === "error"}
                {t("common.failed")}
              {:else}
                {t("common.save")}
              {/if}
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- Manual Update Button -->
    <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
      <div class="flex items-center justify-between">
        <div>
          <h5 class="font-semibold text-gray-900">{t("settings.manualUpdate")}</h5>
          <p class="mt-1 text-sm text-gray-600">
            {t("settings.manualUpdateDescription")}
          </p>
          <p class="mt-1 text-sm text-gray-600">
            <span class="font-medium text-gray-700">{t("settings.lastUpdated")}:</span>
            {lastUpdatedLabel}
          </p>
          <p class="mt-1 text-sm text-gray-600">
            <span class="font-medium text-gray-700">{t("settings.downloadSource")}:</span>
            {lastUpdatedSourceLabel}
          </p>
        </div>
        <Button variant="primary" disabled={manualUpdating || !hasSavedSettings} class="ml-4" onclick={manualUpdate}>
          {#if manualUpdateState === "saving"}
            {t("settings.downloading")}...
          {:else if manualUpdateState === "saved"}
            {t("common.updated")}
          {:else if manualUpdateState === "error"}
            {t("common.failed")}
          {:else}
            {t("settings.updateNow")}
          {/if}
        </Button>
      </div>
      {#if manualUpdateError}
        <div class="mt-2 text-sm text-red-600">
          {manualUpdateError}
        </div>
      {/if}
    </div>
  </div>
</div>
