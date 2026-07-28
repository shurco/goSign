/**
 * Single source of truth for the API base URL.
 *
 * Resolution order:
 * 1. Runtime config (static/config.js, regenerated at container start
 *    from GOSIGN_API_URL) — lets one prebuilt image serve any domain.
 * 2. VITE_API_URL at build time.
 * 3. Same-origin "/v1" (dev proxy, single domain).
 */
declare global {
  interface Window {
    __GOSIGN_API_URL__?: string;
  }
}

const runtimeApiUrl = typeof window !== "undefined" ? window.__GOSIGN_API_URL__ : undefined;

export const API_BASE_URL: string = (runtimeApiUrl || import.meta.env.VITE_API_URL || "/v1").replace(/\/$/, "");

/** Build a full API URL from an API-relative path like "/templates" or "/auth/signin". */
export function apiUrl(path: string): string {
  return `${API_BASE_URL}${path.startsWith("/") ? path : `/${path}`}`;
}

/** Origin of the API ("" for same-origin deployments). */
export const API_ORIGIN: string =
  API_BASE_URL.startsWith("http://") || API_BASE_URL.startsWith("https://") ? new URL(API_BASE_URL).origin : "";

/**
 * Resolve a backend file path (e.g. "/drive/pages/...") against the API origin.
 * Absolute URLs (external storage) are returned unchanged.
 */
export function fileUrl(path: string): string {
  if (path.startsWith("http://") || path.startsWith("https://")) {
    return path;
  }
  return `${API_ORIGIN}${path}`;
}

/** API sub-paths that never require authentication. */
const PUBLIC_PREFIXES = ["/auth/", "/public/", "/verify/", "/invitations/"];

function basePathname(): string {
  if (API_BASE_URL.startsWith("http://") || API_BASE_URL.startsWith("https://")) {
    try {
      return new URL(API_BASE_URL).pathname.replace(/\/$/, "");
    } catch {
      return "/v1";
    }
  }
  return API_BASE_URL.replace(/\/$/, "");
}

/** True when the URL points into the protected part of the goSign API. */
export function requiresAuth(url: string): boolean {
  let pathname = url;
  if (url.startsWith("http://") || url.startsWith("https://")) {
    try {
      pathname = new URL(url).pathname;
    } catch {
      return false;
    }
  }

  const base = basePathname();
  if (!pathname.startsWith(`${base}/`)) {
    return false;
  }

  const rest = pathname.slice(base.length);
  return !PUBLIC_PREFIXES.some((prefix) => rest.startsWith(prefix));
}
