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

/** Fetch wrapper for public, unauthenticated endpoints (e.g. status pages). */
export async function publicApi<T = unknown>(path: string): Promise<T> {
  const res = await fetch(`${API_URL}${path}`);
  if (res.status === 404) throw new NotFoundError();
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || res.statusText);
  }
  return res.json() as Promise<T>;
}

export class NotFoundError extends Error {
  constructor() {
    super("not found");
    this.name = "NotFoundError";
  }
}

export type Me = {
  id: string;
  email: string;
  isAdmin: boolean;
};

export type AdminStats = {
  userCount: number;
  endpointCount: number;
};

export type AdminUser = {
  id: string;
  email: string;
  name: string | null;
  provider: string | null;
  createdAt: string;
  endpointCount: number;
};

export type Group = {
  id: string;
  name: string;
  createdAt: string;
};

export type StatusPage = {
  id: string;
  slug: string;
  title: string;
  description: string;
  createdAt: string;
  updatedAt: string;
};

export type StatusPageItem = {
  endpointId: string;
  displayName: string | null;
  position: number;
};

/** A single public check, reduced to what is safe to expose. */
export type PublicTick = {
  ok: boolean;
  checkedAt: string;
};

export type PublicStatusItem = {
  name: string;
  state: EndpointState;
  uptimePct: number;
  history: PublicTick[];
};

export type PublicStatusPage = {
  title: string;
  description: string;
  overall: "up" | "down" | "degraded" | "unknown";
  updatedAt: string;
  items: PublicStatusItem[];
};

export type EndpointState = "up" | "down" | "unknown";

/**
 * Accent color for a group section, driven by the worst state among its
 * endpoints: red if any is down, green if all are up, neutral otherwise
 * (pending/unknown, or an empty group). Mirrors the status pill colors so the
 * group header reads as a status at a glance.
 */
export function groupStatusColor(states: EndpointState[]): string {
  if (states.some((s) => s === "down")) return "var(--down)";
  if (states.length > 0 && states.every((s) => s === "up")) return "var(--up)";
  return "var(--unknown)";
}

export type Endpoint = {
  id: string;
  groupId: string | null;
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
  sslExpiresAt: string | null;
  sslLastCheckedAt: string | null;
  sslLastError: string | null;
};

/** Daily warnings begin at or below this many days to expiry. */
export const SSL_ALERT_THRESHOLD_DAYS = 15;

/** Whole days until an ISO timestamp, rounded down. Negative = already past. */
export function daysUntil(iso: string): number {
  const ms = new Date(iso).getTime() - Date.now();
  return Math.floor(ms / 86_400_000);
}

/** Severity bucket for an SSL countdown, used to colour the UI. */
export function sslSeverity(daysLeft: number): "ok" | "warn" | "critical" | "expired" {
  if (daysLeft < 0) return "expired";
  if (daysLeft <= SSL_ALERT_THRESHOLD_DAYS) return "critical";
  if (daysLeft <= 30) return "warn";
  return "ok";
}

export type AlertChannelKind =
  | "email"
  | "telegram"
  | "slack"
  | "discord"
  | "teams"
  | "webhook"
  | "pagerduty"
  | "ntfy"
  | "pushover"
  | "twilio_sms"
  | "opsgenie";

export type AlertChannel = {
  id: string;
  kind: AlertChannelKind;
  label: string;
  config: Record<string, unknown>;
};

export type Capabilities = {
  alertChannelKinds: Partial<Record<AlertChannelKind, boolean>>;
};

/** Human-friendly label for a check interval in seconds. */
export function intervalLabel(sec: number): string {
  if (sec % 86400 === 0) {
    const d = sec / 86400;
    return `${d} day${d === 1 ? "" : "s"}`;
  }
  if (sec % 3600 === 0) return `${sec / 3600} hr`;
  if (sec % 60 === 0) return `${sec / 60} min`;
  return `${sec}s`;
}

export type AssertionSource =
  | "status_code"
  | "response_time"
  | "header"
  | "body"
  | "json_path";

export type AssertionComparison =
  | "equals"
  | "not_equals"
  | "greater_than"
  | "less_than"
  | "contains"
  | "not_contains"
  | "matches";

export type Assertion = {
  id?: number;
  source: AssertionSource;
  property: string;
  comparison: AssertionComparison;
  target: string;
};

export type FailedAssertion = {
  source: AssertionSource;
  property?: string;
  comparison: AssertionComparison;
  target: string;
  actual: string;
  passed: boolean;
};

export type Check = {
  id: number;
  endpointId: string;
  statusCode?: number;
  latencyMs?: number;
  ok: boolean;
  error?: string;
  failedAssertions?: FailedAssertion[];
  checkedAt: string;
};

export type EndpointStats = {
  total: number;
  upCount: number;
  uptimePct: number;
  avgLatencyMs: number | null;
  p50LatencyMs: number | null;
  p95LatencyMs: number | null;
  minLatencyMs: number | null;
  maxLatencyMs: number | null;
};

export type EndpointDetail = {
  endpoint: Endpoint;
  assertions: Assertion[];
  channelIds: string[];
};
