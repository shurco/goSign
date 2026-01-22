<template>
  <div :class="['signing-page', `theme-${theme}`]">
    <!-- Logo -->
    <div v-if="branding.logo_url" class="company-logo mb-6 text-center">
      <img :src="branding.logo_url" :alt="branding.company_name" class="max-h-16" />
    </div>
    
    <!-- Title with company name -->
    <h1 class="mb-4 text-center text-2xl font-bold">
      {{ branding.company_name || 'Document Signing' }}
    </h1>
    
    <slot />
    
    <!-- Footer -->
    <footer class="mt-8 border-t pt-4 text-center text-sm text-gray-500">
      <div v-if="branding.terms_url || branding.privacy_url" class="legal-links mb-2">
        <a
          v-if="branding.terms_url"
          :href="branding.terms_url"
          target="_blank"
          class="mx-2 hover:underline"
        >
          Terms of Service
        </a>
        <a
          v-if="branding.privacy_url"
          :href="branding.privacy_url"
          target="_blank"
          class="mx-2 hover:underline"
        >
          Privacy Policy
        </a>
      </div>
      
      <div v-if="branding.show_powered_by" class="powered-by">
        Powered by goSign
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { BrandingSettings } from '@/models/account'

interface Props {
  branding: BrandingSettings
}

const props = defineProps<Props>()

const theme = computed(() => props.branding.signing_page_theme || 'default')
</script>

<style scoped>
/* Default theme */
.theme-default {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem;
}

/* Minimal theme */
.theme-minimal {
  max-width: 600px;
  margin: 0 auto;
  padding: 1rem;
  background: white;
  box-shadow: none;
}

/* Corporate theme */
.theme-corporate {
  max-width: 1000px;
  margin: 0 auto;
  padding: 3rem;
  background: var(--color-background, #ffffff);
  border-top: 4px solid var(--color-primary, #4F46E5);
}
</style>
