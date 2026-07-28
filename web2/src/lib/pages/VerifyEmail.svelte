<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";

  let isLoading = $state(true);
  let error = $state("");
  let success = $state(false);

  onMount(async () => {
    const token = page.url.searchParams.get("token") || "";

    if (!token) {
      error = "Invalid or missing verification token";
      isLoading = false;
      return;
    }

    try {
      const response = await fetch(`/auth/verify-email?token=${encodeURIComponent(token)}`, {
        method: "GET"
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || "Failed to verify email");
      }

      success = true;
    } catch (err) {
      error = err instanceof Error ? err.message : "An error occurred during verification";
    } finally {
      isLoading = false;
    }
  });
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
  <div class="w-full max-w-md space-y-8">
    <div>
      <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">Email Verification</h2>
    </div>

    <div class="mt-8 space-y-6">
      {#if isLoading}
        <div class="text-center">
          <div class="inline-block h-12 w-12 animate-spin rounded-full border-b-2 border-indigo-600"></div>
          <p class="mt-4 text-gray-600">Verifying your email...</p>
        </div>
      {:else if error}
        <div class="relative rounded border border-red-400 bg-red-50 px-4 py-3 text-red-700" role="alert">
          <p class="font-bold">Verification failed</p>
          <p class="mt-2">{error}</p>
          <div class="mt-4">
            <a href="/auth/signin" class="font-medium text-red-700 underline hover:text-red-600"> Go to sign in </a>
          </div>
        </div>
      {:else if success}
        <div class="relative rounded border border-green-400 bg-green-50 px-4 py-3 text-green-700" role="alert">
          <p class="font-bold">Email verified successfully!</p>
          <p class="mt-2">Your email has been verified. You can now sign in to your account.</p>
          <div class="mt-4">
            <a
              href="/auth/signin"
              class="inline-flex items-center justify-center rounded-md border border-blue-500 bg-blue-50 px-4 py-2 text-sm font-medium text-blue-700 transition-colors hover:bg-blue-100 focus:outline-none"
            >
              Sign in now
            </a>
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>
