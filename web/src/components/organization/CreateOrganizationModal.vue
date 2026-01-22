<template>
  <FormModal
    v-model="isOpen"
    :title="$t('organizations.createOrganization')"
    :submit-text="$t('organizations.createOrganization')"
    :cancel-text="$t('common.cancel')"
    @submit="createOrganization"
    @close="handleClose"
  >
    <template #default="{ formData }">
      <div class="space-y-4">
        <p class="text-sm text-gray-500">{{ $t('organizations.createDescription') }}</p>
        
        <FieldInput
          v-model="formData.name as string"
          type="text"
          :label="$t('organizations.organizationName')"
          :placeholder="$t('organizations.enterOrganizationName')"
          required
        />

        <FormControl :label="$t('organizations.description')">
          <textarea
            v-model="formData.description as string"
            rows="3"
            class="min-h-[3rem] w-full rounded-lg border border-[var(--color-base-300)] bg-[var(--color-base-100)] px-4 py-3 text-sm text-[var(--color-base-content)] transition-all duration-200 hover:border-[var(--color-base-content)]/20 focus:border-[var(--color-primary)] focus:outline-2 focus:outline-offset-2 focus:outline-[var(--color-primary)] focus:outline-none"
            :placeholder="$t('organizations.describeOrganization')"
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
          {{ $t('organizations.createOrganization') }}
        </Button>
      </div>
    </template>
  </FormModal>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { apiPost } from "@/services/api";
import FormModal from "@/components/common/FormModal.vue";
import FieldInput from "@/components/common/FieldInput.vue";
import FormControl from "@/components/ui/FormControl.vue";
import Button from "@/components/ui/Button.vue";

const { t } = useI18n();

interface Props {
  modelValue?: boolean;
}

interface Emits {
  (e: "update:modelValue", value: boolean): void;
  (e: "close"): void;
  (e: "created", organization: any): void;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false
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

const createOrganization = async (formData: Record<string, unknown>) => {
  const name = (formData.name as string)?.trim();
  if (!name) {
    return;
  }

  try {
    const response = await apiPost("/api/v1/organizations", {
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
      emit("created", organization);
      isOpen.value = false;
      emit("close");
    } else {
      console.error("Invalid response structure:", response);
      alert(t('organizations.createError'));
      // Don't close modal if response is invalid
    }
  } catch (error: any) {
    // Don't log if we're being redirected to login
    const isRedirecting = window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin");

    if (!isRedirecting) {
      console.error("Failed to create organization:", error);
      alert(error.message || t('organizations.createError'));
    }
    // If redirecting, close modal silently
    if (isRedirecting) {
      isOpen.value = false;
      emit("close");
    }
  }
};

</script>
