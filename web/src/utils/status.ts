export type BadgeVariant = "ghost" | "primary" | "success" | "warning" | "error" | "info";

function normalizeStatus(status: unknown): string {
  return String(status || "")
    .trim()
    .toLowerCase();
}

// Use a single i18n namespace for all statuses coming from backend.
// This keeps translations consistent across dashboard, submissions, and modals.
export function getI18nStatusKey(status: unknown): string {
  return `status.${normalizeStatus(status)}`;
}

export function getBadgeVariantForSubmissionStatus(status: unknown): BadgeVariant {
  const s = normalizeStatus(status);
  switch (s) {
    case "draft":
      return "ghost";
    case "pending":
      return "warning";
    case "in_progress":
      return "info";
    case "completed":
      return "success";
    case "declined":
      return "error";
    case "expired":
      return "error";
    case "cancelled":
      return "ghost";
    default:
      return "ghost";
  }
}

export function getBadgeVariantForSubmitterStatus(status: unknown): BadgeVariant {
  const s = normalizeStatus(status);
  switch (s) {
    case "pending":
      return "warning";
    case "opened":
      return "info";
    case "completed":
      return "success";
    case "declined":
      return "error";
    default:
      return "ghost";
  }
}

