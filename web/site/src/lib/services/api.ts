import { fetchWithAuth } from "@/utils/auth";

const API_BASE_URL = import.meta.env.VITE_API_URL || "/api/v1";

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

async function apiRequest<T = any>(endpoint: string, options: RequestInit = {}): Promise<T> {
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

  const response = await fetchWithAuth(finalUrl, config);

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ message: "Unknown error" }));
    throw new ApiError(response.status, errorData.message || `HTTP ${response.status}`);
  }

  return await response.json();
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
