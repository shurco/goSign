<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <div class="text-sm font-semibold">{{ $t("settings.smsConfiguration") }}</div>
        <div class="text-xs text-[--color-base-content]/70">{{ $t("settings.smsDescription") }}</div>
      </div>
      <input v-model="sms.twilio_enabled" type="checkbox" class="toggle toggle-sm" />
    </div>

    <FormControl :label="$t('settings.twilioAccountSid')">
      <Input v-model="sms.twilio_account_sid" type="text" placeholder="ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" />
    </FormControl>

    <FormControl :label="$t('settings.twilioAuthToken')">
      <Input v-model="sms.twilio_auth_token" type="password" placeholder="••••••••••••••••••••••••••••••••" />
      <div v-if="sms.twilio_auth_token_set" class="mt-1 text-xs text-[--color-base-content]/70">
        Token is already set (leave empty to keep unchanged).
      </div>
    </FormControl>

    <FormControl :label="$t('settings.twilioFromNumber')">
      <Input v-model="sms.twilio_from_number" type="tel" placeholder="+15551234567" />
    </FormControl>

    <div class="divider"></div>

    <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
      <FormControl :label="$t('settings.testSmsTo')">
        <Input v-model="test.to_phone" type="tel" placeholder="+15551234567" />
      </FormControl>
      <FormControl :label="$t('settings.testSmsMessage')">
        <Input v-model="test.message" type="text" :placeholder="$t('settings.testSmsMessagePlaceholder')" />
      </FormControl>
    </div>

    <div class="flex justify-end gap-3 pt-4">
      <Button variant="ghost" @click="sendTest">{{ $t("settings.testSms") }}</Button>
      <Button variant="primary" @click="save">{{ $t("common.save") }}</Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import FormControl from "@/components/ui/FormControl.vue";
import Input from "@/components/ui/Input.vue";
import Button from "@/components/ui/Button.vue";
import { fetchWithAuth } from "@/utils/auth";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const sms = ref({
  twilio_enabled: false,
  twilio_account_sid: "",
  twilio_from_number: "",
  twilio_auth_token: "", // write-only
  twilio_auth_token_set: false
});

const test = ref({
  to_phone: "",
  message: ""
});

onMounted(async () => {
  await load();
});

async function load(): Promise<void> {
  try {
    const res = await fetchWithAuth("/api/v1/settings");
    if (!res.ok) return;
    const data = await res.json();
    const settings = data.data || data;
    if (settings.sms) {
      sms.value.twilio_enabled = !!settings.sms.twilio_enabled;
      sms.value.twilio_account_sid = String(settings.sms.twilio_account_sid || "");
      sms.value.twilio_from_number = String(settings.sms.twilio_from_number || "");
      sms.value.twilio_auth_token_set = !!settings.sms.twilio_auth_token_set;
      sms.value.twilio_auth_token = "";
    }
  } catch (e) {
    console.error("Failed to load SMS settings:", e);
  }
}

async function save(): Promise<void> {
  try {
    const res = await fetchWithAuth("/api/v1/settings/sms", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        twilio_enabled: sms.value.twilio_enabled,
        twilio_account_sid: sms.value.twilio_account_sid,
        twilio_from_number: sms.value.twilio_from_number,
        twilio_auth_token: sms.value.twilio_auth_token
      })
    });
    if (!res.ok) {
      alert("Failed to save SMS settings");
      return;
    }
    sms.value.twilio_auth_token = "";
    await load();
    alert("SMS settings saved successfully");
  } catch (e) {
    console.error("Failed to save SMS settings:", e);
    alert("Failed to save SMS settings");
  }
}

async function sendTest(): Promise<void> {
  const to = test.value.to_phone?.trim();
  if (!to) {
    alert(t("settings.testSmsToRequired"));
    return;
  }
  try {
    const res = await fetchWithAuth("/api/v1/settings/sms/test", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        to_phone: to,
        message: test.value.message
      })
    });
    if (res.ok) {
      alert(t("settings.testSmsSent"));
      return;
    }
    const data = await res.json().catch(() => ({}));
    alert(data?.data?.error || data?.error || t("settings.testSmsFailed"));
  } catch (e) {
    console.error("Failed to send test SMS:", e);
    alert(t("settings.testSmsFailed"));
  }
}
</script>

