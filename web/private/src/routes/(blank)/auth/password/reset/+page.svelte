<script lang="ts">
  import { onMount } from "svelte";
  import { apiUrl } from "@/services/api";
  import { page } from "$app/state";
  import Button from "@/components/ui/Button.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

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
      const response = await fetch(apiUrl("/auth/password/reset"), {
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

<div class="auth-page">
  <div class="auth-card">
    <div class="auth-header">
      <span class="auth-mark">
        <SvgIcon name="logo" width="26" height="26" />
      </span>
      <h1 class="auth-title">Reset your password</h1>
      <p class="auth-subtitle">Enter your new password</p>
    </div>

    {#if error}
      <div class="auth-error" role="alert">{error}</div>
    {/if}

    {#if success}
      <div class="auth-success" role="alert">
        {success}
        <div class="mt-2">
          <a href="/auth/signin">Go to sign in</a>
        </div>
      </div>
    {/if}

    <form class="auth-form" onsubmit={handleSubmit}>
      <div class="form-group">
        <label for="password">New password</label>
        <input
          id="password"
          bind:value={formData.password}
          name="password"
          type="password"
          autocomplete="new-password"
          required
          minlength="8"
          placeholder="New password (min. 8 characters)"
        />
      </div>

      <div class="form-group">
        <label for="confirm-password">Confirm password</label>
        <input
          id="confirm-password"
          bind:value={formData.confirmPassword}
          name="confirm-password"
          type="password"
          autocomplete="new-password"
          required
          placeholder="Confirm password"
        />
      </div>

      <Button
        type="submit"
        variant="primary"
        size="lg"
        class="w-full"
        loading={isLoading}
        disabled={isLoading || !token}
      >
        {#if isLoading}
          <span>Resetting password...</span>
        {:else if !token}
          <span>Invalid reset link</span>
        {:else}
          <span>Reset password</span>
        {/if}
      </Button>
    </form>
  </div>
</div>
