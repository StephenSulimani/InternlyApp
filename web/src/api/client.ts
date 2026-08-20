import { getToken, clearSession } from "./session";

export class ApiError extends Error {
  status: number;
  body?: unknown;

  constructor(message: string, status: number, body?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }
}

const API_BASE = import.meta.env.VITE_API_URL ?? "/api";

export async function apiClient<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  const headers = new Headers(init?.headers);
  if (!headers.has("Accept")) {
    headers.set("Accept", "application/json");
  }
  if (init?.body != null && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const token = getToken();
  if (token && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
  });

  if (!res.ok) {
    if (
      res.status === 401 &&
      path !== "/login" &&
      path !== "/register"
    ) {
      clearSession();
    }

    let body: unknown;
    let message = res.statusText || "Request failed";
    try {
      body = await res.json();
      if (
        body &&
        typeof body === "object" &&
        "message" in body &&
        typeof (body as { message: string }).message === "string"
      ) {
        message = (body as { message: string }).message;
      }
    } catch {
      body = undefined;
    }
    throw new ApiError(message, res.status, body);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return res.json() as Promise<T>;
}

export { API_BASE };
