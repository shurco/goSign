<template>
  <div class="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
    <div class="w-full max-w-md space-y-8">
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">Create your account</h2>
        <p class="mt-2 text-center text-sm text-gray-600">
          Or
          <router-link to="/auth/signin" class="font-medium text-indigo-600 hover:text-indigo-500">
            sign in to your account
          </router-link>
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

        <div class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label for="first-name" class="sr-only">First name</label>
              <input
                id="first-name"
                v-model="formData.firstName"
                name="first-name"
                type="text"
                autocomplete="given-name"
                required
                class="relative block w-full appearance-none rounded border border-gray-300 bg-white px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
                placeholder="First name"
              />
            </div>
            <div>
              <label for="last-name" class="sr-only">Last name</label>
              <input
                id="last-name"
                v-model="formData.lastName"
                name="last-name"
                type="text"
                autocomplete="family-name"
                required
                class="relative block w-full appearance-none rounded border border-gray-300 bg-white px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
                placeholder="Last name"
              />
            </div>
          </div>

          <div>
            <label for="email-address" class="sr-only">Email address</label>
            <input
              id="email-address"
              v-model="formData.email"
              name="email"
              type="email"
              autocomplete="email"
              required
              class="relative block w-full appearance-none rounded border border-gray-300 bg-white px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
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
              autocomplete="new-password"
              required
              minlength="8"
              class="relative block w-full appearance-none rounded border border-gray-300 bg-white px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
              placeholder="Password (min. 8 characters)"
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
              class="relative block w-full appearance-none rounded border border-gray-300 bg-white px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
              placeholder="Confirm password"
            />
          </div>
        </div>

        <div>
          <Button
            type="submit"
            variant="primary"
            class="w-full"
            :loading="isLoading"
            :disabled="isLoading"
          >
            <span v-if="isLoading">Creating account...</span>
            <span v-else>Sign up</span>
          </Button>
        </div>

        <div class="flex items-center justify-center">
          <div class="text-sm">
            <span class="text-gray-600">Or sign up with</span>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <Button variant="ghost" class="w-full" @click="handleGoogleSignIn">
            Google
          </Button>
          <Button variant="ghost" class="w-full" @click="handleGitHubSignIn">
            GitHub
          </Button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import Button from "@/components/ui/Button.vue";
// import { useRouter } from "vue-router";

// const router = useRouter();

const formData = ref({
  firstName: "",
  lastName: "",
  email: "",
  password: "",
  confirmPassword: ""
});

const isLoading = ref(false);
const error = ref("");
const success = ref("");

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
    const response = await fetch("/auth/signup", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        email: formData.value.email,
        password: formData.value.password,
        first_name: formData.value.firstName,
        last_name: formData.value.lastName
      })
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || "Failed to sign up");
    }

    success.value = "Registration successful! Please check your email to verify your account.";

    // Clear form
    formData.value = {
      firstName: "",
      lastName: "",
      email: "",
      password: "",
      confirmPassword: ""
    };
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
