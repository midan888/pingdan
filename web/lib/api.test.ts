import { afterEach, describe, expect, it, vi } from "vitest";
import {
  daysUntil,
  groupStatusColor,
  intervalLabel,
  monitorTargetSummary,
  sslSeverity,
  supportsSSLMonitoring,
} from "./api";

describe("api utility helpers", () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it("summarizes a group by its worst endpoint state", () => {
    expect(groupStatusColor(["up", "up"])).toBe("var(--up)");
    expect(groupStatusColor(["up", "down"])).toBe("var(--down)");
    expect(groupStatusColor(["unknown"])).toBe("var(--unknown)");
    expect(groupStatusColor([])).toBe("var(--unknown)");
  });

  it("calculates whole days until an ISO timestamp", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-07-10T12:00:00Z"));

    expect(daysUntil("2026-07-11T11:59:00Z")).toBe(0);
    expect(daysUntil("2026-07-12T00:00:00Z")).toBe(1);
    expect(daysUntil("2026-07-10T11:00:00Z")).toBe(-1);
  });

  it("buckets SSL expiry severity", () => {
    expect(sslSeverity(-1)).toBe("expired");
    expect(sslSeverity(0)).toBe("critical");
    expect(sslSeverity(15)).toBe("critical");
    expect(sslSeverity(16)).toBe("warn");
    expect(sslSeverity(30)).toBe("warn");
    expect(sslSeverity(31)).toBe("ok");
  });

  it("formats endpoint intervals", () => {
    expect(intervalLabel(45)).toBe("45s");
    expect(intervalLabel(60)).toBe("1 min");
    expect(intervalLabel(300)).toBe("5 min");
    expect(intervalLabel(3600)).toBe("1 hr");
    expect(intervalLabel(86400)).toBe("1 day");
    expect(intervalLabel(172800)).toBe("2 days");
  });

  it("formats protocol-aware monitor targets", () => {
    expect(monitorTargetSummary({ checkType: "http", method: "GET", url: "https://example.com" }))
      .toBe("HTTP GET https://example.com");
    expect(monitorTargetSummary({ checkType: "tcp", method: "GET", url: "tcp://vpn.example.com:443" }))
      .toBe("TCP tcp://vpn.example.com:443");
    expect(monitorTargetSummary({ checkType: "icmp", method: "GET", url: "icmp://vpn.example.com" }))
      .toBe("ICMP icmp://vpn.example.com");
  });

  it("enables SSL monitoring only for HTTPS HTTP targets", () => {
    expect(supportsSSLMonitoring({ checkType: "http", url: "https://example.com" })).toBe(true);
    expect(supportsSSLMonitoring({ checkType: "http", url: "http://example.com" })).toBe(false);
    expect(supportsSSLMonitoring({ checkType: "tcp", url: "tcp://example.com:443" })).toBe(false);
  });
});
