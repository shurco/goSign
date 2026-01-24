<template>
  <div class="settings-page">
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between">
        <div>
        <h1 class="text-3xl font-bold">{{ pageTitle }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ pageDescription }}</p>
      </div>
      <div v-if="showActionButton" class="flex items-center gap-3">
        <Button v-if="route.name === 'settings-webhooks'" variant="primary" @click="triggerWebhookModal">
          <SvgIcon name="plus" class="mr-2 h-5 w-5" />
          {{ $t('webhooks.addWebhook') }}
        </Button>
        <Button v-else-if="route.name === 'settings-api-keys'" variant="primary" @click="triggerApiKeyModal">
          <SvgIcon name="plus" class="mr-2 h-5 w-5" />
          {{ $t('apikeys.createApiKey') }}
          </Button>
      </div>
        </div>

    <!-- Content Area -->
    <RouterView />
  </div>
</template>

<script setup lang="ts">
import { computed, provide, ref } from "vue";
import { useRoute, RouterView } from "vue-router";
import { useI18n } from "vue-i18n";
import Button from "@/components/ui/Button.vue";
import SvgIcon from "@/components/SvgIcon.vue";

const { t } = useI18n();
const route = useRoute();

// Map route names to title and description translation keys
const pageInfo = computed(() => {
  const routeName = route.name as string;
  
  const infoMap: Record<string, { title: string; description: string }> = {
    "settings-general": {
      title: t('settings.generalSettings'),
      description: t('settings.generalDescription')
    },
    "settings-geolocation": {
      title: t('settings.geolocation'),
      description: t('settings.geolocationSectionDescription')
    },
    "settings-smtp": {
      title: t('settings.smtpConfiguration'),
      description: t('settings.smtpDescription')
    },
    "settings-sms": {
      title: t('settings.smsConfiguration'),
      description: t('settings.smsDescription')
    },
    "settings-storage": {
      title: t('settings.storageConfiguration'),
      description: t('settings.storageDescription')
    },
    "settings-webhooks": {
      title: t('webhooks.title'),
      description: t('settings.description')
    },
    "settings-api-keys": {
      title: t('apikeys.title'),
      description: t('settings.description')
    },
    "settings-branding": {
      title: t('branding.title'),
      description: t('branding.description')
    },
    "settings-email-templates": {
      title: t('settings.emailTemplates'),
      description: t('settings.emailTemplatesDescription')
    }
  };

  return infoMap[routeName] || {
    title: t('settings.title'),
    description: t('settings.description')
  };
});

const pageTitle = computed(() => pageInfo.value.title);
const pageDescription = computed(() => pageInfo.value.description);

// Show action button for webhooks and api-keys pages
const showActionButton = computed(() => {
  return route.name === 'settings-webhooks' || route.name === 'settings-api-keys';
});

// Provide functions to child components
const webhookModalTrigger = ref<(() => void) | null>(null);
const apiKeyModalTrigger = ref<(() => void) | null>(null);

provide('webhookModalTrigger', webhookModalTrigger);
provide('apiKeyModalTrigger', apiKeyModalTrigger);

function triggerWebhookModal(): void {
  if (webhookModalTrigger.value) {
    webhookModalTrigger.value();
  }
}

function triggerApiKeyModal(): void {
  if (apiKeyModalTrigger.value) {
    apiKeyModalTrigger.value();
  }
}
</script>

<style scoped>
.settings-page {
  @apply min-h-full;
}
</style>
