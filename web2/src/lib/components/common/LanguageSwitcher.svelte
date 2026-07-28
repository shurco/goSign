<script lang="ts">
  import Select from "@/components/ui/Select.svelte";
  import { getLocale, setLocale, SUPPORTED_LOCALES, type Locale } from "@/i18n/index.svelte";
  import { apiPut } from "@/services/api";

  const locales = SUPPORTED_LOCALES;

  async function changeLocale(newLocale: string): Promise<void> {
    if (!newLocale || newLocale === getLocale()) {
      return;
    }

    // Update locale immediately for instant UI change
    setLocale(newLocale as Locale);
    localStorage.setItem("locale", newLocale);
    document.documentElement.setAttribute("lang", newLocale);

    // Update user preference on backend (non-blocking)
    try {
      await apiPut("/api/v1/i18n/user/locale", { locale: newLocale });
    } catch (error) {
      // Silently fail - locale is already updated in frontend
      // User can still use the app even if backend update fails
      console.warn("Failed to update user locale on server:", error);
    }
  }

  function handleChange(event: Event): void {
    const value = (event.currentTarget as HTMLSelectElement).value;
    if (value && value !== getLocale()) {
      changeLocale(value);
    }
  }
</script>

<Select class="language-switcher" value={getLocale()} onchange={handleChange}>
  {#each Object.entries(locales) as [code, name] (code)}
    <option value={code}>{name}</option>
  {/each}
</Select>

<style>
  :global(.language-switcher) {
    min-width: 120px;
  }
</style>
