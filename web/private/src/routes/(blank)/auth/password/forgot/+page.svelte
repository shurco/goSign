<script lang="ts">
  import Button from "@/components/ui/Button.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { apiUrl } from "@/services/api";

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
      const response = await fetch(apiUrl("/auth/password/forgot"), {
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

<div class="auth-page">
  <div class="auth-card">
    <div class="auth-header">
      <span class="auth-mark">
        <SvgIcon name="logo" width="26" height="26" />
      </span>
      <h1 class="auth-title">Reset your password</h1>
      <p class="auth-subtitle">Enter your email address and we'll send you a link to reset your password</p>
    </div>

    {#if error}
      <div class="auth-error" role="alert">{error}</div>
    {/if}

    {#if success}
      <div class="auth-success" role="alert">{success}</div>
    {/if}

    <form class="auth-form" onsubmit={handleSubmit}>
      <div class="form-group">
        <label for="email-address">Email address</label>
        <input
          id="email-address"
          bind:value={email}
          name="email"
          type="email"
          autocomplete="email"
          required
          placeholder="Email address"
        />
      </div>

      <Button type="submit" variant="primary" size="lg" class="w-full" loading={isLoading} disabled={isLoading}>
        {#if isLoading}
          <span>Sending...</span>
        {:else}
          <span>Send reset link</span>
        {/if}
      </Button>

      <div class="auth-footer">
        <a href="/auth/signin">Back to sign in</a>
      </div>
    </form>
  </div>
</div>
