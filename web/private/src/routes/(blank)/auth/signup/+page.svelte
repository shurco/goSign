<script lang="ts">
  import Button from "@/components/ui/Button.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { apiUrl } from "@/services/api";
  // import { goto } from "$app/navigation";

  let formData = $state({
    firstName: "",
    lastName: "",
    email: "",
    password: "",
    confirmPassword: ""
  });

  let isLoading = $state(false);
  let error = $state("");
  let success = $state("");

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
      const response = await fetch(apiUrl("/auth/signup"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          email: formData.email,
          password: formData.password,
          first_name: formData.firstName,
          last_name: formData.lastName
        })
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || "Failed to sign up");
      }

      success = "Registration successful! Please check your email to verify your account.";

      // Clear form
      formData = {
        firstName: "",
        lastName: "",
        email: "",
        password: "",
        confirmPassword: ""
      };
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
      <h1 class="auth-title">Create your account</h1>
      <p class="auth-subtitle">
        Or
        <a href="/auth/signin">sign in to your account</a>
      </p>
    </div>

    {#if error}
      <div class="auth-error" role="alert">{error}</div>
    {/if}

    {#if success}
      <div class="auth-success" role="alert">{success}</div>
    {/if}

    <form class="auth-form" onsubmit={handleSubmit}>
      <div class="grid grid-cols-2 gap-3">
        <div class="form-group">
          <label for="first-name">First name</label>
          <input
            id="first-name"
            bind:value={formData.firstName}
            name="first-name"
            type="text"
            autocomplete="given-name"
            required
            placeholder="First name"
          />
        </div>
        <div class="form-group">
          <label for="last-name">Last name</label>
          <input
            id="last-name"
            bind:value={formData.lastName}
            name="last-name"
            type="text"
            autocomplete="family-name"
            required
            placeholder="Last name"
          />
        </div>
      </div>

      <div class="form-group">
        <label for="email-address">Email address</label>
        <input
          id="email-address"
          bind:value={formData.email}
          name="email"
          type="email"
          autocomplete="email"
          required
          placeholder="Email address"
        />
      </div>

      <div class="form-group">
        <label for="password">Password</label>
        <input
          id="password"
          bind:value={formData.password}
          name="password"
          type="password"
          autocomplete="new-password"
          required
          minlength="8"
          placeholder="Password (min. 8 characters)"
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

      <Button type="submit" variant="primary" size="lg" class="w-full" loading={isLoading} disabled={isLoading}>
        {#if isLoading}
          <span>Creating account...</span>
        {:else}
          <span>Sign up</span>
        {/if}
      </Button>

      <div class="auth-divider">Or sign up with</div>

      <div class="grid grid-cols-2 gap-3">
        <Button variant="ghost" class="w-full" onclick={handleGoogleSignIn}>Google</Button>
        <Button variant="ghost" class="w-full" onclick={handleGitHubSignIn}>GitHub</Button>
      </div>
    </form>
  </div>
</div>

<style>
  .auth-divider {
    text-align: center;
    font-size: var(--font-size-12);
    color: var(--base-txt-muted);
  }
</style>
