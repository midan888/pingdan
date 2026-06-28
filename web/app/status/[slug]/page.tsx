"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import {
  publicApi,
  NotFoundError,
  type PublicStatusPage,
  type PublicStatusItem,
  type PublicTick,
} from "@/lib/api";

const UP = "#3fb950";
const DOWN = "#f85149";
const DIM = "#2c3542";

const OVERALL_COPY: Record<PublicStatusPage["overall"], { label: string; color: string }> = {
  up: { label: "All systems operational", color: UP },
  degraded: { label: "Partial service disruption", color: "#d29922" },
  down: { label: "Major outage", color: DOWN },
  unknown: { label: "Status unavailable", color: DIM },
};

function fmtTime(iso: string): string {
  return new Date(iso).toLocaleString([], { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
}

/** A compact uptime strip built from public ticks (oldest → newest). */
function Timeline({ history }: { history: PublicTick[] }) {
  const [tip, setTip] = useState<{ x: number; y: number; lines: string[] } | null>(null);
  if (history.length === 0) {
    return <div style={{ color: "#7d8590", fontSize: "0.8rem" }}>No data yet</div>;
  }
  return (
    <div style={{ position: "relative" }}>
      <div style={{ display: "flex", gap: 2, height: 28 }}>
        {history.map((t, i) => (
          <div
            key={i}
            style={{ flex: 1, minWidth: 0, borderRadius: 2, background: t.ok ? UP : DOWN, opacity: 0.85, cursor: "pointer" }}
            onMouseEnter={(e) =>
              setTip({ x: e.clientX, y: e.clientY, lines: [fmtTime(t.checkedAt), t.ok ? "Operational" : "Down"] })
            }
            onMouseMove={(e) => setTip((p) => (p ? { ...p, x: e.clientX, y: e.clientY } : p))}
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

function ServiceRow({ item }: { item: PublicStatusItem }) {
  const color = item.state === "up" ? UP : item.state === "down" ? DOWN : DIM;
  const stateLabel = item.state === "up" ? "Operational" : item.state === "down" ? "Down" : "Pending";
  return (
    <div className="card" style={{ marginBottom: "0.75rem" }}>
      <div className="spread" style={{ marginBottom: "0.6rem" }}>
        <div className="row">
          <span className={`dot ${item.state}`} />
          <strong>{item.name}</strong>
        </div>
        <span style={{ color, fontSize: "0.85rem", fontWeight: 600 }}>{stateLabel}</span>
      </div>
      <Timeline history={item.history} />
      <div className="spread" style={{ marginTop: "0.6rem", fontSize: "0.8rem" }}>
        <span className="muted">{item.uptimePct.toFixed(2)}% uptime (90d)</span>
      </div>
    </div>
  );
}

export default function PublicStatusPage() {
  const params = useParams();
  const slug = params.slug as string;
  const [data, setData] = useState<PublicStatusPage | null>(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  async function load() {
    try {
      const d = await publicApi<PublicStatusPage>(`/public/status/${slug}`);
      setData(d);
    } catch (e) {
      if (e instanceof NotFoundError) setNotFound(true);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
    const t = setInterval(() => load().catch(() => {}), 30000);
    return () => clearInterval(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [slug]);

  if (notFound) {
    return (
      <div className="container" style={{ maxWidth: 720, margin: "0 auto", paddingTop: "4rem", textAlign: "center" }}>
        <h1>Status page not found</h1>
        <p className="muted">This status page does not exist or has been removed.</p>
      </div>
    );
  }

  if (loading || !data) {
    return (
      <div className="container" style={{ maxWidth: 720, margin: "0 auto", paddingTop: "4rem" }}>
        <p className="muted">Loading…</p>
      </div>
    );
  }

  const overall = OVERALL_COPY[data.overall];

  return (
    <div className="container" style={{ maxWidth: 720, margin: "0 auto", paddingTop: "2.5rem", paddingBottom: "3rem" }}>
      <div style={{ textAlign: "center", marginBottom: "2rem" }}>
        <h1 style={{ marginBottom: "0.25rem" }}>{data.title}</h1>
        {data.description && <p className="muted" style={{ marginTop: 0 }}>{data.description}</p>}
      </div>

      <div
        className="card"
        style={{
          textAlign: "center",
          borderLeft: `4px solid ${overall.color}`,
          marginBottom: "1.5rem",
        }}
      >
        <div style={{ fontSize: "1.1rem", fontWeight: 600, color: overall.color }}>{overall.label}</div>
        <div className="faint" style={{ fontSize: "0.78rem", marginTop: "0.35rem" }}>
          Updated {fmtTime(data.updatedAt)}
        </div>
      </div>

      {data.items.length === 0 ? (
        <div className="empty"><p>No services are listed on this page yet.</p></div>
      ) : (
        data.items.map((item, i) => <ServiceRow key={i} item={item} />)
      )}

      <div style={{ textAlign: "center", marginTop: "2rem" }}>
        <span className="faint" style={{ fontSize: "0.75rem" }}>
          Powered by ping<span style={{ color: "var(--up)" }}>·</span>dan
        </span>
      </div>
    </div>
  );
}
