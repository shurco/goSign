<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-[9999] overflow-y-auto">
      <div class="flex min-h-screen items-end justify-center px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <!-- Background overlay -->
        <div class="bg-opacity-75 fixed inset-0 bg-gray-500 transition-opacity" @click="$emit('close')"></div>

        <!-- Modal panel -->
        <div
          class="relative z-10 inline-block transform overflow-hidden rounded-lg border border-gray-200 bg-white text-left align-bottom transition-all hover:border-gray-300 sm:my-8 sm:w-full sm:max-w-lg sm:align-middle"
        >
          <form @submit.prevent="createOrganization">
            <div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
              <div class="sm:flex sm:items-start">
                <div
                  class="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-blue-100 sm:mx-0 sm:h-10 sm:w-10"
                >
                  <PlusIcon class="h-6 w-6 text-blue-600" />
                </div>
                <div class="mt-3 flex-1 text-center sm:mt-0 sm:ml-4 sm:text-left">
                  <h3 class="text-lg leading-6 font-medium text-gray-900">Create Organization</h3>
                  <div class="mt-2">
                    <p class="text-sm text-gray-500">Create a new organization to collaborate with your team.</p>
                  </div>
                </div>
              </div>

              <!-- Form fields -->
              <div class="mt-6 space-y-4">
                <div>
                  <label for="name" class="mb-1 block text-sm font-medium text-gray-700"> Organization Name * </label>
                  <Input id="name" v-model="form.name" type="text" required placeholder="Enter organization name" />
                </div>

                <div>
                  <label for="description" class="mb-1 block text-sm font-medium text-gray-700"> Description </label>
                  <textarea
                    id="description"
                    v-model="form.description"
                    rows="3"
                    class="min-h-[3rem] w-full rounded-lg border border-[var(--color-base-300)] bg-[var(--color-base-100)] px-4 py-3 text-sm text-[var(--color-base-content)] transition-all duration-200 hover:border-[var(--color-base-content)]/20 focus:border-[var(--color-primary)] focus:outline-2 focus:outline-offset-2 focus:outline-[var(--color-primary)] focus:outline-none"
                    placeholder="Describe your organization (optional)"
                  ></textarea>
                </div>
              </div>
            </div>

            <!-- Footer -->
            <div class="bg-gray-50 px-4 py-3 sm:flex sm:flex-row-reverse sm:px-6">
              <button
                type="submit"
                :disabled="loading"
                class="inline-flex w-full justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-base font-medium text-white hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none disabled:opacity-50 sm:ml-3 sm:w-auto sm:text-sm"
              >
                <span v-if="loading" class="flex items-center">
                  <div class="mr-2 h-4 w-4 animate-spin rounded-full border-b-2 border-white"></div>
                  Creating...
                </span>
                <span v-else>Create Organization</span>
              </button>
              <button
                type="button"
                class="mt-3 inline-flex w-full justify-center rounded-md border border-gray-200 bg-white px-4 py-2 text-base font-medium text-gray-700 transition-colors hover:border-gray-300 hover:bg-gray-50 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
                @click="$emit('close')"
              >
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { PlusIcon } from "@heroicons/vue/24/outline";
import { apiPost } from "@/services/api";
import Input from "@/components/ui/Input.vue";

interface Emits {
  (e: "close"): void;
  (e: "created", organization: any): void;
}

const emit = defineEmits<Emits>();

const form = ref({
  name: "",
  description: ""
});

const loading = ref(false);

const createOrganization = async () => {
  if (!form.value.name.trim()) {
    return;
  }

  try {
    loading.value = true;
    const response = await apiPost("/api/v1/organizations", {
      name: form.value.name.trim(),
      description: form.value.description.trim()
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
      emit("close");
    } else {
      console.error("Invalid response structure:", response);
      alert("Failed to create organization: Invalid response format");
      // Don't close modal if response is invalid
    }

    // Reset form
    form.value = {
      name: "",
      description: ""
    };
  } catch (error: any) {
    // Don't log if we're being redirected to login
    const isRedirecting = window.location.pathname.includes("/auth/") || window.location.pathname.includes("/signin");

    if (!isRedirecting) {
      console.error("Failed to create organization:", error);
      // TODO: Show error message to user
      alert(error.message || "Failed to create organization. Please try again.");
    }
    // If redirecting, close modal silently
    if (isRedirecting) {
      emit("close");
    }
  } finally {
    loading.value = false;
  }
};
</script>
