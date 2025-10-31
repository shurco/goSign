/**
 * Authentication utilities for token management
 */

let isRefreshing = false;
let refreshPromise: Promise<string | null> | null = null;

/**
 * Clear tokens and redirect to login
 */
function clearTokensAndRedirect(): void {
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
  if (window.location.pathname !== "/auth/signin") {
    window.location.href = "/auth/signin";
  }
}

/**
 * Refresh access token using refresh token
 * @returns New access token or null if refresh failed
 */
async function refreshAccessToken(): Promise<string | null> {
  if (isRefreshing && refreshPromise) {
    return refreshPromise;
  }

  isRefreshing = true;
  refreshPromise = (async () => {
    try {
      const refreshToken = localStorage.getItem("refresh_token");
      if (!refreshToken) {
        return null;
      }

      const response = await fetch("/auth/refresh", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken })
      });

      if (!response.ok) {
        clearTokensAndRedirect();
        return null;
      }

      const data = await response.json();
      if (data.data?.access_token) {
        localStorage.setItem("access_token", data.data.access_token);
      }
      if (data.data?.refresh_token) {
        localStorage.setItem("refresh_token", data.data.refresh_token);
      }

      return data.data?.access_token || null;
    } catch {
      clearTokensAndRedirect();
      return null;
    } finally {
      isRefreshing = false;
      refreshPromise = null;
    }
  })();

  return refreshPromise;
}

/**
 * Handle 401 Unauthorized response by refreshing token and retrying request
 * @param originalRequest Original fetch request options
 * @returns Response from retried request or original response
 */
export async function handleUnauthorized(
  originalRequest: RequestInit & { url: string }
): Promise<Response | null> {
  const newToken = await refreshAccessToken();
  
  if (!newToken) {
    // Refresh failed, user will be redirected to login
    return null;
  }

  // Retry original request with new token
  const headers = new Headers(originalRequest.headers);
  headers.set("Authorization", `Bearer ${newToken}`);
  
  return fetch(originalRequest.url, {
    ...originalRequest,
    headers
  });
}

/**
 * Get Authorization header with current token
 * @returns Authorization header string or empty object
 */
export function getAuthHeaders(): Record<string, string> {
  const token = localStorage.getItem("access_token");
  if (!token) {
    return {};
  }
  return {
    Authorization: `Bearer ${token}`
  };
}

/**
 * Fetch with automatic token refresh on 401
 * @param url Request URL
 * @param options Fetch options
 * @returns Response from fetch
 */
export async function fetchWithAuth(url: string, options: RequestInit = {}): Promise<Response> {
  const isAPIv1 = url.startsWith("/api/v1");
  if (isAPIv1) {
    options.headers = {
      ...getAuthHeaders(),
      ...options.headers
    };
  }

  let response = await fetch(url, options);

  if (response.status === 401 && isAPIv1) {
    const retryResponse = await handleUnauthorized({ ...options, url });
    return retryResponse || response;
  }

  return response;
}

