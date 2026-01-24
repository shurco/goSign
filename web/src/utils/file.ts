// Small reusable file helpers (browser-side)

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

