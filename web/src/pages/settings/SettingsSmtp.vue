<template>
  <div class="space-y-4">
    <FormControl :label="$t('settings.smtpHost')">
      <Input v-model="smtpSettings.host" type="text" placeholder="smtp.gmail.com" />
    </FormControl>

    <div class="grid grid-cols-2 gap-4">
      <FormControl :label="$t('settings.port')">
        <Input v-model="smtpSettings.port" type="number" placeholder="587" />
      </FormControl>

      <FormControl :label="$t('settings.encryption')">
        <Select v-model="smtpSettings.encryption">
          <option value="tls">{{ $t('settings.tls') }}</option>
          <option value="ssl">{{ $t('settings.ssl') }}</option>
          <option value="none">{{ $t('settings.none') }}</option>
        </Select>
      </FormControl>
    </div>

    <FormControl :label="$t('settings.username')">
      <Input v-model="smtpSettings.username" type="text" />
    </FormControl>

    <FormControl :label="$t('settings.password')">
      <Input v-model="smtpSettings.password" type="password" />
    </FormControl>

    <FormControl :label="$t('settings.fromEmail')">
      <Input v-model="smtpSettings.from_email" type="email" placeholder="noreply@example.com" />
    </FormControl>

    <div class="flex justify-end gap-3 pt-4">
      <Button variant="ghost" @click="testSmtp">{{ $t('settings.testConnection') }}</Button>
      <Button variant="primary" @click="saveSmtp">{{ $t('common.save') }}</Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import FormControl from "@/components/ui/FormControl.vue";
import Input from "@/components/ui/Input.vue";
import Select from "@/components/ui/Select.vue";
import Button from "@/components/ui/Button.vue";
import { fetchWithAuth } from "@/utils/auth";

const { t } = useI18n();

const smtpSettings = ref({
  host: "",
  port: 587,
  encryption: "tls",
  username: "",
  password: "",
  from_email: ""
});

onMounted(async () => {
  await loadSettings();
});

async function loadSettings(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/settings");
    if (response.ok) {
      const data = await response.json();
      const settings = data.data || data;
      if (settings.email) {
        smtpSettings.value = {
          host: settings.email.smtp_host || "",
          port: settings.email.smtp_port || 587,
          encryption: "tls",
          username: settings.email.smtp_user || "",
          password: "",
          from_email: settings.email.from_email || ""
        };
      }
    }
  } catch (error) {
    if (!window.location.pathname.includes("/auth/") && !window.location.pathname.includes("/signin")) {
      console.error("Failed to load settings:", error);
    }
  }
}

async function saveSmtp(): Promise<void> {
  try {
    const response = await fetchWithAuth("/api/v1/settings/email", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(smtpSettings.value)
    });

    if (response.ok) {
      alert("SMTP settings saved successfully");
    } else {
      alert("Failed to save SMTP settings");
    }
  } catch (error) {
    console.error("Failed to save SMTP settings:", error);
    alert("Failed to save SMTP settings");
  }
}

async function testSmtp(): Promise<void> {
  alert(t('settings.sendingTestEmail'));
}
</script>
