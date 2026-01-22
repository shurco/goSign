<template>
  <Select v-model="currentLocale" class="language-switcher">
    <option v-for="(name, code) in locales" :key="code" :value="code">
      {{ name }}
    </option>
  </Select>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { SUPPORTED_LOCALES } from '@/i18n'
import { apiPut } from '@/services/api'
import Select from '@/components/ui/Select.vue'

const { locale } = useI18n()
const locales = SUPPORTED_LOCALES

const currentLocale = computed({
  get: () => locale.value as string,
  set: (value: string) => {
    if (value && value !== locale.value) {
      changeLocale(value)
    }
  }
})

async function changeLocale(newLocale: string) {
  if (!newLocale || newLocale === locale.value) {
    return
  }
  
  // Update locale immediately for instant UI change
  locale.value = newLocale
  localStorage.setItem('locale', newLocale)
  document.documentElement.setAttribute('lang', newLocale)
  
  // Update user preference on backend (non-blocking)
  try {
    await apiPut('/api/v1/i18n/user/locale', { locale: newLocale })
  } catch (error) {
    // Silently fail - locale is already updated in frontend
    // User can still use the app even if backend update fails
    console.warn('Failed to update user locale on server:', error)
  }
}
</script>

<style scoped>
.language-switcher {
  min-width: 120px;
}
</style>
