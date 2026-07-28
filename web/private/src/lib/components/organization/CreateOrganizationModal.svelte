<script lang="ts">
  import { t } from "@/i18n/index.svelte";
  import { apiPost } from "@/services/api";
  import FormModal from "@/components/common/FormModal.svelte";
  import FieldInput from "@/components/common/FieldInput.svelte";
  import FormControl from "@/components/ui/FormControl.svelte";
  import Button from "@/components/ui/Button.svelte";

  interface Props {
    open?: boolean;
    onClose?: () => void;
    onCreated?: (organization: any) => void;
  }

  let { open = $bindable(false), onClose, onCreated }: Props = $props();

  function handleClose(): void {
    open = false;
    onClose?.();
  }

  function setOpen(value: boolean): void {
    open = value;
    if (!value) {
      onClose?.();
    }
  }

  const createOrganization = async (formData: Record<string, unknown>) => {
    const name = (formData.name as string)?.trim();
    if (!name) {
      return;
    }

    try {
      const response = await apiPost("/company", {
        name: name,
        description: (formData.description as string)?.trim() || ""
      });

      // Response structure: { data: { organization: {...} } }
      // Try to extract organization from response
      let organization = null;

      if (response.data) {
        // Check if response.data has organization property
        if (response.data.organization) {
          organization = response.data.organization;
        } else if (response.data.id) {
          // If response.data itself is the organization object
          organization = response.data;
        }
      }

      if (organization && organization.id) {
        onCreated?.(organization);
        setOpen(false);
      } else {
        console.error("Invalid response structure:", response);
        alert(t("organizations.createError"));
        // Don't close modal if response is invalid
      }
    } catch (error: any) {
      // Don't log if we're being redirected to login
      const isRedirecting = window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin");

      if (!isRedirecting) {
        console.error("Failed to create organization:", error);
        alert(error.message || t("organizations.createError"));
      }
      // If redirecting, close modal silently
      if (isRedirecting) {
        setOpen(false);
      }
    }
  };
</script>

<FormModal
  bind:open
  title={t("organizations.createOrganization")}
  submitText={t("organizations.createOrganization")}
  cancelText={t("common.cancel")}
  onSubmitEvent={createOrganization}
  onClose={handleClose}
>
  {#snippet children(formData)}
    <div class="space-y-4">
      <p class="text-sm text-gray-500">{t("organizations.createDescription")}</p>

      <FieldInput
        bind:value={formData.name}
        type="text"
        placeholder={t("organizations.enterOrganizationName")}
        required
      />

      <FormControl label={t("organizations.description")}>
        <textarea
          bind:value={formData.description}
          rows="3"
          class="min-h-[3rem] w-full rounded-lg border border-[var(--color-base-300)] bg-[var(--color-base-100)] px-4 py-3 text-sm text-[var(--color-base-content)] transition-all duration-200 hover:border-[var(--color-base-content)]/20 focus:border-[var(--color-primary)] focus:outline-2 focus:outline-offset-2 focus:outline-[var(--color-primary)] focus:outline-none"
          placeholder={t("organizations.describeOrganization")}></textarea>
      </FormControl>
    </div>
  {/snippet}

  {#snippet footer(submit, cancel, isSubmitting)}
    <div class="flex justify-end gap-3">
      <Button variant="ghost" disabled={isSubmitting} onclick={cancel}>
        {t("common.cancel")}
      </Button>
      <Button variant="primary" loading={isSubmitting} disabled={isSubmitting} onclick={submit}>
        {t("organizations.createOrganization")}
      </Button>
    </div>
  {/snippet}
</FormModal>
