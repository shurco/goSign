<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { apiUrl } from "@/services/api";
  import { page } from "$app/state";
  import Button from "@/components/ui/Button.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

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
      const response = await fetch(apiUrl(`/auth/verify-email?token=${encodeURIComponent(token)}`), {
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

<div class="auth-page">
  <div class="auth-card">
    <div class="auth-header">
      <span class="auth-mark">
        <SvgIcon name="logo" width="26" height="26" />
      </span>
      <h1 class="auth-title">Email Verification</h1>
    </div>

    {#if isLoading}
      <div class="verify-loading">
        <div class="verify-spinner"></div>
        <p>Verifying your email...</p>
      </div>
    {:else if error}
      <div class="auth-error" role="alert">
        <p class="verify-strong">Verification failed</p>
        <p class="mt-2">{error}</p>
        <div class="mt-4">
          <a href="/auth/signin">Go to sign in</a>
        </div>
      </div>
    {:else if success}
      <div class="auth-success" role="alert">
        <p class="verify-strong">Email verified successfully!</p>
        <p class="mt-2">Your email has been verified. You can now sign in to your account.</p>
      </div>
      <Button variant="primary" size="lg" class="w-full" onclick={() => goto("/auth/signin")}>Sign in now</Button>
    {/if}
  </div>
</div>

<style>
  .verify-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-12);
    color: var(--base-txt-muted);
    font-size: var(--font-size-13);
    padding: var(--space-12) 0;
  }
  .verify-spinner {
    width: 40px;
    height: 40px;
    border-radius: var(--radius-full);
    border: 3px solid var(--base-hlt-selected);
    border-top-color: var(--base-hlt-invert);
    animation: verify-spin 0.8s linear infinite;
  }
  @keyframes verify-spin {
    to {
      transform: rotate(360deg);
    }
  }
  .verify-strong {
    font-weight: var(--font-weight-bold);
    margin: 0;
  }
</style>
