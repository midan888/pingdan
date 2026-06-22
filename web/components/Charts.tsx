"use client";

import { useState } from "react";
import type { Check } from "@/lib/api";

const UP = "#3fb950";
const DOWN = "#f85149";
const DIM = "#2c3542";

type Tip = { x: number; y: number; lines: string[] } | null;

function fmtTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

/**
 * downsample reduces an array to at most `max` evenly-spaced items, preserving
 * the first and last. Keeps bars/segments from overflowing the container when a
 * wide time window returns more checks than there are pixels to draw them.
 */
function downsample<T>(arr: T[], max: number): T[] {
  if (arr.length <= max) return arr;
  const step = arr.length / max;
  const out: T[] = [];
  for (let i = 0; i < max; i++) out.push(arr[Math.floor(i * step)]);
  return out;
}

// Upper bound on how many bars/segments we draw; beyond this they'd be
// sub-pixel and the flex row overflows its container.
const MAX_BARS = 240;

/**
 * ResponseTimeChart renders one bar per check (oldest → newest, left → right).
 * Bar height = latency, colour = pass/fail. Hover shows a tooltip.
 */
export function ResponseTimeChart({ checks, height = 140 }: { checks: Check[]; height?: number }) {
  const [tip, setTip] = useState<Tip>(null);
  // checks come newest-first; show oldest on the left, capped so bars fit
  const data = downsample([...checks].reverse(), MAX_BARS);
  if (data.length === 0) {
    return <div className="empty" style={{ padding: "2rem" }}>No checks recorded yet.</div>;
  }
  const maxLat = Math.max(1, ...data.map((c) => c.latencyMs ?? 0));

  return (
    <div className="chart-wrap">
      <div className="chart-bars" style={{ height }}>
        {data.map((c) => {
          const lat = c.latencyMs ?? 0;
          // a failed check with no latency still shows a small marker
          const pct = c.latencyMs == null ? 100 : (lat / maxLat) * 100;
          const color = c.ok ? UP : DOWN;
          return (
            <div
              key={c.id}
              style={{
                flex: 1,
                minWidth: 2,
                height: `${Math.max(pct, 3)}%`,
                background: color,
                borderRadius: "2px 2px 0 0",
                opacity: 0.85,
                cursor: "pointer",
              }}
              onMouseEnter={(e) =>
                setTip({
                  x: e.clientX,
                  y: e.clientY,
                  lines: [
                    fmtTime(c.checkedAt),
                    c.latencyMs != null ? `${c.latencyMs} ms` : "no response",
                    c.statusCode != null ? `HTTP ${c.statusCode}` : (c.error ?? "error"),
                    c.ok ? "✓ passed" : "✗ failed",
                  ],
                })
              }
              onMouseMove={(e) => setTip((t) => (t ? { ...t, x: e.clientX, y: e.clientY } : t))}
              onMouseLeave={() => setTip(null)}
            />
          );
        })}
      </div>
      <div className="chart-legend">
        <span><span className="sw" style={{ background: UP }} />Passing</span>
        <span><span className="sw" style={{ background: DOWN }} />Failing</span>
        <span style={{ marginLeft: "auto" }} className="faint">peak {maxLat} ms</span>
      </div>
      {tip && (
        <div className="bar-tip" style={{ left: tip.x + 12, top: tip.y + 12 }}>
          {tip.lines.map((l, i) => (
            <div key={i} style={{ color: i === 0 ? "var(--text-dim)" : "var(--text)" }}>{l}</div>
          ))}
        </div>
      )}
    </div>
  );
}

/**
 * StatusTimeline renders a compact uptime strip (UptimeRobot style): one segment
 * per check, green = up, red = down. Oldest → newest, left → right.
 */
export function StatusTimeline({ checks }: { checks: Check[] }) {
  const [tip, setTip] = useState<Tip>(null);
  const data = downsample([...checks].reverse(), MAX_BARS);
  if (data.length === 0) {
    return <div className="faint" style={{ fontSize: "0.8rem" }}>—</div>;
  }
  return (
    <div className="chart-wrap">
      <div style={{ display: "flex", gap: 2, height: 28 }}>
        {data.map((c) => (
          <div
            key={c.id}
            style={{
              flex: 1,
              minWidth: 2,
              borderRadius: 2,
              background: c.ok ? UP : DOWN,
              opacity: 0.85,
              cursor: "pointer",
            }}
            onMouseEnter={(e) =>
              setTip({
                x: e.clientX,
                y: e.clientY,
                lines: [
                  fmtTime(c.checkedAt),
                  c.ok ? "Operational" : (c.error ?? "Down"),
                  c.statusCode != null ? `HTTP ${c.statusCode}` : "",
                ].filter(Boolean),
              })
            }
            onMouseMove={(e) => setTip((t) => (t ? { ...t, x: e.clientX, y: e.clientY } : t))}
            onMouseLeave={() => setTip(null)}
          />
        ))}
      </div>
      {tip && (
        <div className="bar-tip" style={{ left: tip.x + 12, top: tip.y + 12 }}>
          {tip.lines.map((l, i) => (
            <div key={i} style={{ color: i === 0 ? "var(--text-dim)" : "var(--text)" }}>{l}</div>
          ))}
        </div>
      )}
    </div>
  );
}

/**
 * Sparkline draws a tiny inline SVG line of recent latencies for dashboard cards.
 * It scales to its container width via viewBox, so width/height only define the
 * internal coordinate space (aspect ratio), never the rendered pixel size.
 */
export function Sparkline({ checks, width = 180, height = 36 }: { checks: Check[]; width?: number; height?: number }) {
  const data = [...checks].reverse().filter((c) => c.latencyMs != null);
  const svgProps = {
    viewBox: `0 0 ${width} ${height}`,
    preserveAspectRatio: "none" as const,
    style: { display: "block", width: "100%", height },
  };
  if (data.length < 2) {
    return <svg {...svgProps} />;
  }
  const lats = data.map((c) => c.latencyMs as number);
  const max = Math.max(...lats);
  const min = Math.min(...lats);
  const span = Math.max(1, max - min);
  const step = width / (data.length - 1);
  const pts = lats.map((l, i) => {
    const x = i * step;
    const y = height - ((l - min) / span) * (height - 4) - 2;
    return `${x.toFixed(1)},${y.toFixed(1)}`;
  });
  const anyDown = data.some((c) => !c.ok);
  const stroke = anyDown ? DOWN : UP;
  return (
    <svg {...svgProps}>
      <polyline
        points={pts.join(" ")}
        fill="none"
        stroke={stroke}
        strokeWidth={1.5}
        strokeLinejoin="round"
        strokeLinecap="round"
        vectorEffect="non-scaling-stroke"
      />
    </svg>
  );
}

/** UptimeBars: small fixed-count status bars used in dashboard cards. */
export function MiniStatusBar({ checks, count = 30 }: { checks: Check[]; count?: number }) {
  const data = [...checks].reverse().slice(-count);
  return (
    <div style={{ display: "flex", gap: 2, height: 18 }}>
      {Array.from({ length: count }).map((_, i) => {
        const c = data[i - (count - data.length)];
        const bg = c == null ? DIM : c.ok ? UP : DOWN;
        return <div key={i} style={{ flex: 1, borderRadius: 1.5, background: bg, opacity: c == null ? 0.5 : 0.85 }} />;
      })}
    </div>
  );
}
