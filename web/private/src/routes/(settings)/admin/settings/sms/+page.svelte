<script lang="ts">
  import { onMount } from "svelte";
  import { ApiError, apiGet, apiPost, apiPut } from "@/services/api";
  import { t } from "@/i18n/index.svelte";
  import FormControl from "@/components/ui/FormControl.svelte";
  import Input from "@/components/ui/Input.svelte";
  import Button from "@/components/ui/Button.svelte";

  let sms = $state({
    twilio_enabled: false,
    twilio_account_sid: "",
    twilio_from_number: "",
    twilio_auth_token: "", // write-only
    twilio_auth_token_set: false
  });

  let test = $state({
    to_phone: "",
    message: ""
  });

  onMount(async () => {
    await load();
  });

  async function load(): Promise<void> {
    try {
      const data = await apiGet("/settings");
      const settings = data.data || data;
      if (settings.sms) {
        sms.twilio_enabled = !!settings.sms.twilio_enabled;
        sms.twilio_account_sid = String(settings.sms.twilio_account_sid || "");
        sms.twilio_from_number = String(settings.sms.twilio_from_number || "");
        sms.twilio_auth_token_set = !!settings.sms.twilio_auth_token_set;
        sms.twilio_auth_token = "";
      }
    } catch (e) {
      console.error("Failed to load SMS settings:", e);
    }
  }

  async function save(): Promise<void> {
    try {
      await apiPut("/settings/sms", {
        twilio_enabled: sms.twilio_enabled,
        twilio_account_sid: sms.twilio_account_sid,
        twilio_from_number: sms.twilio_from_number,
        twilio_auth_token: sms.twilio_auth_token
      });
      sms.twilio_auth_token = "";
      await load();
      alert("SMS settings saved successfully");
    } catch (e) {
      console.error("Failed to save SMS settings:", e);
      alert("Failed to save SMS settings");
    }
  }

  async function sendTest(): Promise<void> {
    const to = test.to_phone?.trim();
    if (!to) {
      alert(t("settings.testSmsToRequired"));
      return;
    }
    try {
      await apiPost("/settings/sms/test", { to_phone: to, message: test.message });
      alert(t("settings.testSmsSent"));
    } catch (e) {
      console.error("Failed to send test SMS:", e);
      alert(e instanceof ApiError ? e.message : t("settings.testSmsFailed"));
    }
  }
</script>

<div class="space-y-4">
  <div class="flex items-center justify-between">
    <div>
      <div class="text-sm font-semibold">{t("settings.smsConfiguration")}</div>
      <div class="text-xs text-[--color-base-content]/70">{t("settings.smsDescription")}</div>
    </div>
    <input bind:checked={sms.twilio_enabled} type="checkbox" class="toggle toggle-sm" />
  </div>

  <FormControl label={t("settings.twilioAccountSid")}>
    <Input bind:value={sms.twilio_account_sid} type="text" placeholder="ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" />
  </FormControl>

  <FormControl label={t("settings.twilioAuthToken")}>
    <Input bind:value={sms.twilio_auth_token} type="password" placeholder="••••••••••••••••••••••••••••••••" />
    {#if sms.twilio_auth_token_set}
      <div class="mt-1 text-xs text-[--color-base-content]/70">
        Token is already set (leave empty to keep unchanged).
      </div>
    {/if}
  </FormControl>

  <FormControl label={t("settings.twilioFromNumber")}>
    <Input bind:value={sms.twilio_from_number} type="tel" placeholder="+15551234567" />
  </FormControl>

  <div class="divider"></div>

  <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
    <FormControl label={t("settings.testSmsTo")}>
      <Input bind:value={test.to_phone} type="tel" placeholder="+15551234567" />
    </FormControl>
    <FormControl label={t("settings.testSmsMessage")}>
      <Input bind:value={test.message} type="text" placeholder={t("settings.testSmsMessagePlaceholder")} />
    </FormControl>
  </div>

  <div class="flex justify-end gap-3 pt-4">
    <Button variant="ghost" onclick={sendTest}>{t("settings.testSms")}</Button>
    <Button variant="primary" onclick={save}>{t("common.save")}</Button>
  </div>
</div>
