<template>
  <FormModal
    ref="formModalRef"
    v-model="isOpen"
    :title="$t('organizations.editOrganization')"
    :submit-text="$t('common.save')"
    :cancel-text="$t('common.cancel')"
    @submit="updateOrganization"
    @close="handleClose"
  >
    <template #default="{ formData }">
      <div class="space-y-4">
        <FieldInput
          v-model="localName"
          type="text"
          :label="$t('organizations.organizationName')"
          :placeholder="$t('organizations.enterOrganizationName')"
          required
          @update:model-value="(val: string) => { formData.name = val; localName = val; }"
        />

        <FormControl :label="$t('organizations.description')">
          <textarea
            v-model="localDescription"
            rows="3"
            class="min-h-[3rem] w-full rounded-lg border border-[var(--color-base-300)] bg-[var(--color-base-100)] px-4 py-3 text-sm text-[var(--color-base-content)] transition-all duration-200 hover:border-[var(--color-base-content)]/20 focus:border-[var(--color-primary)] focus:outline-2 focus:outline-offset-2 focus:outline-[var(--color-primary)] focus:outline-none"
            :placeholder="$t('organizations.describeOrganization')"
            @input="(e: Event) => { const val = (e.target as HTMLTextAreaElement).value; formData.description = val; localDescription = val; }"
          ></textarea>
        </FormControl>
      </div>
    </template>
    
    <template #footer="{ submit, cancel, isSubmitting }">
      <div class="flex justify-end gap-3">
        <Button variant="ghost" :disabled="isSubmitting" @click="cancel">
          {{ $t('common.cancel') }}
        </Button>
        <Button variant="primary" :loading="isSubmitting" :disabled="isSubmitting" @click="submit">
          {{ $t('common.save') }}
        </Button>
      </div>
    </template>
  </FormModal>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { apiPut } from "@/services/api";
import FormModal from "@/components/common/FormModal.vue";
import FieldInput from "@/components/common/FieldInput.vue";
import FormControl from "@/components/ui/FormControl.vue";
import Button from "@/components/ui/Button.vue";
import type { Organization } from "@/models";

const { t } = useI18n();
const formModalRef = ref<InstanceType<typeof FormModal> | null>(null);
const localName = ref("");
const localDescription = ref("");

interface Props {
  modelValue?: boolean;
  organization?: Organization | null;
}

interface Emits {
  (e: "update:modelValue", value: boolean): void;
  (e: "close"): void;
  (e: "updated", organization: Organization): void;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  organization: null
});

const emit = defineEmits<Emits>();

const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => {
    emit("update:modelValue", value);
    if (!value) {
      emit("close");
    }
  }
});

function handleClose(): void {
  isOpen.value = false;
}

// Watch for modal opening and organization changes to initialize form data
watch(
  [() => isOpen.value, () => props.organization],
  async ([isModalOpen, org]) => {
    if (isModalOpen && org) {
      // Initialize local values immediately
      localName.value = org.name || "";
      localDescription.value = org.description || "";
      
      // Wait for modal to fully render
      await nextTick();
      await nextTick();
      
      // Initialize formData in FormModal
      if (formModalRef.value && typeof formModalRef.value.setFormData === 'function') {
        formModalRef.value.setFormData({
          name: localName.value,
          description: localDescription.value
        });
      }
    } else if (!isModalOpen) {
      // Reset form when modal closes
      localName.value = "";
      localDescription.value = "";
      if (formModalRef.value && typeof formModalRef.value.resetForm === 'function') {
        formModalRef.value.resetForm();
      }
    }
  },
  { immediate: true }
);

const updateOrganization = async (formData: Record<string, unknown>) => {
  if (!props.organization?.id) {
    return;
  }

  // Use local values if formData is empty, otherwise use formData
  const name = ((formData.name as string) || localName.value)?.trim();
  const description = ((formData.description as string) || localDescription.value)?.trim() || "";
  
  if (!name) {
    return;
  }

  try {
    const response = await apiPut(`/api/v1/organizations/${props.organization.id}`, {
      name: name,
      description: description
    });


    // Response structure: { success: true, message: "...", data: { organization: {...} } }
    let organization = null;

    if (response && response.data) {
      // Check if data.organization exists (nested structure)
      if (response.data.organization && typeof response.data.organization === 'object') {
        organization = response.data.organization;
      } 
      // Check if data itself is the organization (has id field)
      else if (response.data.id) {
        organization = response.data;
      }
    }

    // If organization is still null, create updated object from form data and current organization
    if (!organization || !organization.id) {
      console.log("Creating organization object from form data");
      organization = {
        ...props.organization,
        name: name,
        description: description
      };
    }

    // Always emit updated event with the organization data
    if (organization && organization.id) {
      emit("updated", organization as Organization);
      isOpen.value = false;
      emit("close");
    } else {
      console.error("Invalid response structure - no valid organization found:", response);
      alert(t('organizations.updateError') || 'Failed to update organization');
    }
  } catch (error: any) {
    const isRedirecting = window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin");

    if (!isRedirecting) {
      console.error("Failed to update organization:", error);
      alert(error.message || t('organizations.updateError') || 'Failed to update organization');
    }
    if (isRedirecting) {
      isOpen.value = false;
      emit("close");
    }
  }
};

</script>
