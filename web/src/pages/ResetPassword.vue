<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
    <div class="w-full max-w-md space-y-8">
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">Reset your password</h2>
        <p class="mt-2 text-center text-sm text-gray-600">Enter your new password</p>
      </div>

      <form class="mt-8 space-y-6" @submit.prevent="handleSubmit">
        <div v-if="error" class="relative rounded border border-red-400 bg-red-50 px-4 py-3 text-red-700" role="alert">
          <span class="block sm:inline">{{ error }}</span>
        </div>

        <div
          v-if="success"
          class="relative rounded border border-green-400 bg-green-50 px-4 py-3 text-green-700"
          role="alert"
        >
          <span class="block sm:inline">{{ success }}</span>
          <div class="mt-2">
            <router-link to="/auth/signin" class="font-medium text-green-700 underline hover:text-green-600">
              Go to sign in
            </router-link>
          </div>
        </div>

        <div class="space-y-4 rounded-md border border-gray-200 bg-white p-6 transition-colors">
          <div>
            <label for="password" class="sr-only">New password</label>
            <input
              id="password"
              v-model="formData.password"
              name="password"
              type="password"
              autocomplete="new-password"
              required
              minlength="8"
              class="relative block w-full appearance-none rounded border border-gray-300 px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
              placeholder="New password (min. 8 characters)"
            />
          </div>

          <div>
            <label for="confirm-password" class="sr-only">Confirm password</label>
            <input
              id="confirm-password"
              v-model="formData.confirmPassword"
              name="confirm-password"
              type="password"
              autocomplete="new-password"
              required
              class="relative block w-full appearance-none rounded border border-gray-300 px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
              placeholder="Confirm password"
            />
          </div>
        </div>

        <div>
          <button
            type="submit"
            :disabled="isLoading || !token"
            class="group relative flex w-full justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:outline-none disabled:opacity-50"
          >
            <span v-if="isLoading">Resetting password...</span>
            <span v-else-if="!token">Invalid reset link</span>
            <span v-else>Reset password</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

const formData = ref({
  password: "",
  confirmPassword: ""
});

const token = ref("");
const isLoading = ref(false);
const error = ref("");
const success = ref("");

onMounted(() => {
  token.value = (route.query.token as string) || "";

  if (!token.value) {
    error.value = "Invalid or missing reset token";
  }
});

const handleSubmit = async () => {
  error.value = "";
  success.value = "";

  if (formData.value.password !== formData.value.confirmPassword) {
    error.value = "Passwords do not match";
    return;
  }

  if (formData.value.password.length < 8) {
    error.value = "Password must be at least 8 characters";
    return;
  }

  isLoading.value = true;

  try {
    const response = await fetch("/auth/password/reset", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        token: token.value,
        new_password: formData.value.password
      })
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || "Failed to reset password");
    }

    success.value = "Password reset successfully! You can now sign in with your new password.";
    formData.value = {
      password: "",
      confirmPassword: ""
    };
  } catch (err) {
    error.value = err instanceof Error ? err.message : "An error occurred";
  } finally {
    isLoading.value = false;
  }
};
</script>
