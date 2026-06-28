"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Nav } from "@/components/Nav";
import { Sparkline, MiniStatusBar } from "@/components/Charts";
import { api, getToken, groupStatusColor, type Check, type Endpoint, type EndpointStats, type Group } from "@/lib/api";

type Row = { endpoint: Endpoint; checks: Check[]; stats: EndpointStats | null };

// Sentinel filter values that aren't real group ids.
const ALL = "__all__";
const UNGROUPED = "__ungrouped__";

function EndpointCard({ endpoint, checks, stats }: Row) {
  const last = checks[0];
  return (
    <Link href={`/endpoints/${endpoint.id}`} style={{ color: "inherit", textDecoration: "none" }}>
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
}

export default function DashboardPage() {
  const router = useRouter();
  const [allRows, setAllRows] = useState<Row[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [filter, setFilter] = useState<string>(ALL);
  const [loading, setLoading] = useState(true);

  async function load() {
    const [eps, grps] = await Promise.all([
      api<Endpoint[]>("/endpoints"),
      api<Group[]>("/groups").catch(() => [] as Group[]),
    ]);
    setGroups(grps);
    const rows = await Promise.all(
      eps.map(async (endpoint) => {
        const [checks, stats] = await Promise.all([
          api<Check[]>(`/endpoints/${endpoint.id}/checks?limit=40`).catch(() => [] as Check[]),
          api<EndpointStats>(`/endpoints/${endpoint.id}/stats?hours=24`).catch(() => null),
        ]);
        return { endpoint, checks, stats };
      })
    );
    setAllRows(rows);
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

  const rows =
    filter === ALL
      ? allRows
      : filter === UNGROUPED
      ? allRows.filter((r) => !r.endpoint.groupId)
      : allRows.filter((r) => r.endpoint.groupId === filter);

  const groupName = (id: string | null) => groups.find((g) => g.id === id)?.name ?? null;
  const hasUngrouped = allRows.some((r) => !r.endpoint.groupId);

  // Bucket the visible rows into sections: one per group (in the group list's
  // order), then an "Ungrouped" section last. Empty sections are dropped.
  const sections: { id: string; name: string; rows: Row[] }[] = [];
  for (const g of groups) {
    const grouped = rows.filter((r) => r.endpoint.groupId === g.id);
    if (grouped.length > 0) sections.push({ id: g.id, name: g.name, rows: grouped });
  }
  const ungrouped = rows.filter((r) => !r.endpoint.groupId || !groupName(r.endpoint.groupId));
  if (ungrouped.length > 0) sections.push({ id: UNGROUPED, name: "Ungrouped", rows: ungrouped });

  const eps = rows.map((r) => r.endpoint);
  const up = eps.filter((e) => e.currentState === "up").length;
  const down = eps.filter((e) => e.currentState === "down").length;
  const unknown = eps.filter((e) => e.currentState === "unknown").length;
  const avgUptime =
    rows.length > 0
      ? rows.reduce((acc, r) => acc + (r.stats?.uptimePct ?? 0), 0) / rows.length
      : null;
  const failedChecks = rows.reduce(
    (acc, r) => acc + (r.stats ? r.stats.total - r.stats.upCount : 0),
    0
  );

  return (
    <>
      <Nav />
      <div className="container">
        <div className="page-head">
          <div>
            <h1>Dashboard</h1>
            <div className="subtitle">Live status of all monitored endpoints</div>
          </div>
          <div className="row" style={{ gap: "0.6rem" }}>
            {groups.length > 0 && (
              <select
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
                aria-label="Filter by group"
              >
                <option value={ALL}>All groups</option>
                {groups.map((g) => (
                  <option key={g.id} value={g.id}>{g.name}</option>
                ))}
                {hasUngrouped && <option value={UNGROUPED}>Ungrouped</option>}
              </select>
            )}
            <Link href="/endpoints/new"><button className="primary">+ New endpoint</button></Link>
          </div>
        </div>

        {/* summary cards */}
        <div className="grid grid-5" style={{ marginBottom: "1.5rem" }}>
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
            <div className="label">Failed checks (24h)</div>
            <div className="value" style={{ color: failedChecks > 0 ? "var(--down)" : undefined }}>{failedChecks}</div>
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
          sections.map((section) => {
            const states = section.rows.map((r) => r.endpoint.currentState);
            const sUp = states.filter((s) => s === "up").length;
            const sDown = states.filter((s) => s === "down").length;
            const accent = groupStatusColor(states);
            return (
              <section key={section.id} className="group-section" style={{ ["--group-accent" as string]: accent }}>
                <div className="group-header">
                  <h2>{section.name}</h2>
                  <span className="group-count">{section.rows.length}</span>
                  {sDown > 0 && <span className="pill down">{sDown} down</span>}
                  {sDown === 0 && sUp === section.rows.length && <span className="pill up">all up</span>}
                </div>
                <div className="grid grid-auto">
                  {section.rows.map(({ endpoint, checks, stats }) => (
                    <EndpointCard key={endpoint.id} endpoint={endpoint} checks={checks} stats={stats} />
                  ))}
                </div>
              </section>
            );
          })
        )}
      </div>
    </>
  );
}
