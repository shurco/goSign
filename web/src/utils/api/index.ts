import { done, start } from "nprogress";

type RequestMethod = "GET" | "POST" | "PATCH" | "DELETE";

interface RequestOptions extends RequestInit {
  method: RequestMethod;
  credentials: RequestCredentials;
}

async function request<T = unknown>(url: string, method: RequestMethod, body?: unknown): Promise<T> {
  const options: RequestOptions = {
    method,
    credentials: "include"
  };

  if (body !== undefined) {
    options.body = typeof body === "object" ? JSON.stringify(body) : (body as BodyInit);
    if (typeof body === "object") {
      options.headers = { "Content-Type": "application/json" };
    }
  }

  try {
    start();
    const response = await fetch(url, options);
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
