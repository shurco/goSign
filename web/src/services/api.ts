// API service for making HTTP requests
const API_BASE_URL = import.meta.env.VITE_API_URL || "/api/v1";
import { fetchWithAuth } from "@/utils/auth";

interface ApiResponse<T = any> {
  data: T;
  message?: string;
}

class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = "ApiError";
  }
}

function isAuthPage(): boolean {
  const path = window.location.pathname;
  return path.includes("/auth/") || path.includes("/signin");
}

async function apiRequest<T = any>(endpoint: string, options: RequestInit = {}): Promise<T> {
  // Normalize endpoint
  let finalUrl: string;
  if (endpoint.startsWith("/api/v1")) {
    finalUrl = endpoint;
  } else if (endpoint.startsWith("/api/")) {
    finalUrl = endpoint.replace("/api/", "/api/v1/");
  } else {
    finalUrl = `${API_BASE_URL}${endpoint.startsWith("/") ? endpoint : `/${endpoint}`}`;
  }

  const config: RequestInit = {
    headers: {
      "Content-Type": "application/json",
      ...options.headers
    },
    credentials: "include",
    ...options
  };

  try {
    if (isAuthPage()) {
      return { success: false, data: null } as T;
    }

    const response = await fetchWithAuth(finalUrl, config);

    if (isAuthPage()) {
      return { success: false, data: null } as T;
    }

    if (!response.ok) {
      if (response.status === 401) {
        // Redirect will be handled by fetchWithAuth, just return empty response
        if (isAuthPage()) {
            return { success: false, data: null } as T;
        }
      }

      const errorData = await response.json().catch(() => ({ message: "Unknown error" }));
      throw new ApiError(response.status, errorData.message || `HTTP ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    if (isAuthPage()) {
      return { success: false, data: null } as T;
    }

    if (error instanceof ApiError) {
      throw error;
    }

    throw new ApiError(0, error instanceof Error ? error.message : "Network error");
  }
}

export async function apiGet<T = any>(endpoint: string): Promise<ApiResponse<T>> {
  return apiRequest(endpoint, { method: "GET" });
}

export async function apiPost<T = any>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
  return apiRequest(endpoint, {
    method: "POST",
    body: data ? JSON.stringify(data) : undefined
  });
}

export async function apiPut<T = any>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
  return apiRequest(endpoint, {
    method: "PUT",
    body: data ? JSON.stringify(data) : undefined
  });
}

export async function apiPatch<T = any>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
  return apiRequest(endpoint, {
    method: "PATCH",
    body: data ? JSON.stringify(data) : undefined
  });
}

export async function apiDelete<T = any>(endpoint: string): Promise<ApiResponse<T>> {
  return apiRequest(endpoint, { method: "DELETE" });
}
