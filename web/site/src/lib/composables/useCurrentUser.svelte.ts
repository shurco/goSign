import { apiGet } from "@/services/api";

export interface CurrentUserData {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: number;
}

const USER_ROLE_KEY = "user_role";
const USER_CACHE_KEY = "user_cache";

// Shared state across all layout components to avoid flicker when switching Sidebar <-> SettingsSidebar
const state = $state<{ userData: CurrentUserData | null; cachedUserRole: number | null }>({
  userData: null,
  cachedUserRole: null
});

function getCachedUser(): CurrentUserData | null {
  try {
    const raw = sessionStorage.getItem(USER_CACHE_KEY);
    if (!raw) {
      return null;
    }
    const data = JSON.parse(raw) as CurrentUserData;
    return data && typeof data.id === "string" ? data : null;
  } catch {
    return null;
  }
}

function hydrateFromCache(): void {
  if (state.userData) {
    return;
  }
  const cached = getCachedUser();
  if (cached) {
    state.userData = cached;
    if (cached.role !== undefined) {
      state.cachedUserRole = cached.role;
    }
  }
}

export function useCurrentUser() {
  // Hydrate from cache on first use so layout shows user without waiting for API
  hydrateFromCache();

  const isAdmin = $derived((state.userData?.role ?? state.cachedUserRole) === 3);

  async function loadUserData(): Promise<void> {
    try {
      const cachedRole = localStorage.getItem(USER_ROLE_KEY);
      if (cachedRole) {
        state.cachedUserRole = parseInt(cachedRole, 10);
      }
      // Restore from cache immediately so UI does not flicker on remount
      hydrateFromCache();

      const response = await apiGet("/api/v1/users/me");
      if (response?.data) {
        const data = response.data as CurrentUserData;
        state.userData = data;
        if (data.role !== undefined) {
          state.cachedUserRole = data.role;
          localStorage.setItem(USER_ROLE_KEY, String(data.role));
        }
        sessionStorage.setItem(USER_CACHE_KEY, JSON.stringify(data));
      }
    } catch (error) {
      console.error("Failed to load user data:", error);
    }
  }

  function clearUser(): void {
    state.userData = null;
    state.cachedUserRole = null;
    sessionStorage.removeItem(USER_CACHE_KEY);
    localStorage.removeItem(USER_ROLE_KEY);
  }

  return {
    get userData() {
      return state.userData;
    },
    get isAdmin() {
      return isAdmin;
    },
    loadUserData,
    clearUser
  };
}
