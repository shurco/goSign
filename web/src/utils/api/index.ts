import { done, start } from "nprogress";
import { getAuthHeaders, handleUnauthorized } from "./auth";

type RequestMethod = "GET" | "POST" | "PATCH" | "PUT" | "DELETE";

interface RequestOptions extends RequestInit {
  method: RequestMethod;
  credentials: RequestCredentials;
}

async function request<T = unknown>(url: string, method: RequestMethod, body?: unknown): Promise<T> {
  const options: RequestOptions = {
    method,
    credentials: "include"
  };

  const isAPIv1 = url.startsWith("/api/v1");
  if (isAPIv1) {
    options.headers = { ...getAuthHeaders() };
  }

  if (body !== undefined) {
    options.body = typeof body === "object" ? JSON.stringify(body) : (body as BodyInit);
    if (typeof body === "object") {
      options.headers = {
        ...options.headers,
        "Content-Type": "application/json"
      };
    }
  }

  try {
    start();
    let response = await fetch(url, options);
    
    if (response.status === 401 && isAPIv1) {
      const retryResponse = await handleUnauthorized({ ...options, url });
      if (retryResponse) {
        response = retryResponse;
      } else {
        throw new Error("Authentication failed");
      }
    }
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    return await response.json();
  } finally {
    done();
  }
}

export const apiGet = <T = unknown>(url: string): Promise<T> => request<T>(url, "GET");
export const apiPost = <T = unknown>(url: string, body: unknown): Promise<T> => request<T>(url, "POST", body);
export const apiUpdate = <T = unknown>(url: string, body: unknown): Promise<T> => request<T>(url, "PATCH", body);
export const apiDelete = <T = unknown>(url: string): Promise<T> => request<T>(url, "DELETE");
