import { fetchWithAuth } from "@/utils/auth";
import { apiUrl } from "@/services/api-base";

export { apiUrl, API_BASE_URL } from "@/services/api-base";

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
  const finalUrl = apiUrl(endpoint);

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
