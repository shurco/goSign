<template>
  <div class="space-y-4">
    <FormControl :label="$t('settings.language')">
      <div class="flex flex-wrap gap-2">
        <button
          v-for="(name, code) in locales"
          :key="code"
          :class="[
            'rounded-md border px-4 py-2 text-sm font-medium transition-colors cursor-pointer',
            currentLocale === code
              ? 'border-blue-500 bg-blue-50 text-blue-700'
              : 'border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
          ]"
          @click="changeLocale(code)"
        >
          {{ name }}
        </button>
      </div>
    </FormControl>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { SUPPORTED_LOCALES } from '@/i18n';
import { apiPut } from '@/services/api';
import FormControl from "@/components/ui/FormControl.vue";

const { locale } = useI18n();
const locales = SUPPORTED_LOCALES;

const currentLocale = computed(() => locale.value as string);

async function changeLocale(newLocale: string) {
  if (!newLocale || newLocale === locale.value) {
    return;
  }
  
  // Update locale immediately for instant UI change
  locale.value = newLocale;
  localStorage.setItem('locale', newLocale);
  document.documentElement.setAttribute('lang', newLocale);
  
  // Update user preference on backend (non-blocking)
  try {
    await apiPut('/api/v1/i18n/user/locale', { locale: newLocale });
  } catch (error) {
    // Silently fail - locale is already updated in frontend
    // User can still use the app even if backend update fails
    console.warn('Failed to update user locale on server:', error);
  }
}
</script>
