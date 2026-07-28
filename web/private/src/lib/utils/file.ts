// Small reusable file helpers (browser-side)

import { t } from "@/i18n/index.svelte";
import { apiUrl } from "@/services/api-base";
import { fetchWithAuth } from "@/utils/auth";

/**
 * Convert a File into base64 payload (WITHOUT data URL prefix).
 * Example: "JVBERi0xLjcKJc..." for PDFs.
 */
export async function fileToBase64Payload(file: File): Promise<string> {
  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      try {
        const result = reader.result as string;
        const parts = result.split(",");
        // readAsDataURL returns "data:<mime>;base64,<payload>"
        resolve(parts.length > 1 ? parts[1] : "");
      } catch (err) {
        reject(err);
      }
    };
    reader.onerror = () => reject(new Error("Failed to read file"));
    reader.readAsDataURL(file);
  });
}

/**
 * Open a Blob in a new tab/window.
 * Useful for authenticated fetch flows (blob URLs keep auth out of the URL).
 */
export function openBlobInNewTab(blob: Blob): void {
  const url = URL.createObjectURL(blob);
  window.open(url, "_blank", "noopener,noreferrer");
  // Revoke later to avoid breaking viewers that load progressively.
  window.setTimeout(() => URL.revokeObjectURL(url), 60_000);
}

/**
 * Download the completed (signed) PDF of a submission and open it in a new tab.
 * Binary response, so it uses fetchWithAuth directly instead of the JSON api helpers.
 */
export async function openCompletedDocument(submissionId: string): Promise<void> {
  const id = String(submissionId || "");
  if (!id) {
    return;
  }
  try {
    const res = await fetchWithAuth(apiUrl(`/signing-links/${encodeURIComponent(id)}/document`), { method: "GET" });
    if (res.status === 409) {
      alert(t("submissionStatus.errors.notCompletedYet"));
      return;
    }
    if (res.status === 404 || res.status === 403) {
      alert(t("submissionStatus.errors.onlyOwnerCanView"));
      return;
    }
    if (!res.ok) {
      alert(t("submissionStatus.errors.failedToLoadDocument"));
      return;
    }
    const buf = await res.arrayBuffer();
    openBlobInNewTab(new Blob([buf], { type: "application/pdf" }));
  } catch (e) {
    console.error("Failed to open completed document:", e);
    alert(t("submissionStatus.errors.failedToLoadDocument"));
  }
}
