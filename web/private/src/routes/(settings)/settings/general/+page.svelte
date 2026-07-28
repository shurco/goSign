<script lang="ts">
  import { t, getLocale, setLocale, SUPPORTED_LOCALES, type Locale } from "@/i18n/index.svelte";
  import FormControl from "@/components/ui/FormControl.svelte";
  import { apiPut } from "@/services/api";

  const locales = SUPPORTED_LOCALES;

  async function changeLocale(newLocale: string) {
    if (!newLocale || newLocale === getLocale()) {
      return;
    }

    // Update locale immediately for instant UI change
    setLocale(newLocale as Locale);
    localStorage.setItem("locale", newLocale);
    document.documentElement.setAttribute("lang", newLocale);

    // Update user preference on backend (non-blocking)
    try {
      await apiPut("/settings/i18n/user/locale", { locale: newLocale });
    } catch (error) {
      // Silently fail - locale is already updated in frontend
      // User can still use the app even if backend update fails
      console.warn("Failed to update user locale on server:", error);
    }
  }
</script>

<div class="space-y-6">
  <FormControl label={t("settings.language")}>
    <div class="flex flex-wrap gap-2">
      {#each Object.entries(locales) as [code, name] (code)}
        <button
          class={[
            "cursor-pointer rounded-md border px-4 py-2 text-sm font-medium transition-colors",
            getLocale() === code
              ? "border-blue-500 bg-blue-50 text-blue-700"
              : "border-gray-300 bg-white text-gray-700 hover:bg-gray-50"
          ].join(" ")}
          onclick={() => changeLocale(code)}
        >
          {name}
        </button>
      {/each}
    </div>
  </FormControl>
</div>
