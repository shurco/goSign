<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import Button from "@/components/ui/Button.svelte";

  let formData = $state({
    password: "",
    confirmPassword: ""
  });

  let token = $state("");
  let isLoading = $state(false);
  let error = $state("");
  let success = $state("");

  onMount(() => {
    token = page.url.searchParams.get("token") || "";

    if (!token) {
      error = "Invalid or missing reset token";
    }
  });

  const handleSubmit = async (e: SubmitEvent) => {
    e.preventDefault();
    error = "";
    success = "";

    if (formData.password !== formData.confirmPassword) {
      error = "Passwords do not match";
      return;
    }

    if (formData.password.length < 8) {
      error = "Password must be at least 8 characters";
      return;
    }

    isLoading = true;

    try {
      const response = await fetch("/auth/password/reset", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          token,
          new_password: formData.password
        })
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || "Failed to reset password");
      }

      success = "Password reset successfully! You can now sign in with your new password.";
      formData = {
        password: "",
        confirmPassword: ""
      };
    } catch (err) {
      error = err instanceof Error ? err.message : "An error occurred";
    } finally {
      isLoading = false;
    }
  };
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
  <div class="w-full max-w-md space-y-8">
    <div>
      <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">Reset your password</h2>
      <p class="mt-2 text-center text-sm text-gray-600">Enter your new password</p>
    </div>

    <form class="mt-8 space-y-6" onsubmit={handleSubmit}>
      {#if error}
        <div class="relative rounded border border-red-400 bg-red-50 px-4 py-3 text-red-700" role="alert">
          <span class="block sm:inline">{error}</span>
        </div>
      {/if}

      {#if success}
        <div class="relative rounded border border-green-400 bg-green-50 px-4 py-3 text-green-700" role="alert">
          <span class="block sm:inline">{success}</span>
          <div class="mt-2">
            <a href="/auth/signin" class="font-medium text-green-700 underline hover:text-green-600"> Go to sign in </a>
          </div>
        </div>
      {/if}

      <div class="space-y-4">
        <div>
          <label for="password" class="sr-only">New password</label>
          <input
            id="password"
            bind:value={formData.password}
            name="password"
            type="password"
            autocomplete="new-password"
            required
            minlength="8"
            class="relative block w-full appearance-none rounded border border-gray-300 bg-white px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
            placeholder="New password (min. 8 characters)"
          />
        </div>

        <div>
          <label for="confirm-password" class="sr-only">Confirm password</label>
          <input
            id="confirm-password"
            bind:value={formData.confirmPassword}
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
        <Button type="submit" variant="primary" class="w-full" loading={isLoading} disabled={isLoading || !token}>
          {#if isLoading}
            <span>Resetting password...</span>
          {:else if !token}
            <span>Invalid reset link</span>
          {:else}
            <span>Reset password</span>
          {/if}
        </Button>
      </div>
    </form>
  </div>
</div>
