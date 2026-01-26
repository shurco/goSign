import { ref, computed } from "vue";
import { apiGet } from "@/services/api";

export interface CurrentUserData {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: number;
}

// Shared state across all layout components to avoid flicker when switching Sidebar <-> SettingsSidebar
const userData = ref<CurrentUserData | null>(null);
const cachedUserRole = ref<number | null>(null);

const USER_ROLE_KEY = "user_role";
const USER_CACHE_KEY = "user_cache";

function getCachedUser(): CurrentUserData | null {
  try {
    const raw = sessionStorage.getItem(USER_CACHE_KEY);
    if (!raw) return null;
    const data = JSON.parse(raw) as CurrentUserData;
    return data && typeof data.id === "string" ? data : null;
  } catch {
    return null;
  }
}

export function useCurrentUser() {
  const isAdmin = computed(() => {
    const role = userData.value?.role ?? cachedUserRole.value;
    return role === 3;
  });

  async function loadUserData(): Promise<void> {
    try {
      const cachedRole = localStorage.getItem(USER_ROLE_KEY);
      if (cachedRole) {
        cachedUserRole.value = parseInt(cachedRole, 10);
      }
      // Restore from cache immediately so UI does not flicker on remount
      if (!userData.value) {
        const cached = getCachedUser();
        if (cached) {
          userData.value = cached;
          if (cached.role !== undefined) cachedUserRole.value = cached.role;
        }
      }

      const response = await apiGet("/api/v1/users/me");
      if (response?.data) {
        const data = response.data as CurrentUserData;
        userData.value = data;
        if (data.role !== undefined) {
          cachedUserRole.value = data.role;
          localStorage.setItem(USER_ROLE_KEY, String(data.role));
        }
        sessionStorage.setItem(USER_CACHE_KEY, JSON.stringify(data));
      }
    } catch (error) {
      console.error("Failed to load user data:", error);
    }
  }

  function clearUser(): void {
    userData.value = null;
    cachedUserRole.value = null;
    sessionStorage.removeItem(USER_CACHE_KEY);
    localStorage.removeItem(USER_ROLE_KEY);
  }

  // Hydrate from cache on first use so layout shows user without waiting for API
  if (!userData.value) {
    const cached = getCachedUser();
    if (cached) {
      userData.value = cached;
      if (cached.role !== undefined) cachedUserRole.value = cached.role;
    }
  }

  return {
    userData,
    isAdmin,
    loadUserData,
    clearUser
  };
}
