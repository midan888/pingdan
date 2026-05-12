export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const TOKEN_KEY = "pingdan_token";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return window.localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  window.localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  window.localStorage.removeItem(TOKEN_KEY);
}

export async function api<T = unknown>(path: string, init: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers = new Headers(init.headers);
  headers.set("Content-Type", "application/json");
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const res = await fetch(`${API_URL}${path}`, { ...init, headers });
  if (res.status === 401) {
    clearToken();
    if (typeof window !== "undefined") window.location.href = "/login";
    throw new Error("unauthorized");
  }
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export type Endpoint = {
  id: string;
  name: string;
  url: string;
  method: string;
  expectedStatus: number;
  intervalSec: number;
  timeoutSec: number;
  failureThreshold: number;
  enabled: boolean;
  currentState: "up" | "down" | "unknown";
  consecutiveFailures: number;
  lastCheckedAt: string | null;
  createdAt: string;
};

export type AlertChannel = {
  id: string;
  kind: "email" | "telegram";
  label: string;
  config: Record<string, unknown>;
};
