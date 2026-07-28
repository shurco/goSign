import { redirect } from "@sveltejs/kit";
import { apiGet } from "@/services/api";

/**
 * Client-side auth guard (the app runs with ssr=false, so load functions
 * always execute in the browser). Mirrors the previous router.beforeEach.
 */
export function requireAuth(fullPath: string): void {
  const token = localStorage.getItem("access_token");
  if (!token) {
    redirect(307, `/auth/signin?redirect=${encodeURIComponent(fullPath)}`);
  }
}

let cachedAdminRole: { role: number | null; checkedAt: number } = { role: null, checkedAt: 0 };
const ADMIN_CACHE_TTL = 5 * 60 * 1000;

if (typeof window !== "undefined") {
  window.addEventListener("gosign:clear-admin-cache", () => {
    cachedAdminRole = { role: null, checkedAt: 0 };
  });
}

/** Admin guard with a 5-minute role cache, invalidated by "gosign:clear-admin-cache". */
export async function requireAdmin(fullPath: string): Promise<void> {
  const now = Date.now();
  if (!cachedAdminRole.role || now - cachedAdminRole.checkedAt > ADMIN_CACHE_TTL) {
    try {
      const response = await apiGet("/users/me");
      cachedAdminRole = { role: response.data?.role ?? null, checkedAt: now };
    } catch (error) {
      console.error("Failed to check admin access:", error);
      cachedAdminRole = { role: null, checkedAt: 0 };
      redirect(307, `/auth/signin?redirect=${encodeURIComponent(fullPath)}`);
    }
  }
  if (cachedAdminRole.role !== 3) {
    redirect(307, "/dashboard");
  }
}

/** Redirect authenticated users away from signin/signup pages. */
export function redirectIfAuthenticated(): void {
  const token = localStorage.getItem("access_token");
  if (token) {
    redirect(307, "/dashboard");
  }
}
