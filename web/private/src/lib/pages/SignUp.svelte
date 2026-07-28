<script lang="ts">
  import Button from "@/components/ui/Button.svelte";
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

<div class="flex min-h-screen items-center justify-center bg-gray-50 px-4 py-12 sm:px-6 lg:px-8">
  <div class="w-full max-w-md space-y-8">
    <div>
      <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">Create your account</h2>
      <p class="mt-2 text-center text-sm text-gray-600">
        Or
        <a href="/auth/signin" class="font-medium text-indigo-600 hover:text-indigo-500"> sign in to your account </a>
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

      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label for="first-name" class="sr-only">First name</label>
            <input
              id="first-name"
              bind:value={formData.firstName}
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
              bind:value={formData.lastName}
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
            bind:value={formData.email}
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
            bind:value={formData.password}
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
        <Button type="submit" variant="primary" class="w-full" loading={isLoading} disabled={isLoading}>
          {#if isLoading}
            <span>Creating account...</span>
          {:else}
            <span>Sign up</span>
          {/if}
        </Button>
      </div>

      <div class="flex items-center justify-center">
        <div class="text-sm">
          <span class="text-gray-600">Or sign up with</span>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <Button variant="ghost" class="w-full" onclick={handleGoogleSignIn}>Google</Button>
        <Button variant="ghost" class="w-full" onclick={handleGitHubSignIn}>GitHub</Button>
      </div>
    </form>
  </div>
</div>
