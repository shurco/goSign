<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
    <div class="w-full max-w-md space-y-8">
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">Reset your password</h2>
        <p class="mt-2 text-center text-sm text-gray-600">
          Enter your email address and we'll send you a link to reset your password
        </p>
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
        </div>

        <div class="rounded-md shadow-sm">
          <div>
            <label for="email-address" class="sr-only">Email address</label>
            <input
              id="email-address"
              v-model="email"
              name="email"
              type="email"
              autocomplete="email"
              required
              class="relative block w-full appearance-none rounded border border-gray-300 px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
              placeholder="Email address"
            />
          </div>
        </div>

        <div>
          <button
            type="submit"
            :disabled="isLoading"
            class="group relative flex w-full justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:outline-none disabled:opacity-50"
          >
            <span v-if="isLoading">Sending...</span>
            <span v-else>Send reset link</span>
          </button>
        </div>

        <div class="text-center">
          <router-link to="/auth/signin" class="font-medium text-indigo-600 hover:text-indigo-500">
            Back to sign in
          </router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

const email = ref("");
const isLoading = ref(false);
const error = ref("");
const success = ref("");

const handleSubmit = async () => {
  error.value = "";
  success.value = "";
  isLoading.value = true;

  try {
    const response = await fetch("/auth/password/forgot", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        email: email.value
      })
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || "Failed to send reset link");
    }

    success.value = "If the email exists, a password reset link has been sent to your email";
    email.value = "";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "An error occurred";
  } finally {
    isLoading.value = false;
  }
};
</script>
