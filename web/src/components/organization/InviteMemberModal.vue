<template>
  <div class="fixed inset-0 z-50 overflow-y-auto">
    <div class="flex min-h-screen items-end justify-center px-4 pt-4 pb-20 text-center sm:block sm:p-0">
      <!-- Background overlay -->
      <div class="bg-opacity-75 fixed inset-0 bg-gray-500 transition-opacity" @click="$emit('close')"></div>

      <!-- Modal panel -->
      <div
        class="inline-block transform overflow-hidden rounded-lg border border-gray-200 bg-white text-left align-bottom transition-all hover:border-gray-300 sm:my-8 sm:w-full sm:max-w-lg sm:align-middle"
      >
        <form @submit.prevent="inviteMember">
          <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <div class="sm:flex sm:items-start">
              <div
                class="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-blue-100 sm:mx-0 sm:h-10 sm:w-10"
              >
                <UserPlusIcon class="h-6 w-6 text-blue-600" />
              </div>
              <div class="mt-3 flex-1 text-center sm:mt-0 sm:ml-4 sm:text-left">
                <h3 class="text-lg leading-6 font-medium text-gray-900">Invite Team Member</h3>
                <div class="mt-2">
                  <p class="text-sm text-gray-500">
                    Invite a new member to join {{ props.organization?.name || "your organization" }}.
                  </p>
                </div>
              </div>
            </div>

            <!-- Form fields -->
            <div class="mt-6 space-y-4">
              <div>
                <label for="email" class="block text-sm font-medium text-gray-700"> Email Address * </label>
                <input
                  id="email"
                  v-model="form.email"
                  type="email"
                  required
                  class="mt-1 block w-full rounded-md border border-gray-300 focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                  placeholder="member@example.com"
                />
              </div>

              <div>
                <label for="role" class="block text-sm font-medium text-gray-700"> Role * </label>
                <select
                  id="role"
                  v-model="form.role"
                  required
                  class="mt-1 block w-full rounded-md border-gray-300 py-2 pr-10 pl-3 text-sm focus:border-blue-500 focus:ring-blue-500 focus:outline-none"
                >
                  <option value="viewer">Viewer - Can view documents and submissions</option>
                  <option value="member">Member - Can create and edit documents</option>
                  <option v-if="canInviteAsAdmin" value="admin">
                    Admin - Can manage members and organization settings
                  </option>
                </select>
              </div>

              <!-- Success message -->
              <div v-if="invitationSent" class="rounded-md border border-green-200 bg-green-50 p-4">
                <div class="flex">
                  <div class="flex-shrink-0">
                    <CheckCircleIcon class="h-5 w-5 text-green-400" />
                  </div>
                  <div class="ml-3">
                    <h3 class="text-sm font-medium text-green-800">Invitation sent successfully!</h3>
                    <div class="mt-2 text-sm text-green-700">
                      <p>{{ form.email }} has been invited to join {{ organization?.name }} as a {{ form.role }}.</p>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Error message -->
              <div v-if="error" class="rounded-md border border-red-200 bg-red-50 p-4">
                <div class="flex">
                  <div class="flex-shrink-0">
                    <XCircleIcon class="h-5 w-5 text-red-400" />
                  </div>
                  <div class="ml-3">
                    <h3 class="text-sm font-medium text-red-800">Failed to send invitation</h3>
                    <div class="mt-2 text-sm text-red-700">
                      <p>{{ error }}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="bg-gray-50 px-4 py-3 sm:flex sm:flex-row-reverse sm:px-6">
            <button
              v-if="!invitationSent"
              type="submit"
              :disabled="loading"
              class="inline-flex w-full justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-base font-medium text-white hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none disabled:opacity-50 sm:ml-3 sm:w-auto sm:text-sm"
            >
              <span v-if="loading" class="flex items-center">
                <div class="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-white"></div>
                Sending...
              </span>
              <span v-else>Send Invitation</span>
            </button>
            <button
              type="button"
              class="mt-3 inline-flex w-full justify-center rounded-md border border-gray-200 bg-white px-4 py-2 text-base font-medium text-gray-700 transition-colors hover:border-gray-300 hover:bg-gray-50 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
              @click="closeModal"
            >
              {{ invitationSent ? "Close" : "Cancel" }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { CheckCircleIcon, UserPlusIcon, XCircleIcon } from "@heroicons/vue/24/outline";
import { apiPost } from "@/services/api";

interface Props {
  organization: {
    id: string;
    name: string;
    description?: string;
    owner_id: string;
    created_at: any;
    updated_at: any;
  } | null;
}

const props = defineProps<Props>();

interface Emits {
  (e: "close"): void;
  (e: "invited"): void;
}

const emit = defineEmits<Emits>();

const form = ref({
  email: "",
  role: "member"
});

const loading = ref(false);
const invitationSent = ref(false);
const error = ref("");

// TODO: Get current user role from organization context
const currentUserRole = ref("owner");

const canInviteAsAdmin = computed(() => {
  return currentUserRole.value === "owner";
});

const inviteMember = async () => {
  if (!form.value.email.trim() || !form.value.role || !props.organization) {
    return;
  }

  try {
    loading.value = true;
    error.value = "";

    await apiPost(`/api/v1/organizations/${props.organization.id}/members/invite`, {
      email: form.value.email.trim(),
      role: form.value.role
    });

    invitationSent.value = true;
    emit("invited");
  } catch (err: any) {
    console.error("Failed to invite member:", err);
    error.value = err.response?.data?.error || "Failed to send invitation";
  } finally {
    loading.value = false;
  }
};

const closeModal = () => {
  if (invitationSent.value) {
    // Reset form for next use
    form.value = {
      email: "",
      role: "member"
    };
    invitationSent.value = false;
    error.value = "";
  }
  emit("close");
};
</script>
