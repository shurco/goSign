<template>
  <div class="space-y-4">
    <ResourceTable
      :data="apiKeys"
      :columns="apiKeyColumns"
      :has-actions="true"
      :show-pagination="false"
      :searchable="false"
      :show-filters="false"
      :show-edit="false"
      :empty-message="$t('apikeys.noApiKeys')"
      @delete="deleteAPIKey"
    >
      <template #cell-enabled="{ value }">
        <Badge :variant="value ? 'success' : 'ghost'">
          {{ value ? $t('apikeys.active') : $t('apikeys.disabled') }}
        </Badge>
      </template>

      <template #actions="{ item }">
        <Button
          size="sm"
          :variant="(item as APIKey).enabled ? 'warning' : 'success'"
          @click="toggleAPIKey(item as APIKey)"
        >
          {{ (item as APIKey).enabled ? $t('apikeys.disable') : $t('apikeys.enable') }}
        </Button>
      </template>
    </ResourceTable>

    <!-- API Key Modal -->
    <FormModal v-model="showAPIKeyModal" :title="$t('apikeys.createApiKey')" @submit="saveAPIKey">
      <template #default="{ formData }">
        <div class="space-y-4">
          <FieldInput
            v-model="formData.name as string"
            type="text"
            :label="$t('apikeys.keyName')"
            :placeholder="$t('apikeys.keyNamePlaceholder')"
            required
          />

          <FieldInput v-model="formData.expires_at as string" type="date" :label="$t('apikeys.expiresAt')" />

          <Alert v-if="newAPIKey" variant="info">
            <template #icon>
              <SvgIcon name="info" class="h-6 w-6 shrink-0" />
            </template>
            <div>
              <p class="font-bold">{{ $t('apikeys.saveThisKey') }}</p>
              <p class="text-sm">{{ newAPIKey }}</p>
              <p class="mt-1 text-xs">{{ $t('apikeys.wontBeShownAgain') }}</p>
            </div>
          </Alert>
        </div>
      </template>
    </FormModal>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import ResourceTable from "@/components/common/ResourceTable.vue";
import FormModal from "@/components/common/FormModal.vue";
import FieldInput from "@/components/common/FieldInput.vue";
import Button from "@/components/ui/Button.vue";
import Badge from "@/components/ui/Badge.vue";
import Alert from "@/components/ui/Alert.vue";
import SvgIcon from "@/components/SvgIcon.vue";
import { fetchWithAuth } from "@/utils/auth";

const { t } = useI18n();

interface APIKey {
  id: string;
  name: string;
  enabled: boolean;
  last_used_at?: string;
  expires_at?: string;
}

const apiKeys = ref<APIKey[]>([]);
const showAPIKeyModal = ref(false);
const newAPIKey = ref("");

// Get trigger ref from parent and register open function
const apiKeyModalTrigger = inject<{ value: (() => void) | null }>('apiKeyModalTrigger');

// Register open function with parent
if (apiKeyModalTrigger) {
  apiKeyModalTrigger.value = () => {
    openAPIKeyModal();
  };
}

const apiKeyColumns = computed(() => [
  { key: "name", label: t('apikeys.name'), sortable: true },
  {
    key: "last_used_at",
    label: t('apikeys.lastUsed'),
    formatter: (value: unknown): string => (value ? new Date(String(value)).toLocaleDateString() : t('apikeys.never'))
  },
  {
    key: "expires_at",
    label: t('apikeys.expires'),
    formatter: (value: unknown): string => (value ? new Date(String(value)).toLocaleDateString() : t('apikeys.never'))
  },
  { key: "enabled", label: t('apikeys.status') }
]);

onMounted(async () => {
  await loadAPIKeys();
});

async function loadAPIKeys(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/apikeys");
    if (response.ok) {
      const data = await response.json();
      if (Array.isArray(data.data)) {
        apiKeys.value = data.data;
      } else {
        apiKeys.value = [];
      }
    }
  } catch (error) {
    if (!window.location.pathname.includes("/auth/") && !window.location.pathname.includes("/signin")) {
      console.error("Failed to load API keys:", error);
    }
  }
}

function openAPIKeyModal(): void {
  newAPIKey.value = "";
  showAPIKeyModal.value = true;
}

async function saveAPIKey(formData: any): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/apikeys", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(formData)
    });

    if (response.ok) {
      const data = await response.json();
      newAPIKey.value = data.data.key;
      await loadAPIKeys();
      setTimeout(() => {
        if (confirm(t('apikeys.savedKeyConfirm'))) {
          showAPIKeyModal.value = false;
          newAPIKey.value = "";
        }
      }, 2000);
    } else {
      alert(t('apikeys.createError'));
    }
  } catch (error) {
    console.error("Failed to create API key:", error);
    alert(t('apikeys.createError'));
  }
}

async function deleteAPIKey(apiKey: any): Promise<void> {
  if (!confirm(t('apikeys.deleteConfirm'))) {
    return;
  }

  try {
    const response = await fetchWithAuth(`/api/v1/apikeys/${apiKey.id}`, {
      method: "DELETE"
    });

    if (response.ok) {
      await loadAPIKeys();
    } else {
      alert(t('apikeys.deleteError'));
    }
  } catch (error) {
    console.error("Failed to delete API key:", error);
    alert(t('apikeys.deleteError'));
  }
}

async function toggleAPIKey(apiKey: APIKey): Promise<void> {
  const action = apiKey.enabled ? "disable" : "enable";

  try {
    const response = await fetchWithAuth(`/api/v1/apikeys/${apiKey.id}/${action}`, {
      method: "PUT"
    });

    if (response.ok) {
      await loadAPIKeys();
    } else {
      alert(t(`apikeys.${action}Error`));
    }
  } catch (error) {
    console.error(`Failed to ${action} API key:`, error);
    alert(t(`apikeys.${action}Error`));
  }
}
</script>
