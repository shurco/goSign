<template>
  <div class="settings-page">
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold">{{ pageTitle }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ pageDescription }}</p>
      </div>
    </div>

    <!-- Content Area -->
    <RouterView />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute, RouterView } from "vue-router";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const route = useRoute();

// Map route names to title and description translation keys
const pageInfo = computed(() => {
  const routeName = route.name as string;
  
  const infoMap: Record<string, { title: string; description: string }> = {
    "admin-settings-smtp": {
      title: t('settings.smtpConfiguration'),
      description: t('settings.smtpDescription')
    },
    "admin-settings-sms": {
      title: t('settings.smsConfiguration'),
      description: t('settings.smsDescription')
    },
    "admin-settings-storage": {
      title: t('settings.storageConfiguration'),
      description: t('settings.storageDescription')
    },
    "admin-settings-geolocation": {
      title: t('settings.geolocation'),
      description: t('settings.geolocationSectionDescription')
    },
    "admin-settings-email-templates": {
      title: t('settings.emailTemplates'),
      description: t('settings.emailTemplatesDescription')
    }
  };

  return infoMap[routeName] || {
    title: t('settings.adminSettings'),
    description: t('settings.adminSettingsDescription')
  };
});

const pageTitle = computed(() => pageInfo.value.title);
const pageDescription = computed(() => pageInfo.value.description);
</script>

<style scoped>
.settings-page {
  @apply min-h-full;
}
</style>
