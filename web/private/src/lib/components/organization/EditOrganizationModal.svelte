<script lang="ts">
  import { tick } from "svelte";
  import { t } from "@/i18n/index.svelte";
  import { apiPut } from "@/services/api";
  import FormModal from "@/components/common/FormModal.svelte";
  import FieldInput from "@/components/common/FieldInput.svelte";
  import FormControl from "@/components/ui/FormControl.svelte";
  import Button from "@/components/ui/Button.svelte";
  import type { Organization } from "@/models";

  interface Props {
    open?: boolean;
    organization?: Organization | null;
    onClose?: () => void;
    onUpdated?: (organization: Organization) => void;
  }

  let { open = $bindable(false), organization = null, onClose, onUpdated }: Props = $props();

  let formModalRef: ReturnType<typeof FormModal> | undefined = $state();
  let localName = $state("");
  let localDescription = $state("");

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

  $effect(() => {
    const isModalOpen = open;
    const org = organization;

    void (async () => {
      if (isModalOpen && org) {
        // Initialize local values immediately
        localName = org.name || "";
        localDescription = org.description || "";

        // Wait for modal to fully render
        await tick();
        await tick();

        // Initialize formData in FormModal
        if (formModalRef && typeof formModalRef.setFormData === "function") {
          formModalRef.setFormData({
            name: localName,
            description: localDescription
          });
        }
      } else if (!isModalOpen) {
        // Reset form when modal closes
        localName = "";
        localDescription = "";
        if (formModalRef && typeof formModalRef.resetForm === "function") {
          formModalRef.resetForm();
        }
      }
    })();
  });

  $effect(() => {
    if (!open || !formModalRef) {
      return;
    }
    const fd = formModalRef.getFormData();
    fd.name = localName;
    fd.description = localDescription;
  });

  const updateOrganization = async (formData: Record<string, unknown>) => {
    if (!organization?.id) {
      return;
    }

    // Use local values if formData is empty, otherwise use formData
    const name = ((formData.name as string) || localName)?.trim();
    const description = ((formData.description as string) || localDescription)?.trim() || "";

    if (!name) {
      return;
    }

    try {
      const response = await apiPut(`/company/${organization.id}`, {
        name: name,
        description: description
      });

      // Response structure: { success: true, message: "...", data: { organization: {...} } }
      let updatedOrganization = null;

      if (response && response.data) {
        // Check if data.organization exists (nested structure)
        if (response.data.organization && typeof response.data.organization === "object") {
          updatedOrganization = response.data.organization;
        }
        // Check if data itself is the organization (has id field)
        else if (response.data.id) {
          updatedOrganization = response.data;
        }
      }

      // If organization is still null, create updated object from form data and current organization
      if (!updatedOrganization || !updatedOrganization.id) {
        updatedOrganization = {
          ...organization,
          name: name,
          description: description
        };
      }

      // Always emit updated event with the organization data
      if (updatedOrganization && updatedOrganization.id) {
        onUpdated?.(updatedOrganization as Organization);
        setOpen(false);
      } else {
        console.error("Invalid response structure - no valid organization found:", response);
        alert(t("organizations.updateError") || "Failed to update organization");
      }
    } catch (error: any) {
      const isRedirecting = window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin");

      if (!isRedirecting) {
        console.error("Failed to update organization:", error);
        alert(error.message || t("organizations.updateError") || "Failed to update organization");
      }
      if (isRedirecting) {
        setOpen(false);
      }
    }
  };
</script>

<FormModal
  bind:this={formModalRef}
  bind:open
  title={t("organizations.editOrganization")}
  submitText={t("common.save")}
  cancelText={t("common.cancel")}
  onSubmitEvent={updateOrganization}
  onClose={handleClose}
>
  {#snippet children(formData)}
    <div class="space-y-4">
      <FieldInput bind:value={localName} type="text" placeholder={t("organizations.enterOrganizationName")} required />

      <FormControl label={t("organizations.description")}>
        <textarea
          bind:value={localDescription}
          rows="3"
          class="min-h-[3rem] w-full rounded-lg border border-[var(--color-base-300)] bg-[var(--color-base-100)] px-4 py-3 text-sm text-[var(--color-base-content)] transition-all duration-200 hover:border-[var(--color-base-content)]/20 focus:border-[var(--color-primary)] focus:outline-2 focus:outline-offset-2 focus:outline-[var(--color-primary)] focus:outline-none"
          placeholder={t("organizations.describeOrganization")}
          oninput={(e) => {
            const val = (e.currentTarget as HTMLTextAreaElement).value;
            formData.description = val;
            localDescription = val;
          }}></textarea>
      </FormControl>
    </div>
  {/snippet}

  {#snippet footer(submit, cancel, isSubmitting)}
    <div class="flex justify-end gap-3">
      <Button variant="ghost" disabled={isSubmitting} onclick={cancel}>
        {t("common.cancel")}
      </Button>
      <Button variant="primary" loading={isSubmitting} disabled={isSubmitting} onclick={submit}>
        {t("common.save")}
      </Button>
    </div>
  {/snippet}
</FormModal>
