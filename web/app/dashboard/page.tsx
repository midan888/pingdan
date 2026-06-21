"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Nav } from "@/components/Nav";
import { Sparkline, MiniStatusBar } from "@/components/Charts";
import { api, getToken, type Check, type Endpoint, type EndpointStats } from "@/lib/api";

type Row = { endpoint: Endpoint; checks: Check[]; stats: EndpointStats | null };

export default function DashboardPage() {
  const router = useRouter();
  const [rows, setRows] = useState<Row[]>([]);
  const [loading, setLoading] = useState(true);

  async function load() {
    const eps = await api<Endpoint[]>("/endpoints");
    const rows = await Promise.all(
      eps.map(async (endpoint) => {
        const [checks, stats] = await Promise.all([
          api<Check[]>(`/endpoints/${endpoint.id}/checks?limit=40`).catch(() => [] as Check[]),
          api<EndpointStats>(`/endpoints/${endpoint.id}/stats?hours=24`).catch(() => null),
        ]);
        return { endpoint, checks, stats };
      })
    );
    setRows(rows);
    setLoading(false);
  }

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    load().catch(() => setLoading(false));
    const t = setInterval(() => load().catch(() => {}), 15000);
    return () => clearInterval(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [router]);

  const eps = rows.map((r) => r.endpoint);
  const up = eps.filter((e) => e.currentState === "up").length;
  const down = eps.filter((e) => e.currentState === "down").length;
  const unknown = eps.filter((e) => e.currentState === "unknown").length;
  const avgUptime =
    rows.length > 0
      ? rows.reduce((acc, r) => acc + (r.stats?.uptimePct ?? 0), 0) / rows.length
      : null;

  return (
    <>
      <Nav />
      <div className="container">
        <div className="page-head">
          <div>
            <h1>Dashboard</h1>
            <div className="subtitle">Live status of all monitored endpoints</div>
          </div>
          <Link href="/endpoints/new"><button className="primary">+ New endpoint</button></Link>
        </div>

        {/* summary cards */}
        <div className="grid grid-4" style={{ marginBottom: "1.5rem" }}>
          <div className="card stat">
            <div className="label">Operational</div>
            <div className="value" style={{ color: "var(--up)" }}>{up}</div>
          </div>
          <div className="card stat">
            <div className="label">Down</div>
            <div className="value" style={{ color: down > 0 ? "var(--down)" : undefined }}>{down}</div>
          </div>
          <div className="card stat">
            <div className="label">Pending</div>
            <div className="value">{unknown}</div>
          </div>
          <div className="card stat">
            <div className="label">Avg uptime (24h)</div>
            <div className="value">{avgUptime != null ? avgUptime.toFixed(2) : "—"}<small>%</small></div>
          </div>
        </div>

        {loading ? (
          <p className="muted">Loading…</p>
        ) : rows.length === 0 ? (
          <div className="empty">
            <p>No endpoints to monitor yet.</p>
            <Link href="/endpoints/new"><button className="primary">Create your first endpoint</button></Link>
          </div>
        ) : (
          <div className="grid grid-auto">
            {rows.map(({ endpoint, checks, stats }) => {
              const last = checks[0];
              return (
                <Link key={endpoint.id} href={`/endpoints/${endpoint.id}`} style={{ color: "inherit", textDecoration: "none" }}>
                  <div className="card hoverable">
                    <div className="spread" style={{ marginBottom: "0.75rem" }}>
                      <div className="row">
                        <span className={`dot ${endpoint.currentState}`} />
                        <strong>{endpoint.name}</strong>
                      </div>
                      <span className={`pill ${endpoint.currentState}`}>{endpoint.currentState}</span>
                    </div>

                    <div className="mono muted" style={{ fontSize: "0.78rem", marginBottom: "0.75rem", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
                      {endpoint.url}
                    </div>

                    <Sparkline checks={checks} width={400} height={40} />
                    <div style={{ marginTop: "0.5rem" }}>
                      <MiniStatusBar checks={checks} count={30} />
                    </div>

                    <div className="spread" style={{ marginTop: "0.85rem", fontSize: "0.82rem" }}>
                      <span className="muted">
                        {stats ? `${stats.uptimePct.toFixed(1)}% uptime` : "—"}
                      </span>
                      <span className="mono muted">
                        {last?.latencyMs != null ? `${last.latencyMs} ms` : "—"}
                      </span>
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
        )}
      </div>
    </>
  );
}
