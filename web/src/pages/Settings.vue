<template>
  <div class="settings-page">
    <h1 class="mb-6 text-3xl font-bold">Settings</h1>

    <Tabs v-model="activeTab" class="mb-6">
      <Tab v-for="tab in tabs" :key="tab.id" :value="tab.id">
        {{ tab.label }}
      </Tab>
    </Tabs>

    <!-- SMTP Settings -->
    <Card v-if="activeTab === 'smtp'">
      <template #header>
        <div>
          <h2 class="text-lg font-semibold">SMTP Configuration</h2>
          <p class="text-sm text-gray-500">Configure email server settings for sending notifications</p>
        </div>
      </template>

      <div class="space-y-4">
        <FormControl label="SMTP Host">
          <Input v-model="smtpSettings.host" type="text" placeholder="smtp.gmail.com" />
        </FormControl>

        <div class="grid grid-cols-2 gap-4">
          <FormControl label="Port">
            <Input v-model="smtpSettings.port" type="number" placeholder="587" />
          </FormControl>

          <FormControl label="Encryption">
            <Select v-model="smtpSettings.encryption">
              <option value="tls">TLS</option>
              <option value="ssl">SSL</option>
              <option value="none">None</option>
            </Select>
          </FormControl>
        </div>

        <FormControl label="Username">
          <Input v-model="smtpSettings.username" type="text" />
        </FormControl>

        <FormControl label="Password">
          <Input v-model="smtpSettings.password" type="password" />
        </FormControl>

        <FormControl label="From Email">
          <Input v-model="smtpSettings.from_email" type="email" placeholder="noreply@example.com" />
        </FormControl>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <Button variant="ghost" @click="testSmtp">Test Connection</Button>
          <Button variant="primary" @click="saveSmtp">Save</Button>
        </div>
      </template>
    </Card>

    <!-- Storage Settings -->
    <Card v-if="activeTab === 'storage'">
      <template #header>
        <div>
          <h2 class="text-lg font-semibold">Storage Configuration</h2>
          <p class="text-sm text-gray-500">Configure where documents and attachments are stored</p>
        </div>
      </template>

      <div class="space-y-4">
        <FormControl label="Storage Provider">
          <Select v-model="storageSettings.provider">
            <option value="local">Local Filesystem</option>
            <option value="s3">Amazon S3</option>
            <option value="gcs">Google Cloud Storage</option>
            <option value="azure">Azure Blob Storage</option>
          </Select>
        </FormControl>

        <template v-if="storageSettings.provider === 's3'">
          <FormControl label="S3 Bucket">
            <Input v-model="storageSettings.s3_bucket" type="text" />
          </FormControl>

          <FormControl label="Region">
            <Input v-model="storageSettings.s3_region" type="text" placeholder="us-east-1" />
          </FormControl>

          <div class="grid grid-cols-2 gap-4">
            <FormControl label="Access Key ID">
              <Input v-model="storageSettings.s3_access_key" type="text" />
            </FormControl>

            <FormControl label="Secret Access Key">
              <Input v-model="storageSettings.s3_secret_key" type="password" />
            </FormControl>
          </div>
        </template>

        <template v-else-if="storageSettings.provider === 'local'">
          <FormControl label="Storage Path">
            <Input v-model="storageSettings.local_path" type="text" placeholder="/var/lib/gosign/storage" />
          </FormControl>
        </template>
      </div>
      <template #footer>
        <div class="flex justify-end">
          <Button variant="primary" @click="saveStorage">Save</Button>
        </div>
      </template>
    </Card>

    <!-- Webhooks -->
    <Card v-if="activeTab === 'webhooks'">
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-semibold">Webhooks</h2>
          <Button variant="primary" size="sm" @click="openWebhookModal">+ Add Webhook</Button>
        </div>
      </template>

      <ResourceTable
        :data="webhooks"
        :columns="webhookColumns"
        :has-actions="true"
        :show-pagination="false"
        :searchable="false"
        :show-filters="false"
        empty-message="No webhooks configured"
        @edit="editWebhook"
        @delete="deleteWebhook"
      >
        <template #cell-enabled="{ value }">
          <Badge :variant="value ? 'success' : 'ghost'">
            {{ value ? "Active" : "Inactive" }}
          </Badge>
        </template>

        <template #cell-events="{ value }">
          <div class="flex flex-wrap gap-1">
            <Badge v-for="event in value.slice(0, 3)" :key="event" size="sm">
              {{ event }}
            </Badge>
            <Badge v-if="value.length > 3" variant="ghost" size="sm">+{{ value.length - 3 }} more</Badge>
          </div>
        </template>
      </ResourceTable>
    </Card>

    <!-- API Keys -->
    <Card v-if="activeTab === 'api_keys'">
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-semibold">API Keys</h2>
          <Button variant="primary" size="sm" @click="openAPIKeyModal">+ Create API Key</Button>
        </div>
      </template>

      <ResourceTable
        :data="apiKeys"
        :columns="apiKeyColumns"
        :has-actions="true"
        :show-pagination="false"
        :searchable="false"
        :show-filters="false"
        :show-edit="false"
        empty-message="No API keys created"
        @delete="deleteAPIKey"
      >
        <template #cell-enabled="{ value }">
          <Badge :variant="value ? 'success' : 'ghost'">
            {{ value ? "Active" : "Disabled" }}
          </Badge>
        </template>

        <template #actions="{ item }">
          <Button
            size="sm"
            :variant="(item as APIKey).enabled ? 'warning' : 'success'"
            @click="toggleAPIKey(item as APIKey)"
          >
            {{ (item as APIKey).enabled ? "Disable" : "Enable" }}
          </Button>
        </template>
      </ResourceTable>
    </Card>

    <!-- Branding -->
    <Card v-if="activeTab === 'branding'">
      <template #header>
        <div>
          <h2 class="text-lg font-semibold">Branding</h2>
          <p class="text-sm text-gray-500">Customize the appearance of documents and emails</p>
        </div>
      </template>

      <div class="space-y-4">
        <FormControl label="Company Name">
          <Input v-model="brandingSettings.company_name" type="text" />
        </FormControl>

        <FormControl label="Company Logo">
          <FileInput accept="image/*" @change="handleLogoUpload" />
          <img
            v-if="brandingSettings.logo_url"
            :src="brandingSettings.logo_url"
            alt="Company Logo"
            class="mt-2 max-h-20"
          />
        </FormControl>

        <FormControl label="Primary Color">
          <Input v-model="brandingSettings.primary_color" type="color" class="w-32" />
        </FormControl>
      </div>
      <template #footer>
        <div class="flex justify-end">
          <Button variant="primary" @click="saveBranding">Save</Button>
        </div>
      </template>
    </Card>

    <!-- Webhook Modal -->
    <FormModal v-model="showWebhookModal" title="Configure Webhook" @submit="saveWebhook">
      <template #default="{ formData }">
        <div class="space-y-4">
          <FieldInput
            v-model="formData.url as string"
            type="text"
            label="Webhook URL"
            placeholder="https://example.com/webhook"
            required
          />

          <FormControl label="Events to Subscribe">
            <div class="space-y-2">
              <label v-for="event in availableEvents" :key="event" class="flex cursor-pointer items-center gap-3">
                <Switch
                  size="sm"
                  :checked="Array.isArray(formData.events) && formData.events.includes(event)"
                  @change="toggleEvent(formData, event)"
                />
                <span class="text-sm text-gray-700">{{ event }}</span>
              </label>
            </div>
          </FormControl>

          <FieldInput
            v-model="formData.secret as string"
            type="text"
            label="Secret (for signature verification)"
            placeholder="Optional"
          />
        </div>
      </template>
    </FormModal>

    <!-- API Key Modal -->
    <FormModal v-model="showAPIKeyModal" title="Create API Key" @submit="saveAPIKey">
      <template #default="{ formData }">
        <div class="space-y-4">
          <FieldInput
            v-model="formData.name as string"
            type="text"
            label="Key Name"
            placeholder="Production Integration"
            required
          />

          <FieldInput v-model="formData.expires_at as string" type="date" label="Expires At (optional)" />

          <Alert v-if="newAPIKey" variant="info">
            <template #icon>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                class="h-6 w-6 shrink-0 stroke-current"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </template>
            <div>
              <p class="font-bold">Save this key!</p>
              <p class="text-sm">{{ newAPIKey }}</p>
              <p class="mt-1 text-xs">It won't be shown again.</p>
            </div>
          </Alert>
        </div>
      </template>
    </FormModal>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import ResourceTable from "@/components/common/ResourceTable.vue";
import FormModal from "@/components/common/FormModal.vue";
import FieldInput from "@/components/common/FieldInput.vue";
import Tabs from "@/components/ui/Tabs.vue";
import Tab from "@/components/ui/Tab.vue";
import Card from "@/components/ui/Card.vue";
import FormControl from "@/components/ui/FormControl.vue";
import Input from "@/components/ui/Input.vue";
import Select from "@/components/ui/Select.vue";
import Button from "@/components/ui/Button.vue";
import Badge from "@/components/ui/Badge.vue";
import FileInput from "@/components/ui/FileInput.vue";
import Switch from "@/components/ui/Switch.vue";
import Alert from "@/components/ui/Alert.vue";
import { fetchWithAuth } from "@/utils/api/auth";

const activeTab = ref("smtp");
const tabs = [
  { id: "smtp", label: "Email (SMTP)" },
  { id: "storage", label: "Storage" },
  { id: "webhooks", label: "Webhooks" },
  { id: "api_keys", label: "API Keys" },
  { id: "branding", label: "Branding" }
];

const smtpSettings = ref({
  host: "",
  port: 587,
  encryption: "tls",
  username: "",
  password: "",
  from_email: ""
});

const storageSettings = ref({
  provider: "local",
  local_path: "/var/lib/gosign/storage",
  s3_bucket: "",
  s3_region: "us-east-1",
  s3_access_key: "",
  s3_secret_key: ""
});

const brandingSettings = ref({
  company_name: "",
  logo_url: "",
  primary_color: "#4F46E5"
});

const webhooks = ref([]);
interface APIKey {
  id: string;
  name: string;
  enabled: boolean;
  last_used_at?: string;
  expires_at?: string;
}

const apiKeys = ref<APIKey[]>([]);
const showWebhookModal = ref(false);
const showAPIKeyModal = ref(false);
const newAPIKey = ref("");

const availableEvents = [
  "submission.created",
  "submission.sent",
  "submission.completed",
  "submitter.opened",
  "submitter.completed",
  "submitter.declined"
];

const webhookColumns = [
  { key: "url", label: "URL", sortable: true },
  { key: "events", label: "Events" },
  { key: "enabled", label: "Status" }
];

const apiKeyColumns = [
  { key: "name", label: "Name", sortable: true },
  {
    key: "last_used_at",
    label: "Last Used",
    formatter: (value: unknown): string => (value ? new Date(String(value)).toLocaleDateString() : "Never")
  },
  {
    key: "expires_at",
    label: "Expires",
    formatter: (value: unknown): string => (value ? new Date(String(value)).toLocaleDateString() : "Never")
  },
  { key: "enabled", label: "Status" }
];

onMounted(async () => {
  await Promise.all([loadSettings(), loadWebhooks(), loadAPIKeys()]);
});

async function loadSettings(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/settings");
    if (response.ok) {
      const data = await response.json();
      if (data.smtp) {
        smtpSettings.value = data.smtp;
      }
      if (data.storage) {
        storageSettings.value = data.storage;
      }
      if (data.branding) {
        brandingSettings.value = data.branding;
      }
    }
  } catch (error) {
    console.error("Failed to load settings:", error);
  }
}

async function saveSmtp(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/settings/email", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(smtpSettings.value)
    });

    if (response.ok) {
      alert("SMTP settings saved successfully");
    } else {
      alert("Failed to save SMTP settings");
    }
  } catch (error) {
    console.error("Failed to save SMTP settings:", error);
    alert("Failed to save SMTP settings");
  }
}

async function testSmtp(): Promise<void> {
  alert("Sending test email...");
  // TODO: Implement test email endpoint
}

async function saveStorage(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/settings/storage", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(storageSettings.value)
    });

    if (response.ok) {
      alert("Storage settings saved successfully");
    } else {
      alert("Failed to save storage settings");
    }
  } catch (error) {
    console.error("Failed to save storage settings:", error);
    alert("Failed to save storage settings");
  }
}

async function saveBranding(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/settings/branding", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(brandingSettings.value)
    });

    if (response.ok) {
      alert("Branding settings saved successfully");
    } else {
      alert("Failed to save branding settings");
    }
  } catch (error) {
    console.error("Failed to save branding settings:", error);
    alert("Failed to save branding settings");
  }
}

async function loadWebhooks(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/webhooks");
    if (response.ok) {
      const data = await response.json();
      webhooks.value = data.data || [];
    }
  } catch (error) {
    console.error("Failed to load webhooks:", error);
  }
}

async function loadAPIKeys(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/apikeys");
    if (response.ok) {
      const data = await response.json();
      apiKeys.value = data.data || [];
    }
  } catch (error) {
    console.error("Failed to load API keys:", error);
  }
}

function openWebhookModal(): void {
  showWebhookModal.value = true;
}

function openAPIKeyModal(): void {
  newAPIKey.value = "";
  showAPIKeyModal.value = true;
}

async function saveWebhook(formData: any): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/webhooks", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(formData)
    });

    if (response.ok) {
      showWebhookModal.value = false;
      await loadWebhooks();
    } else {
      alert("Failed to save webhook");
    }
  } catch (error) {
    console.error("Failed to save webhook:", error);
    alert("Failed to save webhook");
  }
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
      // Don't close modal immediately so user can copy key
      setTimeout(() => {
        if (confirm("Have you saved the API key?")) {
          showAPIKeyModal.value = false;
          newAPIKey.value = "";
        }
      }, 2000);
    } else {
      alert("Failed to create API key");
    }
  } catch (error) {
    console.error("Failed to create API key:", error);
    alert("Failed to create API key");
  }
}

function editWebhook(webhook: any): void {
  console.log("Edit webhook:", webhook);
  // TODO: Implement edit
}

async function deleteWebhook(webhook: any): Promise<void> {
  if (!confirm("Are you sure you want to delete this webhook?")) {
    return;
  }

  try {
    const response = await fetchWithAuth(`/api/v1/webhooks/${webhook.id}`, {
      method: "DELETE"
    });

    if (response.ok) {
      await loadWebhooks();
    } else {
      alert("Failed to delete webhook");
    }
  } catch (error) {
    console.error("Failed to delete webhook:", error);
    alert("Failed to delete webhook");
  }
}

async function deleteAPIKey(apiKey: any): Promise<void> {
  if (!confirm("Are you sure you want to delete this API key?")) {
    return;
  }

  try {
    const response = await fetchWithAuth(`/api/v1/apikeys/${apiKey.id}`, {
      method: "DELETE"
    });

    if (response.ok) {
      await loadAPIKeys();
    } else {
      alert("Failed to delete API key");
    }
  } catch (error) {
    console.error("Failed to delete API key:", error);
    alert("Failed to delete API key");
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
      alert(`Failed to ${action} API key`);
    }
  } catch (error) {
    console.error(`Failed to ${action} API key:`, error);
    alert(`Failed to ${action} API key`);
  }
}

function toggleEvent(formData: Record<string, unknown>, event: string): void {
  if (!Array.isArray(formData.events)) {
    formData.events = [];
  }

  const events = formData.events as string[];
  const index = events.indexOf(event);
  if (index > -1) {
    events.splice(index, 1);
  } else {
    events.push(event);
  }
}

function handleLogoUpload(event: Event): void {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (file) {
    const reader = new FileReader();
    reader.onload = (e) => {
      brandingSettings.value.logo_url = e.target?.result as string;
    };
    reader.readAsDataURL(file);
  }
}
</script>

<style scoped>
.settings-page {
  @apply min-h-full;
}
</style>
