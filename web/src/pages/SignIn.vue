<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
    <div class="w-full max-w-md space-y-8">
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">Sign in to your account</h2>
        <p class="mt-2 text-center text-sm text-gray-600">
          Or
          <router-link to="/auth/signup" class="font-medium text-indigo-600 hover:text-indigo-500">
            create a new account
          </router-link>
        </p>
      </div>

      <form class="mt-8 space-y-6" @submit.prevent="handleSubmit">
        <div v-if="error" class="relative rounded border border-red-400 bg-red-50 px-4 py-3 text-red-700" role="alert">
          <span class="block sm:inline">{{ error }}</span>
        </div>

        <div class="space-y-4 rounded-md border border-gray-200 bg-white p-6 transition-colors">
          <div>
            <label for="email-address" class="sr-only">Email address</label>
            <input
              id="email-address"
              v-model="formData.email"
              name="email"
              type="email"
              autocomplete="email"
              required
              :disabled="requires2FA"
              class="relative block w-full appearance-none rounded border border-gray-300 px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none disabled:bg-gray-100 disabled:opacity-50 sm:text-sm"
              placeholder="Email address"
            />
          </div>

          <div>
            <label for="password" class="sr-only">Password</label>
            <input
              id="password"
              v-model="formData.password"
              name="password"
              type="password"
              autocomplete="current-password"
              required
              :disabled="requires2FA"
              class="relative block w-full appearance-none rounded border border-gray-300 px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none disabled:bg-gray-100 disabled:opacity-50 sm:text-sm"
              placeholder="Password"
            />
          </div>

          <div v-if="requires2FA">
            <label for="code" class="sr-only">2FA Code</label>
            <input
              id="code"
              v-model="formData.code"
              name="code"
              type="text"
              inputmode="numeric"
              pattern="[0-9]*"
              maxlength="6"
              required
              class="relative block w-full appearance-none rounded border border-gray-300 px-3 py-2 text-center text-2xl tracking-widest text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
              placeholder="000000"
            />
            <p class="mt-2 text-center text-sm text-gray-500">Enter the 6-digit code from your authenticator app</p>
          </div>
        </div>

        <div class="flex items-center justify-between">
          <div class="text-sm">
            <router-link to="/auth/password/forgot" class="font-medium text-indigo-600 hover:text-indigo-500">
              Forgot your password?
            </router-link>
          </div>
        </div>

        <div>
          <button
            type="submit"
            :disabled="isLoading"
            class="group relative flex w-full justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:outline-none disabled:opacity-50"
          >
            <span v-if="isLoading">Signing in...</span>
            <span v-else>Sign in</span>
          </button>
        </div>

        <div v-if="!requires2FA" class="flex items-center justify-center">
          <div class="text-sm">
            <span class="text-gray-600">Or sign in with</span>
          </div>
        </div>

        <div v-if="!requires2FA" class="grid grid-cols-2 gap-3">
          <button
            type="button"
            class="inline-flex w-full justify-center rounded-md border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-500 transition-colors hover:border-gray-300 hover:bg-gray-50"
            @click="handleGoogleSignIn"
          >
            Google
          </button>
          <button
            type="button"
            class="inline-flex w-full justify-center rounded-md border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-500 transition-colors hover:border-gray-300 hover:bg-gray-50"
            @click="handleGitHubSignIn"
          >
            GitHub
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRouter, useRoute } from "vue-router";

const router = useRouter();
const route = useRoute();

const formData = ref({
  email: "",
  password: "",
  code: ""
});

const isLoading = ref(false);
const error = ref("");
const requires2FA = ref(false);

const handleSubmit = async () => {
  error.value = "";
  isLoading.value = true;

  try {
    const response = await fetch("/auth/signin", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        email: formData.value.email,
        password: formData.value.password,
        code: formData.value.code || undefined
      })
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || "Failed to sign in");
    }

    // Check if 2FA is required
    if (data.data?.requires_2fa) {
      requires2FA.value = true;
      return;
    }

    // Store tokens
    if (data.data?.access_token) {
      localStorage.setItem("access_token", data.data.access_token);
    }
    if (data.data?.refresh_token) {
      localStorage.setItem("refresh_token", data.data.refresh_token);
    }

    // Redirect to dashboard or to redirect query parameter if present
    const redirectPath = (route.query.redirect as string) || "/dashboard";
    router.push(redirectPath);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "An error occurred";
  } finally {
    isLoading.value = false;
  }
};

const handleGoogleSignIn = () => {
  window.location.href = "/auth/oauth/google";
};

const handleGitHubSignIn = () => {
  window.location.href = "/auth/oauth/github";
};
</script>
