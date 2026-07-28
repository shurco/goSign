<script lang="ts">
  import { page } from "$app/state";
  import { apiUrl } from "@/services/api";
  import { goto } from "$app/navigation";
  import { t } from "@/i18n/index.svelte";
  import Button from "@/components/ui/Button.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

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
      const response = await fetch(apiUrl("/auth/signin"), {
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
    window.location.href = apiUrl("/auth/oauth/google");
  };

  const handleGitHubSignIn = () => {
    window.location.href = apiUrl("/auth/oauth/github");
  };
</script>

<div class="auth-page">
  <div class="auth-card">
    <div class="auth-header">
      <span class="auth-mark">
        <SvgIcon name="logo" width="26" height="26" />
      </span>
      <h1 class="auth-title">goSign</h1>
      <p class="auth-subtitle">
        {t("auth.signin")}
        {t("common.or")}
        <a href="/auth/signup">{t("auth.createAccount")}</a>
      </p>
    </div>

    {#if error}
      <div class="auth-error" role="alert">{error}</div>
    {/if}

    <form class="auth-form" onsubmit={handleSubmit}>
      <div class="form-group">
        <label for="email-address">{t("auth.email")}</label>
        <input
          id="email-address"
          bind:value={formData.email}
          name="email"
          type="email"
          autocomplete="email"
          required
          disabled={requires2FA}
          placeholder={t("auth.email")}
        />
      </div>

      <div class="form-group">
        <label for="password">Password</label>
        <input
          id="password"
          bind:value={formData.password}
          name="password"
          type="password"
          autocomplete="current-password"
          required
          disabled={requires2FA}
          placeholder="Password"
        />
      </div>

      {#if requires2FA}
        <div class="form-group">
          <label for="code">2FA Code</label>
          <input
            id="code"
            bind:value={formData.code}
            name="code"
            type="text"
            inputmode="numeric"
            pattern="[0-9]*"
            maxlength="6"
            required
            class="auth-code-input"
            placeholder="000000"
          />
          <p class="auth-hint">Enter the 6-digit code from your authenticator app</p>
        </div>
      {/if}

      <div class="auth-links">
        <a href="/auth/password/forgot">{t("auth.forgotPassword")}</a>
      </div>

      <Button type="submit" variant="primary" size="lg" class="w-full" loading={isLoading} disabled={isLoading}>
        {t("auth.signin")}
      </Button>

      {#if !requires2FA}
        <div class="auth-divider">Or sign in with</div>

        <div class="grid grid-cols-2 gap-3">
          <Button variant="ghost" class="w-full" onclick={handleGoogleSignIn}>Google</Button>
          <Button variant="ghost" class="w-full" onclick={handleGitHubSignIn}>GitHub</Button>
        </div>
      {/if}
    </form>
  </div>
</div>

<style>
  .auth-code-input {
    text-align: center;
    font-size: var(--font-size-22);
    letter-spacing: 0.3em;
  }
  .auth-hint {
    margin: var(--space-4) 0 0;
    text-align: center;
    font-size: var(--font-size-12);
    color: var(--base-txt-muted);
  }
  .auth-links {
    display: flex;
    justify-content: flex-end;
    font-size: var(--font-size-13);
  }
  .auth-divider {
    text-align: center;
    font-size: var(--font-size-12);
    color: var(--base-txt-muted);
  }
</style>
