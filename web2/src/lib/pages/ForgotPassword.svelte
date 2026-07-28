<script lang="ts">
  import Button from "@/components/ui/Button.svelte";

  let email = $state("");
  let isLoading = $state(false);
  let error = $state("");
  let success = $state("");

  const handleSubmit = async (e: SubmitEvent) => {
    e.preventDefault();
    error = "";
    success = "";
    isLoading = true;

    try {
      const response = await fetch("/auth/password/forgot", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          email
        })
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || "Failed to send reset link");
      }

      success = "If the email exists, a password reset link has been sent to your email";
      email = "";
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
      <p class="mt-2 text-center text-sm text-gray-600">
        Enter your email address and we'll send you a link to reset your password
      </p>
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
        </div>
      {/if}

      <div>
        <label for="email-address" class="sr-only">Email address</label>
        <input
          id="email-address"
          bind:value={email}
          name="email"
          type="email"
          autocomplete="email"
          required
          class="relative block w-full appearance-none rounded border border-gray-300 bg-white px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
          placeholder="Email address"
        />
      </div>

      <div>
        <Button type="submit" variant="primary" class="w-full" loading={isLoading} disabled={isLoading}>
          {#if isLoading}
            <span>Sending...</span>
          {:else}
            <span>Send reset link</span>
          {/if}
        </Button>
      </div>

      <div class="text-center">
        <a href="/auth/signin" class="font-medium text-indigo-600 hover:text-indigo-500"> Back to sign in </a>
      </div>
    </form>
  </div>
</div>
