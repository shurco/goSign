<script lang="ts">
  import { onMount } from "svelte";
  import { t } from "@/i18n/index.svelte";
  import FormControl from "@/components/ui/FormControl.svelte";
  import Input from "@/components/ui/Input.svelte";
  import Select from "@/components/ui/Select.svelte";
  import Button from "@/components/ui/Button.svelte";
  import { fetchWithAuth } from "@/utils/auth";

  let smtpSettings = $state({
    host: "",
    port: "587",
    encryption: "tls",
    username: "",
    password: "",
    from_email: "",
    from_name: "goSign",
    provider: "smtp"
  });

  onMount(async () => {
    await loadSettings();
  });

  async function loadSettings(): Promise<void> {
    try {
      const response = await fetchWithAuth("/api/v1/settings");
      if (response.ok) {
        const data = await response.json();
        const settings = data.data || data;
        if (settings.email) {
          smtpSettings = {
            provider: settings.email.provider || "smtp",
            host: settings.email.smtp_host || "",
            port: String(settings.email.smtp_port || "587"),
            encryption: "tls",
            username: settings.email.smtp_user || "",
            password: "",
            from_email: settings.email.from_email || "",
            from_name: settings.email.from_name || "goSign"
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
      // Convert UI fields to API format
      const payload = {
        provider: smtpSettings.provider,
        smtp_host: smtpSettings.host,
        smtp_port: smtpSettings.port,
        smtp_user: smtpSettings.username,
        smtp_pass: smtpSettings.password,
        from_email: smtpSettings.from_email,
        from_name: smtpSettings.from_name
      };

      const response = await fetchWithAuth("/api/v1/settings/email", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
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
    alert(t("settings.sendingTestEmail"));
  }
</script>

<div class="space-y-4">
  <FormControl label={t("settings.smtpHost")}>
    <Input bind:value={smtpSettings.host} type="text" placeholder="smtp.gmail.com" />
  </FormControl>

  <div class="grid grid-cols-2 gap-4">
    <FormControl label={t("settings.port")}>
      <Input bind:value={smtpSettings.port} type="number" placeholder="587" />
    </FormControl>

    <FormControl label={t("settings.encryption")}>
      <Select bind:value={smtpSettings.encryption}>
        <option value="tls">{t("settings.tls")}</option>
        <option value="ssl">{t("settings.ssl")}</option>
        <option value="none">{t("settings.none")}</option>
      </Select>
    </FormControl>
  </div>

  <FormControl label={t("settings.username")}>
    <Input bind:value={smtpSettings.username} type="text" />
  </FormControl>

  <FormControl label={t("settings.password")}>
    <Input bind:value={smtpSettings.password} type="password" />
  </FormControl>

  <FormControl label={t("settings.fromEmail")}>
    <Input bind:value={smtpSettings.from_email} type="email" placeholder="noreply@example.com" />
  </FormControl>

  <div class="flex justify-end gap-3 pt-4">
    <Button variant="ghost" onclick={testSmtp}>{t("settings.testConnection")}</Button>
    <Button variant="primary" onclick={saveSmtp}>{t("common.save")}</Button>
  </div>
</div>
