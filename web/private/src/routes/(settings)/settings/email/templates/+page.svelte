<script lang="ts">
  import type { Snippet } from "svelte";
  import { onMount, tick } from "svelte";
  import { t, getLocale, SUPPORTED_LOCALES } from "@/i18n/index.svelte";
  import { ApiError, apiDelete, apiGet, apiPost, apiPut } from "@/services/api";
  import ResourceTable from "@/components/common/ResourceTable.svelte";
  import FormControl from "@/components/ui/FormControl.svelte";
  import Input from "@/components/ui/Input.svelte";
  import FormModal from "@/components/common/FormModal.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

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

  let templates = $state<EmailTemplate[]>([]);
  let editingTemplate = $state<EmailTemplate | null>(null);
  let isEditModalOpen = $state(false);
  let formModalRef: ReturnType<typeof FormModal> | undefined = $state();
  let selectedLocale = $state<string>(getLocale());
  const supportedLocales = SUPPORTED_LOCALES;

  const templateColumns = $derived.by(() => [
    { key: "name", label: t("settings.templateName"), sortable: true },
    { key: "subject", label: t("settings.subject") },
    {
      key: "updated_at",
      label: t("settings.lastUpdated"),
      formatter: (value: unknown): string => formatDate(String(value))
    }
  ]);

  onMount(async () => {
    await loadTemplates();
  });

  async function loadTemplates(): Promise<void> {
    try {
      const url = `/settings/email/templates?locale=${selectedLocale}`;
      const response = await apiGet<{ templates?: EmailTemplate[] }>(url);
      const raw = response?.data;
      if (raw && typeof raw === "object" && "templates" in raw && Array.isArray(raw.templates)) {
        templates = raw.templates;
      } else if (Array.isArray(raw)) {
        templates = raw;
      } else {
        templates = [];
      }
    } catch (error) {
      console.error("Failed to load email templates:", error);
      templates = [];
    }
  }

  function getTemplateDisplayName(name: string): string {
    const names: Record<string, string> = {
      base: t("settings.templateBase"),
      invitation: t("settings.templateInvitation"),
      reminder: t("settings.templateReminder"),
      completed: t("settings.templateCompleted")
    };
    return names[name] || name;
  }

  function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString();
  }

  async function editTemplate(template: EmailTemplate): Promise<void> {
    editingTemplate = { ...template };
    isEditModalOpen = true;

    // Initialize formData with template data
    await tick();
    if (formModalRef) {
      formModalRef.setFormData({
        locale: template.locale || selectedLocale,
        subject: template.subject || "",
        content: template.content
      });
    }
  }

  async function handleLocaleChange(newLocale: string, formData: Record<string, unknown>): Promise<void> {
    if (!editingTemplate?.name) {
      return;
    }

    const templateName = editingTemplate.name;
    formData.locale = newLocale;

    // Helper to reset template for new translation
    const resetForNewTranslation = () => {
      formData.subject = "";
      formData.content = "";
      if (editingTemplate) {
        const current = editingTemplate;
        editingTemplate = {
          id: "",
          locale: newLocale,
          name: templateName,
          content: "",
          subject: "",
          is_system: current.is_system,
          created_at: current.created_at || new Date().toISOString(),
          updated_at: current.updated_at || new Date().toISOString()
        };
      }
    };

    try {
      const data: any = await apiGet(`/settings/email/templates/${templateName}?locale=${newLocale}`);
      const template = data.data?.template || data.template;

      if (template) {
        formData.subject = template.subject || "";
        formData.content = template.content || "";
        editingTemplate = { ...template };
      } else {
        resetForNewTranslation();
      }
    } catch (error) {
      // HTTP error (e.g. 404) just means there is no translation yet.
      if (!(error instanceof ApiError)) {
        console.error("Failed to load template for locale:", error);
      }
      resetForNewTranslation();
    }
  }

  function closeEditModal(): void {
    editingTemplate = null;
    isEditModalOpen = false;
  }

  async function saveTemplate(formData: Record<string, unknown>): Promise<void> {
    if (!editingTemplate || !editingTemplate.name) {
      return;
    }

    try {
      const locale = (formData.locale as string) || selectedLocale;
      const templateName = editingTemplate.name;

      // Use existing template ID if available, otherwise check if template exists
      let templateId: string | null = editingTemplate.id || null;

      if (!templateId) {
        try {
          const checkData: any = await apiGet(`/settings/email/templates/${templateName}?locale=${locale}`);
          const existingTemplate = checkData.data?.template || checkData.template;
          if (existingTemplate) {
            templateId = existingTemplate.id;
          }
        } catch (error) {
          // HTTP error means the template doesn't exist yet — create a new one.
          if (!(error instanceof ApiError)) {
            throw error;
          }
        }
      }

      const payload = {
        name: templateName,
        locale: locale,
        subject: formData.subject || "",
        content: formData.content
      };

      if (templateId) {
        await apiPut(`/settings/email/templates/${templateId}`, payload);
      } else {
        await apiPost("/settings/email/templates", payload);
      }

      await loadTemplates();
      closeEditModal();
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
      await apiDelete(`/settings/email/templates/${template.id}`);
      await loadTemplates();
    } catch (error) {
      console.error("Failed to delete email template:", error);
      alert(t("settings.deleteError"));
    }
  }
</script>

{#snippet nameCell(item: unknown, _value: string)}
  <span class="font-medium">{getTemplateDisplayName((item as EmailTemplate).name)}</span>
{/snippet}

{#snippet subjectCell(_item: unknown, value: string)}
  {#if value}
    <span class="text-gray-700">{value}</span>
  {:else}
    <span class="text-gray-400 italic">{t("common.optional")}</span>
  {/if}
{/snippet}

{#snippet actionsCell(item: unknown)}
  <div class="flex justify-end gap-1">
    {#if !(item as EmailTemplate).is_system}
      <button
        class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
        title={t("common.delete")}
        onclick={(e) => {
          e.stopPropagation();
          deleteTemplate(item as EmailTemplate);
        }}
      >
        <SvgIcon name="trash-x" class="h-5 w-5 stroke-[2]" />
      </button>
    {/if}
    <button
      class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
      title={t("common.edit")}
      onclick={(e) => {
        e.stopPropagation();
        editTemplate(item as EmailTemplate);
      }}
    >
      <SvgIcon name="settings" class="h-5 w-5 stroke-[2]" />
    </button>
  </div>
{/snippet}

<div class="space-y-4">
  <!-- Language Filter -->
  <div class="flex items-center gap-4">
    <FormControl label={t("settings.language")}>
      <div class="flex flex-wrap gap-2">
        {#each Object.entries(supportedLocales) as [code, name] (code)}
          <button
            class={[
              "cursor-pointer rounded-md border px-4 py-2 text-sm font-medium transition-colors",
              selectedLocale === code
                ? "border-blue-500 bg-blue-50 text-blue-700"
                : "border-gray-300 bg-white text-gray-700 hover:bg-gray-50"
            ].join(" ")}
            onclick={() => {
              selectedLocale = code;
              loadTemplates();
            }}
          >
            {name}
          </button>
        {/each}
      </div>
    </FormControl>
  </div>

  <ResourceTable
    data={templates}
    columns={templateColumns}
    hasActions={true}
    showPagination={false}
    searchable={false}
    showFilters={false}
    showEdit={false}
    showDelete={false}
    emptyMessage={t("settings.noEmailTemplates")}
    cellSnippets={{
      name: nameCell as Snippet<[item: unknown, value: string]>,
      subject: subjectCell as Snippet<[item: unknown, value: string]>
    }}
    actions={actionsCell}
  />

  <!-- Edit Modal -->
  {#if editingTemplate}
    <FormModal
      bind:this={formModalRef}
      bind:open={isEditModalOpen}
      title={editingTemplate.id ? t("settings.editEmailTemplate") : t("settings.createEmailTemplate")}
      size="xl"
      onClose={closeEditModal}
      onSubmit={saveTemplate}
    >
      {#snippet children(formData)}
        <FormControl label={t("settings.language")}>
          <div class="flex flex-wrap gap-2">
            {#each Object.entries(supportedLocales) as [code, name] (code)}
              <button
                class={[
                  "cursor-pointer rounded-md border px-4 py-2 text-sm font-medium transition-colors",
                  (formData.locale || selectedLocale) === code
                    ? "border-blue-500 bg-blue-50 text-blue-700"
                    : "border-gray-300 bg-white text-gray-700 hover:bg-gray-50"
                ].join(" ")}
                onclick={() => handleLocaleChange(code, formData)}
              >
                {name}
              </button>
            {/each}
          </div>
        </FormControl>

        <FormControl label={t("settings.subject")}>
          <Input
            value={String(formData.subject || "")}
            placeholder={t("settings.emailSubjectPlaceholder")}
            oninput={(e) => (formData.subject = e.currentTarget.value)}
          />
        </FormControl>

        <FormControl label={t("settings.templateContent")} required>
          <textarea
            value={String(formData.content || "")}
            class="min-h-[400px] w-full rounded-md border border-gray-300 px-3 py-2 font-mono text-sm"
            placeholder={t("settings.emailTemplatePlaceholder")}
            oninput={(e) => (formData.content = e.currentTarget.value)}></textarea>
        </FormControl>

        <div class="rounded-md bg-blue-50 p-3 text-sm text-blue-800">
          <p class="font-medium">{t("settings.availableVariables")}:</p>
          <ul class="mt-1 list-inside list-disc space-y-1">
            <li>{t("settings.variableRecipientName")}</li>
            <li>{t("settings.variableDocumentName")}</li>
            <li>{t("settings.variableSigningLink")}</li>
            <li>{t("settings.variableExpiresAt")}</li>
            <li>{t("settings.variableSenderName")}</li>
            <li>{t("settings.variableCustomMessage")}</li>
          </ul>
        </div>
      {/snippet}
    </FormModal>
  {/if}
</div>
