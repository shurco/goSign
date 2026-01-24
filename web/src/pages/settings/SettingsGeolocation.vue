<template>
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
              <h5 class="font-semibold text-gray-900">{{ $t('settings.downloadFromMaxMind') }}</h5>
            </div>
            <p class="mb-4 text-sm text-gray-600">
              {{ $t('settings.downloadFromMaxMindDescription') }}
            </p>

            <!-- Saved key (more visible) -->
            <div
              v-if="maxmindLicenseKeyMasked"
              class="mb-4 rounded-lg border border-gray-200 bg-gray-50 px-4 py-3"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <div class="text-xs font-medium text-gray-600">{{ $t('settings.savedKeyLabel') }}</div>
                  <div class="mt-1 font-mono text-base font-semibold text-gray-900">
                    {{ maxmindLicenseKeyMasked }}
                  </div>
                </div>
                <button
                  class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
                  @click.stop="deleteMaxMindKey"
                  :disabled="deletingKey || !maxmindLicenseKeySet"
                  :title="$t('settings.deleteKey')"
                  :aria-label="$t('settings.deleteKey')"
                  type="button"
                >
                  <LoadingSpinner v-if="deleteKeyState === 'saving'" size="md" />
                  <SvgIcon v-else name="trash-x" class="h-5 w-5 stroke-[2]" />
                </button>
              </div>
              <div v-if="deleteKeyError" class="mt-2 text-xs text-red-600">
                {{ deleteKeyError }}
              </div>
            </div>

            <div class="space-y-3">
              <FormControl :label="$t('settings.maxmindLicenseKey')">
                <Input
                  v-model="maxmindLicenseKeyForDownload"
                  type="password"
                  :placeholder="
                    maxmindLicenseKeySet
                      ? $t('settings.useSavedKeyOrEnterNew')
                      : $t('settings.maxmindLicenseKeyPlaceholder')
                  "
                  class="w-full"
                />
                <div v-if="maxmindLicenseKeySet" class="mt-1 text-xs text-gray-500">
                  {{ $t('settings.savedKeyWillBeUsedIfEmpty') }}
                </div>
                <div v-if="saveMaxMindError" class="mt-1 text-xs text-red-600">
                  {{ saveMaxMindError }}
                </div>
              </FormControl>

              <div class="flex">
                <Button
                  variant="ghost"
                  @click="saveMaxMindMethod"
                  :disabled="savingMaxMind || (!maxmindLicenseKeySet && !maxmindLicenseKeyForDownload)"
                  class="w-full"
                >
                  <span v-if="saveMaxMindState === 'saving'">{{ $t('common.saving') }}...</span>
                  <span v-else-if="saveMaxMindState === 'saved'">{{ $t('common.saved') }}</span>
                  <span v-else-if="saveMaxMindState === 'error'">{{ $t('common.failed') }}</span>
                  <span v-else>{{ $t('common.save') }}</span>
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
              <h5 class="font-semibold text-gray-900">{{ $t('settings.downloadFromUrl') }}</h5>
            </div>
            <p class="mb-4 text-sm text-gray-600">
              {{ $t('settings.downloadUrlDescription') }}
            </p>
            <div class="space-y-3">
              <FormControl :label="$t('settings.downloadUrl')">
                <Input
                  v-model="downloadUrl"
                  type="url"
                  :placeholder="$t('settings.downloadUrlPlaceholder')"
                  class="w-full"
                />
                <div v-if="saveUrlError" class="mt-1 text-xs text-red-600">
                  {{ saveUrlError }}
                </div>
              </FormControl>
              <div class="flex gap-2">
                <Button variant="ghost" @click="saveUrlMethod" :disabled="savingUrl || !downloadUrl" class="w-full">
                  <span v-if="saveUrlState === 'saving'">{{ $t('common.saving') }}...</span>
                  <span v-else-if="saveUrlState === 'saved'">{{ $t('common.saved') }}</span>
                  <span v-else-if="saveUrlState === 'error'">{{ $t('common.failed') }}</span>
                  <span v-else>{{ $t('common.save') }}</span>
                </Button>
              </div>
            </div>
          </div>
        </div>

      <!-- Manual Update Button -->
      <div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
        <div class="flex items-center justify-between">
          <div>
            <h5 class="font-semibold text-gray-900">{{ $t('settings.manualUpdate') }}</h5>
            <p class="mt-1 text-sm text-gray-600">
              {{ $t('settings.manualUpdateDescription') }}
            </p>
            <p class="mt-1 text-sm text-gray-600">
              <span class="font-medium text-gray-700">{{ $t('settings.lastUpdated') }}:</span>
              {{ lastUpdatedLabel }}
            </p>
            <p class="mt-1 text-sm text-gray-600">
              <span class="font-medium text-gray-700">{{ $t('settings.downloadSource') }}:</span>
              {{ lastUpdatedSourceLabel }}
            </p>
          </div>
          <Button variant="primary" @click="manualUpdate" :disabled="manualUpdating || !hasSavedSettings" class="ml-4">
            <span v-if="manualUpdateState === 'saving'">{{ $t('settings.downloading') }}...</span>
            <span v-else-if="manualUpdateState === 'saved'">{{ $t('common.updated') }}</span>
            <span v-else-if="manualUpdateState === 'error'">{{ $t('common.failed') }}</span>
            <span v-else>{{ $t('settings.updateNow') }}</span>
          </Button>
        </div>
        <div v-if="manualUpdateError" class="mt-2 text-sm text-red-600">
          {{ manualUpdateError }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { apiDelete, apiGet, apiPost, apiPut } from "@/services/api";
import FormControl from "@/components/ui/FormControl.vue";
import Input from "@/components/ui/Input.vue";
import Button from "@/components/ui/Button.vue";
import LoadingSpinner from "@/components/ui/LoadingSpinner.vue";
import SvgIcon from "@/components/SvgIcon.vue";

const { t } = useI18n();

type ActionState = "idle" | "saving" | "saved" | "error";

const saveUrlState = ref<ActionState>("idle");
const saveMaxMindState = ref<ActionState>("idle");
const manualUpdateState = ref<ActionState>("idle");

const savingUrl = computed(() => saveUrlState.value === "saving");
const savingMaxMind = computed(() => saveMaxMindState.value === "saving");
const manualUpdating = computed(() => manualUpdateState.value === "saving");

const saveUrlError = ref("");
const saveMaxMindError = ref("");
const manualUpdateError = ref("");
const deleteKeyError = ref("");

const downloadUrl = ref("");
const maxmindLicenseKeyForDownload = ref("");

const maxmindLicenseKeySet = ref(false);
const maxmindLicenseKeyMasked = ref("");
const savedDownloadUrl = ref("");
const lastUpdatedAt = ref<string>("");
const lastUpdatedSource = ref<string>("");

async function loadSettings() {
  try {
    const response = await apiGet("/api/v1/settings");
    if (response.data?.geolocation) {
      maxmindLicenseKeySet.value = response.data.geolocation.maxmind_license_key_set === true;
      maxmindLicenseKeyMasked.value = response.data.geolocation.maxmind_license_key_masked || "";
      savedDownloadUrl.value = response.data.geolocation.download_url || "";
      lastUpdatedAt.value = response.data.geolocation.last_updated_at || "";
      lastUpdatedSource.value = response.data.geolocation.last_updated_source || "";

      // Pre-fill URL if saved
      if (savedDownloadUrl.value && !downloadUrl.value) {
        downloadUrl.value = savedDownloadUrl.value;
      }
    }
  } catch (error) {
    console.error("Failed to load geolocation settings:", error);
  }
}

const lastUpdatedLabel = computed(() => {
  if (!lastUpdatedAt.value) return t("settings.notUpdatedYet");
  const d = new Date(lastUpdatedAt.value);
  if (Number.isNaN(d.getTime())) return lastUpdatedAt.value;
  return d.toLocaleString();
});

const lastUpdatedSourceLabel = computed(() => {
  if (!lastUpdatedSource.value) return t("settings.unknownSource");
  if (lastUpdatedSource.value === "maxmind") return t("settings.sourceMaxMind");
  if (lastUpdatedSource.value === "url") return t("settings.sourceUrl");
  return t("settings.unknownSource");
});

const deleteKeyState = ref<ActionState>("idle");
const deletingKey = computed(() => deleteKeyState.value === "saving");

async function deleteMaxMindKey() {
  deleteKeyError.value = "";
  if (!maxmindLicenseKeySet.value) return;
  if (deletingKey.value) return;

  deleteKeyState.value = "saving";
  try {
    await apiDelete("/api/v1/settings/geolocation/maxmind-key");
    // Refresh state from backend
    await loadSettings();
    deleteKeyState.value = "saved";
    window.setTimeout(() => (deleteKeyState.value = "idle"), 1500);
  } catch (error) {
    console.error("Failed to delete MaxMind key:", error);
    const msg =
      error && typeof error === "object" && "message" in error && typeof (error as any).message === "string"
        ? String((error as any).message)
        : t("settings.failedToSaveSettings");
    deleteKeyError.value = msg;
    deleteKeyState.value = "error";
    window.setTimeout(() => (deleteKeyState.value = "idle"), 1500);
  }
}

async function forceDownloadFromUrl(urlOverride: string) {
  const urlToUse = urlOverride || downloadUrl.value || savedDownloadUrl.value;
  if (!urlToUse) {
    manualUpdateError.value = t("settings.pleaseEnterDownloadUrl");
    return false;
  }

  try {
    const response = await apiPost("/api/v1/settings/geolocation/download", { url: urlToUse, force: true });

    if (response.data?.status === "success" || response.data?.status === "skipped") {
      // No popups; just return success (button will show state)
      return true;
    } else {
      throw new Error(response.message || "Failed to download database");
    }
  } catch (error) {
    console.error("Failed to download database:", error);
    manualUpdateError.value = error instanceof Error ? error.message : t("settings.failedToDownloadDatabase");
    return false;
  } finally {
  }
}

async function forceDownloadFromMaxMind() {
  if (!maxmindLicenseKeySet.value) {
    manualUpdateError.value = t("settings.pleaseEnterMaxMindLicenseKey");
    return false;
  }

  try {
    const response = await apiPost("/api/v1/settings/geolocation/download-maxmind", {
      force: true
    });

    if (response.data?.status === "success" || response.data?.status === "skipped") {
      return true;
    } else {
      throw new Error(response.message || "Failed to download database from MaxMind");
    }
  } catch (error) {
    console.error("Failed to download database from MaxMind:", error);
    manualUpdateError.value = error instanceof Error ? error.message : t("settings.failedToDownloadDatabase");
    return false;
  } finally {
  }
}

async function saveUrlMethod() {
  saveUrlError.value = "";
  saveUrlState.value = "idle";

  const url = downloadUrl.value.trim();
  if (!url) {
    saveUrlError.value = t("settings.pleaseEnterDownloadUrl");
    saveUrlState.value = "error";
    return;
  }

  saveUrlState.value = "saving";
  try {
    await apiPut("/api/v1/settings/geolocation", { download_url: url, download_method: "url" });
    savedDownloadUrl.value = url;
    saveUrlState.value = "saved";
    await loadSettings();
  } catch (error) {
    console.error("Failed to save URL settings:", error);
    const msg =
      error && typeof error === "object" && "message" in error && typeof (error as any).message === "string"
        ? String((error as any).message)
        : t("settings.failedToSaveSettings");
    saveUrlError.value = msg;
    saveUrlState.value = "error";
  } finally {
    if (saveUrlState.value === "saved") {
      window.setTimeout(() => {
        saveUrlState.value = "idle";
      }, 1500);
    }
  }
}

async function saveMaxMindMethod() {
  saveMaxMindError.value = "";
  saveMaxMindState.value = "idle";

  const licenseKey = maxmindLicenseKeyForDownload.value.trim();

  if (!maxmindLicenseKeySet.value && !licenseKey) {
    saveMaxMindError.value = t("settings.pleaseEnterMaxMindLicenseKey");
    saveMaxMindState.value = "error";
    return;
  }

  saveMaxMindState.value = "saving";
  try {
    const payload: any = { download_method: "maxmind" };
    if (licenseKey) {
      payload.maxmind_license_key = licenseKey;
    }

    await apiPut("/api/v1/settings/geolocation", payload);

    if (licenseKey) {
      maxmindLicenseKeySet.value = true;
      maxmindLicenseKeyForDownload.value = "";
    }

    saveMaxMindState.value = "saved";
    await loadSettings();
  } catch (error) {
    console.error("Failed to save MaxMind settings:", error);
    const msg =
      error && typeof error === "object" && "message" in error && typeof (error as any).message === "string"
        ? String((error as any).message)
        : t("settings.failedToSaveSettings");
    saveMaxMindError.value = msg;
    saveMaxMindState.value = "error";
  } finally {
    if (saveMaxMindState.value === "saved") {
      window.setTimeout(() => {
        saveMaxMindState.value = "idle";
      }, 1500);
    }
  }
}

async function manualUpdate() {
  manualUpdateError.value = "";
  if (manualUpdating.value) return;
  manualUpdateState.value = "saving";

  if (maxmindLicenseKeySet.value) {
    const ok = await forceDownloadFromMaxMind();
    if (ok) {
      await loadSettings();
    }
    manualUpdateState.value = ok ? "saved" : "error";
    window.setTimeout(() => (manualUpdateState.value = "idle"), 1500);
    return;
  }

  if (savedDownloadUrl.value) {
    const ok = await forceDownloadFromUrl(savedDownloadUrl.value);
    if (ok) {
      await loadSettings();
    }
    manualUpdateState.value = ok ? "saved" : "error";
    window.setTimeout(() => (manualUpdateState.value = "idle"), 1500);
    return;
  }

  manualUpdateError.value = t("settings.noMethodConfigured");
  manualUpdateState.value = "error";
  window.setTimeout(() => (manualUpdateState.value = "idle"), 1500);
}

const hasSavedSettings = computed(() => {
  return maxmindLicenseKeySet.value || savedDownloadUrl.value !== "";
});

onMounted(() => {
  loadSettings();
});
</script>
