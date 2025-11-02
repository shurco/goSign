/**
 * Authentication utilities for token management
 */

let isRefreshing = false;
let refreshPromise: Promise<string | null> | null = null;
let routerInstance: any = null;

/**
 * Set router instance for redirects
 */
export function setAuthRouter(router: any): void {
  routerInstance = router;
}

/**
 * Check if current path is an auth page
 */
function isAuthPage(): boolean {
  const path = window.location.pathname;
  return path.startsWith("/auth/") || path === "/signin" || path === "/signup";
}

/**
 * Redirect to login page (unified function)
 */
function redirectToLogin(): void {
  if (isAuthPage()) {
    return;
  }

    if (routerInstance) {
      try {
        routerInstance.push("/auth/signin");
        return;
      } catch {
        // Fall through to window.location
      }
    }

    window.location.href = "/auth/signin";
  }

/**
 * Clear tokens and redirect to login
 */
function clearTokensAndRedirect(): void {
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
  redirectToLogin();
}

/**
 * Logout user - clear tokens and redirect to login
 * Optionally invalidate refresh token on server
 */
export async function logout(): Promise<void> {
  const refreshToken = localStorage.getItem("refresh_token");

  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");

  // Try to invalidate refresh token on server (optional, don't wait for response)
  if (refreshToken) {
    try {
      await fetch("/auth/signout", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken })
      }).catch(() => {
        // Ignore errors
      });
    } catch {
      // Ignore errors
    }
  }

  redirectToLogin();
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
        credentials: "include",
        body: JSON.stringify({ refresh_token: refreshToken })
      });

      if (!response.ok) {
        if (response.status === 401 || response.status === 403) {
          clearTokensAndRedirect();
        }
        return null;
      }

      const data = await response.json();
      const newAccessToken = data.data?.access_token || data.access_token;
      const newRefreshToken = data.data?.refresh_token || data.refresh_token;

      if (newAccessToken) {
        localStorage.setItem("access_token", newAccessToken);
      }
      if (newRefreshToken) {
        localStorage.setItem("refresh_token", newRefreshToken);
      }

      return newAccessToken || null;
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
 * Get Authorization header with current token
 */
export function getAuthHeaders(): Record<string, string> {
  const token = localStorage.getItem("access_token");
  if (!token || token.trim() === "") {
    return {};
  }
  return {
    Authorization: `Bearer ${token.trim()}`
  };
}

/**
 * Merge headers from options into new Headers object
 */
function mergeHeaders(options: RequestInit): Headers {
    const headers = new Headers();

    if (options.headers) {
      if (options.headers instanceof Headers) {
        options.headers.forEach((value, key) => {
          headers.set(key, value);
        });
      } else if (Array.isArray(options.headers)) {
        options.headers.forEach(([key, value]) => {
          headers.set(key, value);
        });
      } else {
        Object.entries(options.headers).forEach(([key, value]) => {
          if (typeof value === "string") {
            headers.set(key, value);
          }
        });
      }
    }

  return headers;
}

/**
 * Create unauthorized response
 */
function createUnauthorizedResponse(): Response {
  return new Response(JSON.stringify({ error: "Unauthorized" }), {
    status: 401,
    statusText: "Unauthorized",
    headers: { "Content-Type": "application/json" }
  });
}

/**
 * Fetch with automatic token refresh on 401
 */
export async function fetchWithAuth(url: string, options: RequestInit = {}): Promise<Response> {
  const isAPIv1 = url.startsWith("/api/v1");
  const isLegacyAPI = url.startsWith("/api/") && !url.startsWith("/api/v1");
  const requiresAuth = isAPIv1 || isLegacyAPI;

  if (requiresAuth && isAuthPage()) {
    return createUnauthorizedResponse();
  }

  // Get or refresh token if auth is required
  if (requiresAuth) {
    let token = localStorage.getItem("access_token");

    if (!token) {
      const newToken = await refreshAccessToken();
      if (!newToken) {
        return createUnauthorizedResponse();
      }
      token = localStorage.getItem("access_token");
    }

    if (!token || token.trim() === "") {
      return createUnauthorizedResponse();
    }

    const headers = mergeHeaders(options);
    headers.set("Authorization", `Bearer ${token.trim()}`);
    options.headers = headers;
  }

  let response: Response;
  try {
    response = await fetch(url, options);
  } catch (error) {
    throw error;
  }

  // Handle 401 with token refresh
  if (response.status === 401 && requiresAuth) {
    if (isRefreshing && refreshPromise) {
      await refreshPromise;
      if (isAuthPage()) {
        return createUnauthorizedResponse();
      }
      const refreshedToken = localStorage.getItem("access_token");
      if (!refreshedToken || refreshedToken.trim() === "") {
        return createUnauthorizedResponse();
      }
      const newHeaders = mergeHeaders(options);
      newHeaders.set("Authorization", `Bearer ${refreshedToken.trim()}`);
      return fetch(url, { ...options, headers: newHeaders });
    }

    const newToken = await refreshAccessToken();
    if (!newToken) {
      return createUnauthorizedResponse();
    }

    if (isAuthPage()) {
      return createUnauthorizedResponse();
    }

    const newHeaders = mergeHeaders(options);
    newHeaders.set("Authorization", `Bearer ${newToken.trim()}`);
    return fetch(url, { ...options, headers: newHeaders });
  }

  return response;
}
