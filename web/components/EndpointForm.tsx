"use client";

import { FormEvent, useEffect, useState } from "react";
import Link from "next/link";
import { api, type AlertChannel, type Assertion } from "@/lib/api";
import { AssertionBuilder } from "./AssertionBuilder";

export type EndpointFormValues = {
  name: string;
  url: string;
  method: string;
  expectedStatus: number;
  intervalSec: number;
  timeoutSec: number;
  failureThreshold: number;
  assertions: Assertion[];
  channelIds: string[];
};

// Fixed monitoring intervals (in seconds).
export const INTERVALS = [
  { sec: 60, label: "1 min" },
  { sec: 120, label: "2 min" },
  { sec: 180, label: "3 min" },
  { sec: 300, label: "5 min" },
  { sec: 480, label: "8 min" },
];

export function emptyEndpoint(): EndpointFormValues {
  return {
    name: "",
    url: "",
    method: "GET",
    expectedStatus: 200,
    intervalSec: 60,
    timeoutSec: 10,
    failureThreshold: 2,
    assertions: [],
    channelIds: [],
  };
}

const METHODS = ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"];
const STATUS_PRESETS = [200, 201, 204, 301, 302, 401, 403];

function Section({
  num,
  title,
  desc,
  children,
}: {
  num: number;
  title: string;
  desc?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="form-section">
      <div className="sec-head">
        <span className="sec-num">{num}</span>
        <span className="sec-title">{title}</span>
        {desc && <span className="sec-desc">{desc}</span>}
      </div>
      <div className="sec-body">{children}</div>
    </div>
  );
}

function Stepper({
  value,
  onChange,
  min = 1,
  max = 999,
  unit,
}: {
  value: number;
  onChange: (n: number) => void;
  min?: number;
  max?: number;
  unit?: string;
}) {
  const clamp = (n: number) => Math.max(min, Math.min(max, n));
  return (
    <div className="stepper">
      <button type="button" onClick={() => onChange(clamp(value - 1))} aria-label="decrease">−</button>
      <span className="val">{value}{unit && <small>{unit}</small>}</span>
      <button type="button" onClick={() => onChange(clamp(value + 1))} aria-label="increase">+</button>
    </div>
  );
}

export function EndpointForm({
  initial,
  submitLabel,
  onSubmit,
  onCancel,
}: {
  initial: EndpointFormValues;
  submitLabel: string;
  onSubmit: (v: EndpointFormValues) => Promise<void>;
  onCancel?: () => void;
}) {
  const [v, setV] = useState<EndpointFormValues>(initial);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [channels, setChannels] = useState<AlertChannel[]>([]);

  useEffect(() => {
    api<AlertChannel[]>("/alert-channels").then(setChannels).catch(() => setChannels([]));
  }, []);

  function set<K extends keyof EndpointFormValues>(key: K, val: EndpointFormValues[K]) {
    setV((prev) => ({ ...prev, [key]: val }));
  }

  function toggleChannel(id: string) {
    setV((prev) => ({
      ...prev,
      channelIds: prev.channelIds.includes(id)
        ? prev.channelIds.filter((c) => c !== id)
        : [...prev.channelIds, id],
    }));
  }

  async function submit(ev: FormEvent) {
    ev.preventDefault();
    setSaving(true);
    setError(null);
    try {
      await onSubmit(v);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }

  const intervalLabel = INTERVALS.find((i) => i.sec === v.intervalSec)?.label ?? `${v.intervalSec}s`;

  return (
    <form onSubmit={submit}>
      {/* 1 — Request */}
      <Section num={1} title="Request" desc="What to call">
        <div className="field">
          <label>Endpoint URL</label>
          <div className="url-bar">
            <select value={v.method} onChange={(e) => set("method", e.target.value)} aria-label="HTTP method">
              {METHODS.map((m) => <option key={m} value={m}>{m}</option>)}
            </select>
            <input
              placeholder="https://api.example.com/healthz"
              value={v.url}
              onChange={(e) => set("url", e.target.value)}
              required
              autoFocus
            />
          </div>
          <div className="hint">The HTTP request we&apos;ll send on every check.</div>
        </div>

        <div className="field" style={{ marginBottom: 0 }}>
          <label>Display name</label>
          <input style={{ maxWidth: 360 }} placeholder="Production API" value={v.name} onChange={(e) => set("name", e.target.value)} required />
        </div>
      </Section>

      {/* 2 — Schedule */}
      <Section num={2} title="Schedule" desc={`Every ${intervalLabel}`}>
        <div className="field">
          <label>Check interval</label>
          <div className="chips">
            {INTERVALS.map((i) => (
              <button
                type="button"
                key={i.sec}
                className={`chip ${v.intervalSec === i.sec ? "selected" : ""}`}
                onClick={() => set("intervalSec", i.sec)}
              >
                {i.label}
              </button>
            ))}
          </div>
        </div>

        <div className="row wrap" style={{ gap: "2rem", marginBottom: 0 }}>
          <div>
            <label>Request timeout</label>
            <Stepper value={v.timeoutSec} onChange={(n) => set("timeoutSec", n)} min={1} max={60} unit="s" />
            <div className="hint">Fail the check if no response within this time.</div>
          </div>
          <div>
            <label>Failures before alerting</label>
            <Stepper value={v.failureThreshold} onChange={(n) => set("failureThreshold", n)} min={1} max={10} />
            <div className="hint">Consecutive failed checks before marked down.</div>
          </div>
        </div>
      </Section>

      {/* 3 — Validation */}
      <Section num={3} title="Validation" desc="What makes a check pass">
        <div className="field">
          <label>Expected status code</label>
          <div className="row wrap">
            <div className="chips">
              {STATUS_PRESETS.map((s) => (
                <button
                  type="button"
                  key={s}
                  className={`chip ${v.expectedStatus === s ? "selected" : ""}`}
                  onClick={() => set("expectedStatus", s)}
                >
                  {s}
                </button>
              ))}
            </div>
            <input
              type="number"
              style={{ width: 110 }}
              value={v.expectedStatus}
              onChange={(e) => set("expectedStatus", Number(e.target.value))}
              aria-label="custom status"
            />
          </div>
          <div className="hint">Any other status fails the check unless an assertion overrides it.</div>
        </div>

        <hr style={{ border: "none", borderTop: "1px solid var(--border)", margin: "0.25rem 0 1rem" }} />

        <AssertionBuilder value={v.assertions} onChange={(a) => set("assertions", a)} />
      </Section>

      {/* 4 — Alerts */}
      <Section
        num={4}
        title="Alerts"
        desc={v.channelIds.length > 0 ? `${v.channelIds.length} selected` : "Optional"}
      >
        <div className="field" style={{ marginBottom: 0 }}>
          <label>Notify these channels</label>
          {channels.length === 0 ? (
            <p className="faint" style={{ fontSize: "0.85rem", margin: "0.25rem 0 0" }}>
              No alert channels yet. <Link href="/channels">Create one</Link> to get notified when this endpoint goes down.
            </p>
          ) : (
            <>
              <div className="chips">
                {channels.map((c) => {
                  const selected = v.channelIds.includes(c.id);
                  const target = c.kind === "email" ? (c.config.to as string) : (c.config.chatId as string);
                  return (
                    <button
                      type="button"
                      key={c.id}
                      className={`chip ${selected ? "selected" : ""}`}
                      onClick={() => toggleChannel(c.id)}
                      title={target}
                    >
                      <span aria-hidden>{c.kind === "email" ? "✉" : "✈"}</span>
                      {c.label}
                      {selected && <span aria-hidden>✓</span>}
                    </button>
                  );
                })}
              </div>
              <div className="hint">
                Alerts fire when this endpoint goes down (and recovers). <Link href="/channels">Manage channels</Link>.
              </div>
            </>
          )}
        </div>
      </Section>

      {error && <p className="error-text">{error}</p>}

      <div className="form-actions">
        <button type="submit" className="primary" disabled={saving}>
          {saving ? "Saving…" : submitLabel}
        </button>
        {onCancel && <button type="button" className="ghost" onClick={onCancel}>Cancel</button>}
      </div>
    </form>
  );
}
