<template>
  <div class="space-y-4">
    <!-- Language Filter -->
    <div class="flex items-center gap-4">
      <FormControl :label="$t('settings.language')">
        <div class="flex flex-wrap gap-2">
          <button
            v-for="(name, code) in supportedLocales"
            :key="code"
            :class="[
              'rounded-md border px-4 py-2 text-sm font-medium transition-colors cursor-pointer',
              selectedLocale === code
                ? 'border-blue-500 bg-blue-50 text-blue-700'
                : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
            ]"
            @click="selectedLocale = code; loadTemplates()"
          >
            {{ name }}
          </button>
        </div>
      </FormControl>
    </div>

    <ResourceTable
      :data="templates"
      :columns="templateColumns"
      :has-actions="true"
      :show-pagination="false"
      :searchable="false"
      :show-filters="false"
      :show-edit="false"
      :show-delete="false"
      :empty-message="$t('settings.noEmailTemplates')"
    >
      <template #cell-name="{ item }">
        <span class="font-medium">{{ getTemplateDisplayName((item as EmailTemplate).name) }}</span>
      </template>

      <template #cell-subject="{ value }">
        <span v-if="value" class="text-gray-700">{{ value }}</span>
        <span v-else class="text-gray-400 italic">{{ $t('common.optional') }}</span>
      </template>

      <template #actions="{ item }">
        <div class="flex justify-end gap-1">
          <button
            v-if="!(item as EmailTemplate).is_system"
            class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
            @click.stop="deleteTemplate(item as EmailTemplate)"
            :title="$t('common.delete')"
          >
            <SvgIcon name="trash-x" class="h-5 w-5 stroke-[2]" />
          </button>
          <button
            class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
            @click.stop="editTemplate(item as EmailTemplate)"
            :title="$t('common.edit')"
          >
            <SvgIcon name="settings" class="h-5 w-5 stroke-[2]" />
          </button>
        </div>
      </template>
    </ResourceTable>

    <!-- Edit Modal -->
    <FormModal
      v-if="editingTemplate"
      ref="formModalRef"
      v-model="isEditModalOpen"
      :title="editingTemplate.id ? $t('settings.editEmailTemplate') : $t('settings.createEmailTemplate')"
      size="xl"
      @close="closeEditModal"
      @submit="saveTemplate"
    >
      <template #default="{ formData }">
        <FormControl :label="$t('settings.language')">
          <div class="flex flex-wrap gap-2">
            <button
              v-for="(name, code) in supportedLocales"
              :key="code"
              :class="[
                'rounded-md border px-4 py-2 text-sm font-medium transition-colors cursor-pointer',
                (formData.locale || selectedLocale) === code
                  ? 'border-blue-500 bg-blue-50 text-blue-700'
                  : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
              ]"
              @click="handleLocaleChange(code, formData)"
            >
              {{ name }}
            </button>
          </div>
        </FormControl>

        <FormControl :label="$t('settings.subject')">
          <Input
            :model-value="String(formData.subject || '')"
            @update:model-value="formData.subject = $event"
            :placeholder="$t('settings.emailSubjectPlaceholder')"
          />
        </FormControl>

        <FormControl :label="$t('settings.templateContent')" required>
          <textarea
            :value="String(formData.content || '')"
            @input="formData.content = ($event.target as HTMLTextAreaElement).value"
            class="min-h-[400px] w-full rounded-md border border-gray-300 px-3 py-2 font-mono text-sm"
            :placeholder="$t('settings.emailTemplatePlaceholder')"
          />
        </FormControl>

        <div class="rounded-md bg-blue-50 p-3 text-sm text-blue-800">
          <p class="font-medium">{{ $t('settings.availableVariables') }}:</p>
          <ul class="mt-1 list-inside list-disc space-y-1">
            <li>{{ $t('settings.variableRecipientName') }}</li>
            <li>{{ $t('settings.variableDocumentName') }}</li>
            <li>{{ $t('settings.variableSigningLink') }}</li>
            <li>{{ $t('settings.variableExpiresAt') }}</li>
            <li>{{ $t('settings.variableSenderName') }}</li>
            <li>{{ $t('settings.variableCustomMessage') }}</li>
          </ul>
        </div>
      </template>
    </FormModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, nextTick } from "vue";
import { useI18n } from "vue-i18n";
import { SUPPORTED_LOCALES } from "@/i18n";
import { fetchWithAuth } from "@/utils/auth";
import ResourceTable from "@/components/common/ResourceTable.vue";
import FormControl from "@/components/ui/FormControl.vue";
import Input from "@/components/ui/Input.vue";
import FormModal from "@/components/common/FormModal.vue";
import SvgIcon from "@/components/SvgIcon.vue";

const { t, locale } = useI18n();

interface EmailTemplate {
  id: string;
  account_id?: string;
  name: string;
  locale: string;
  subject?: string;
  content: string;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

const templates = ref<EmailTemplate[]>([]);
const editingTemplate = ref<EmailTemplate | null>(null);
const isEditModalOpen = ref(false);
const formModalRef = ref<InstanceType<typeof FormModal> | null>(null);
const selectedLocale = ref<string>(locale.value as string);
const supportedLocales = SUPPORTED_LOCALES;

const templateColumns = computed(() => [
  { key: "name", label: t('settings.templateName'), sortable: true },
  { key: "subject", label: t('settings.subject') },
  {
    key: "updated_at",
    label: t('settings.lastUpdated'),
    formatter: (value: unknown): string => formatDate(String(value))
  }
]);

onMounted(async () => {
  await loadTemplates();
});

async function loadTemplates(): Promise<void> {
  try {
    // Use fetchWithAuth directly like other settings pages
    const url = `/api/v1/email-templates?locale=${selectedLocale.value}`;
    const response = await fetchWithAuth(url);
    if (response.ok) {
      const data = await response.json();
      console.log("Email templates API response:", data);
      
      // API returns: { success: true, message: "...", data: { templates: [...] } }
      if (data.data && data.data.templates && Array.isArray(data.data.templates)) {
        templates.value = data.data.templates;
      } else if (data.templates && Array.isArray(data.templates)) {
        // Fallback: if templates are directly in data
        templates.value = data.templates;
      } else if (Array.isArray(data.data)) {
        // Fallback: if data is directly an array
        templates.value = data.data;
      } else {
        console.warn("Unexpected response structure:", data);
        templates.value = [];
      }
    } else {
      const errorText = await response.text().catch(() => "");
      console.error("Failed to load email templates:", response.status, response.statusText, errorText);
      templates.value = [];
    }
  } catch (error) {
    console.error("Failed to load email templates:", error);
    templates.value = [];
  }
}

function getTemplateDisplayName(name: string): string {
  const names: Record<string, string> = {
    base: t("settings.templateBase"),
    invitation: t("settings.templateInvitation"),
    reminder: t("settings.templateReminder"),
    completed: t("settings.templateCompleted"),
  };
  return names[name] || name;
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString();
}

async function editTemplate(template: EmailTemplate): Promise<void> {
  editingTemplate.value = { ...template };
  isEditModalOpen.value = true;
  
  // Initialize formData with template data
  await nextTick();
  if (formModalRef.value) {
    formModalRef.value.setFormData({
      locale: template.locale || selectedLocale.value,
      subject: template.subject || "",
      content: template.content,
    });
  }
}

async function handleLocaleChange(newLocale: string, formData: any): Promise<void> {
  if (!editingTemplate.value?.name) return;
  
  const templateName = editingTemplate.value.name;
  formData.locale = newLocale;
  
  // Helper to reset template for new translation
  const resetForNewTranslation = () => {
    formData.subject = "";
    formData.content = "";
    if (editingTemplate.value) {
      const current = editingTemplate.value;
      editingTemplate.value = {
        id: "",
        locale: newLocale,
        name: templateName,
        content: "",
        subject: "",
        is_system: current.is_system,
        created_at: current.created_at || new Date().toISOString(),
        updated_at: current.updated_at || new Date().toISOString(),
      };
    }
  };
  
  try {
    const url = `/api/v1/email-templates/${templateName}?locale=${newLocale}`;
    const response = await fetchWithAuth(url);
    
    if (response.ok) {
      const data = await response.json();
      const template = data.data?.template || data.template;
      
      if (template) {
        formData.subject = template.subject || "";
        formData.content = template.content || "";
        editingTemplate.value = { ...template };
      } else {
        resetForNewTranslation();
      }
    } else {
      resetForNewTranslation();
    }
  } catch (error) {
    console.error("Failed to load template for locale:", error);
    resetForNewTranslation();
  }
}

function closeEditModal(): void {
  editingTemplate.value = null;
  isEditModalOpen.value = false;
}

async function saveTemplate(formData: any): Promise<void> {
  if (!editingTemplate.value || !editingTemplate.value.name) return;

  try {
    const locale = formData.locale || selectedLocale.value;
    const templateName = editingTemplate.value.name;
    
    // Use existing template ID if available, otherwise check if template exists
    let templateId: string | null = editingTemplate.value.id || null;
    
    if (!templateId) {
      // Check if template exists for this locale
      const checkUrl = `/api/v1/email-templates/${templateName}?locale=${locale}`;
      const checkResponse = await fetchWithAuth(checkUrl);
      
      if (checkResponse.ok) {
        const checkData = await checkResponse.json();
        const existingTemplate = checkData.data?.template || checkData.template;
        if (existingTemplate) {
          templateId = existingTemplate.id;
        }
      }
    }
    
    const url = templateId
      ? `/api/v1/email-templates/${templateId}`
      : "/api/v1/email-templates";
    const method = templateId ? "PUT" : "POST";

    const response = await fetchWithAuth(url, {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: templateName,
        locale: locale,
        subject: formData.subject || "",
        content: formData.content,
      }),
    });

    if (response.ok) {
      await loadTemplates();
      closeEditModal();
    } else {
      alert(t("settings.saveError"));
    }
  } catch (error) {
    console.error("Failed to save email template:", error);
    alert(t("settings.saveError"));
  }
}

async function deleteTemplate(template: EmailTemplate): Promise<void> {
  if (!confirm(t("settings.confirmDeleteTemplate"))) {
    return;
  }

  try {
    const response = await fetchWithAuth(`/api/v1/email-templates/${template.id}`, {
      method: "DELETE",
    });

    if (response.ok) {
      await loadTemplates();
    } else {
      alert(t("settings.deleteError"));
    }
  } catch (error) {
    console.error("Failed to delete email template:", error);
    alert(t("settings.deleteError"));
  }
}
</script>
