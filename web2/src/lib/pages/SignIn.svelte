<script lang="ts">
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import { t } from "@/i18n/index.svelte";
  import Button from "@/components/ui/Button.svelte";

  let formData = $state({
    email: "",
    password: "",
    code: ""
  });

  let isLoading = $state(false);
  let error = $state("");
  let requires2FA = $state(false);

  const handleSubmit = async (e: SubmitEvent) => {
    e.preventDefault();
    error = "";
    isLoading = true;

    try {
      const response = await fetch("/auth/signin", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          email: formData.email,
          password: formData.password,
          code: formData.code || undefined
        })
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || "Failed to sign in");
      }

      // Check if 2FA is required
      if (data.data?.requires_2fa) {
        requires2FA = true;
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
      const redirectPath = page.url.searchParams.get("redirect") || "/dashboard";
      goto(redirectPath);
    } catch (err) {
      error = err instanceof Error ? err.message : "An error occurred";
    } finally {
      isLoading = false;
    }
  };

  const handleGoogleSignIn = () => {
    window.location.href = "/auth/oauth/google";
  };

  const handleGitHubSignIn = () => {
    window.location.href = "/auth/oauth/github";
  };
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
  <div class="w-full max-w-md space-y-8">
    <div>
      <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">{t("auth.signin")}</h2>
      <p class="mt-2 text-center text-sm text-gray-600">
        {t("common.or")}
        <a href="/auth/signup" class="font-medium text-indigo-600 hover:text-indigo-500">
          {t("auth.createAccount")}
        </a>
      </p>
    </div>

    <form class="mt-8 space-y-6" onsubmit={handleSubmit}>
      {#if error}
        <div class="relative rounded border border-red-400 bg-red-50 px-4 py-3 text-red-700" role="alert">
          <span class="block sm:inline">{error}</span>
        </div>
      {/if}

      <div class="space-y-4">
        <div>
          <label for="email-address" class="sr-only">Email address</label>
          <input
            id="email-address"
            bind:value={formData.email}
            name="email"
            type="email"
            autocomplete="email"
            required
            disabled={requires2FA}
            class="relative block w-full appearance-none rounded border border-gray-300 bg-white px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none disabled:bg-gray-100 disabled:opacity-50 sm:text-sm"
            placeholder={t("auth.email")}
          />
        </div>

        <div>
          <label for="password" class="sr-only">Password</label>
          <input
            id="password"
            bind:value={formData.password}
            name="password"
            type="password"
            autocomplete="current-password"
            required
            disabled={requires2FA}
            class="relative block w-full appearance-none rounded border border-gray-300 bg-white px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none disabled:bg-gray-100 disabled:opacity-50 sm:text-sm"
            placeholder="Password"
          />
        </div>

        {#if requires2FA}
          <div>
            <label for="code" class="sr-only">2FA Code</label>
            <input
              id="code"
              bind:value={formData.code}
              name="code"
              type="text"
              inputmode="numeric"
              pattern="[0-9]*"
              maxlength="6"
              required
              class="relative block w-full appearance-none rounded border border-gray-300 bg-white px-3 py-2 text-center text-2xl tracking-widest text-gray-900 placeholder-gray-500 focus:z-10 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none sm:text-sm"
              placeholder="000000"
            />
            <p class="mt-2 text-center text-sm text-gray-500">Enter the 6-digit code from your authenticator app</p>
          </div>
        {/if}
      </div>

      <div class="flex items-center justify-between">
        <div class="text-sm">
          <a href="/auth/password/forgot" class="font-medium text-indigo-600 hover:text-indigo-500">
            {t("auth.forgotPassword")}
          </a>
        </div>
      </div>

      <div>
        <Button type="submit" variant="primary" class="w-full" loading={isLoading} disabled={isLoading}>
          {t("auth.signin")}
        </Button>
      </div>

      {#if !requires2FA}
        <div class="flex items-center justify-center">
          <div class="text-sm">
            <span class="text-gray-600">Or sign in with</span>
          </div>
        </div>
      {/if}

      {#if !requires2FA}
        <div class="grid grid-cols-2 gap-3">
          <Button variant="ghost" class="w-full" onclick={handleGoogleSignIn}>Google</Button>
          <Button variant="ghost" class="w-full" onclick={handleGitHubSignIn}>GitHub</Button>
        </div>
      {/if}
    </form>
  </div>
</div>
