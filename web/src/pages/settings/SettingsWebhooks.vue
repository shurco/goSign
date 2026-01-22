<template>
  <div class="space-y-4">
    <ResourceTable
      :data="webhooks"
      :columns="webhookColumns"
      :has-actions="true"
      :show-pagination="false"
      :searchable="false"
      :show-filters="false"
      :empty-message="$t('webhooks.noWebhooks')"
      @edit="editWebhook"
      @delete="deleteWebhook"
    >
      <template #cell-enabled="{ value }">
        <Badge :variant="value ? 'success' : 'ghost'">
          {{ value ? $t('webhooks.active') : $t('webhooks.inactive') }}
        </Badge>
      </template>

      <template #cell-events="{ value }">
        <div class="flex flex-wrap gap-1">
          <Badge v-for="event in value.slice(0, 3)" :key="event" size="sm">
            {{ event }}
          </Badge>
          <Badge v-if="value.length > 3" variant="ghost" size="sm">+{{ value.length - 3 }} {{ $t('webhooks.more') }}</Badge>
        </div>
      </template>
    </ResourceTable>

    <!-- Webhook Modal -->
    <FormModal v-model="showWebhookModal" :title="$t('webhooks.configureWebhook')" @submit="saveWebhook">
      <template #default="{ formData }">
        <div class="space-y-4">
          <FieldInput
            v-model="formData.url as string"
            type="text"
            :label="$t('webhooks.webhookUrl')"
            placeholder="https://example.com/webhook"
            required
          />

          <FormControl :label="$t('webhooks.eventsToSubscribe')">
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
            :label="$t('webhooks.secret')"
            :placeholder="$t('common.optional')"
          />
        </div>
      </template>
    </FormModal>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import ResourceTable from "@/components/common/ResourceTable.vue";
import FormModal from "@/components/common/FormModal.vue";
import FieldInput from "@/components/common/FieldInput.vue";
import FormControl from "@/components/ui/FormControl.vue";
import Badge from "@/components/ui/Badge.vue";
import Switch from "@/components/ui/Switch.vue";
import { fetchWithAuth } from "@/utils/auth";

const { t } = useI18n();

const webhooks = ref([]);
const showWebhookModal = ref(false);

// Get trigger ref from parent and register open function
const webhookModalTrigger = inject<{ value: (() => void) | null }>('webhookModalTrigger');

// Register open function with parent
if (webhookModalTrigger) {
  webhookModalTrigger.value = () => {
    openWebhookModal();
  };
}

const availableEvents = [
  "submission.created",
  "submission.sent",
  "submission.completed",
  "submitter.opened",
  "submitter.completed",
  "submitter.declined"
];

const webhookColumns = computed(() => [
  { key: "url", label: t('webhooks.url'), sortable: true },
  { key: "events", label: t('webhooks.events') },
  { key: "enabled", label: t('webhooks.status') }
]);

onMounted(async () => {
  await loadWebhooks();
});

async function loadWebhooks(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/webhooks");
    if (response.ok) {
      const data = await response.json();
      if (data.data && data.data.items) {
        webhooks.value = data.data.items || [];
      } else if (Array.isArray(data.data)) {
        webhooks.value = data.data;
      } else {
        webhooks.value = [];
      }
    }
  } catch (error) {
    if (!window.location.pathname.includes("/auth/") && !window.location.pathname.includes("/signin")) {
      console.error("Failed to load webhooks:", error);
    }
  }
}

function openWebhookModal(): void {
  showWebhookModal.value = true;
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
      alert(t('webhooks.saveError'));
    }
  } catch (error) {
    console.error("Failed to save webhook:", error);
    alert(t('webhooks.saveError'));
  }
}

function editWebhook(webhook: any): void {
}

async function deleteWebhook(webhook: any): Promise<void> {
  if (!confirm(t('webhooks.deleteConfirm'))) {
    return;
  }

  try {
    const response = await fetchWithAuth(`/api/v1/webhooks/${webhook.id}`, {
      method: "DELETE"
    });

    if (response.ok) {
      await loadWebhooks();
    } else {
      alert(t('webhooks.deleteError'));
    }
  } catch (error) {
    console.error("Failed to delete webhook:", error);
    alert(t('webhooks.deleteError'));
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
</script>
