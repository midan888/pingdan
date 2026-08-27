"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Nav } from "@/components/Nav";
import { Sparkline, MiniStatusBar } from "@/components/Charts";
import { api, displayState, getToken, groupStatusColor, type Check, type Endpoint, type EndpointHistory, type Group } from "@/lib/api";

type Row = { endpoint: Endpoint; checks: Check[]; history: EndpointHistory | null };

// Window the dashboard cards summarise. The status bar and the uptime figure
// on a card both cover this period.
const WINDOW_HOURS = 24;
const BUCKETS = 30;

// Sentinel filter values that aren't real group ids.
const ALL = "__all__";
const UNGROUPED = "__ungrouped__";

function EndpointCard({ endpoint, checks, history }: Row) {
  const last = checks[0];
  const state = displayState(endpoint);
  // A recovered monitor is legitimately "up", so the incident has to read from
  // the window figures instead of the live state.
  const uptimeColor =
    history == null || history.total === 0
      ? "var(--text-dim)"
      : history.uptimePct >= 99.9
      ? "var(--text-dim)"
      : history.uptimePct >= 99
      ? "var(--warn)"
      : "var(--down)";
  return (
    <Link href={`/endpoints/${endpoint.id}`} style={{ color: "inherit", textDecoration: "none" }}>
      <div className="card hoverable">
        <div className="spread" style={{ marginBottom: "0.75rem" }}>
          <div className="row">
            <span className={`dot ${state}`} />
            <strong>{endpoint.name}</strong>
          </div>
          <span
            className={`pill ${state}`}
            title={
              state === "degraded"
                ? `${endpoint.consecutiveFailures} failed check${endpoint.consecutiveFailures === 1 ? "" : "s"} in a row — alerts at ${endpoint.failureThreshold}`
                : undefined
            }
          >
            {state}
          </span>
        </div>

        <div className="mono muted" style={{ fontSize: "0.78rem", marginBottom: "0.75rem", whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis" }}>
          {endpoint.url}
        </div>

        <Sparkline checks={checks} width={400} height={40} />
        <div style={{ marginTop: "0.5rem" }}>
          <MiniStatusBar buckets={history?.buckets ?? []} windowLabel={`last ${WINDOW_HOURS}h`} />
        </div>

        <div className="spread" style={{ marginTop: "0.85rem", fontSize: "0.82rem" }}>
          <span style={{ color: uptimeColor }}>
            {history ? `${history.uptimePct.toFixed(1)}% uptime` : "—"}
            {history && history.failed > 0 && (
              <span className="faint"> · {history.failed} failed</span>
            )}
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
        const [checks, history] = await Promise.all([
          api<Check[]>(`/endpoints/${endpoint.id}/checks?limit=40`).catch(() => [] as Check[]),
          api<EndpointHistory>(
            `/endpoints/${endpoint.id}/history?hours=${WINDOW_HOURS}&buckets=${BUCKETS}`
          ).catch(() => null),
        ]);
        return { endpoint, checks, history };
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

  const states = rows.map((r) => displayState(r.endpoint));
  const up = states.filter((s) => s === "up").length;
  const down = states.filter((s) => s === "down").length;
  const degraded = states.filter((s) => s === "degraded").length;
  const unknown = states.filter((s) => s === "unknown").length;
  const avgUptime =
    rows.length > 0
      ? rows.reduce((acc, r) => acc + (r.history?.uptimePct ?? 0), 0) / rows.length
      : null;
  const failedChecks = rows.reduce((acc, r) => acc + (r.history?.failed ?? 0), 0);

  return (
    <>
      <Nav />
      <div className="container">
        <div className="page-head">
          <div>
            <h1>Dashboard</h1>
            <div className="subtitle">Live status of all monitors</div>
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
            <Link href="/endpoints/new" className="button-link primary">+ New monitor</Link>
          </div>
        </div>

        {/* summary cards */}
        <div className="grid grid-6 stat-strip" style={{ marginBottom: "1.5rem" }}>
          <div className="card stat">
            <div className="label">Operational</div>
            <div className="value" style={{ color: "var(--up)" }}>{up}</div>
          </div>
          <div className="card stat">
            <div className="label">Down</div>
            <div className="value" style={{ color: down > 0 ? "var(--down)" : undefined }}>{down}</div>
          </div>
          <div className="card stat" title="Latest check failed, but not enough failures in a row to alert yet">
            <div className="label">Degraded</div>
            <div className="value" style={{ color: degraded > 0 ? "var(--warn)" : undefined }}>{degraded}</div>
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
            <p>No monitors yet.</p>
            <Link href="/endpoints/new" className="button-link primary">Create your first monitor</Link>
          </div>
        ) : (
          sections.map((section) => {
            const sectionStates = section.rows.map((r) => displayState(r.endpoint));
            const sUp = sectionStates.filter((s) => s === "up").length;
            const sDown = sectionStates.filter((s) => s === "down").length;
            const sDegraded = sectionStates.filter((s) => s === "degraded").length;
            // Monitors that are healthy now but failed at some point in the
            // window. Without this a section reads "all up" the moment an
            // outage ends, which is what a green dot alone can't say.
            const sRecovered = section.rows.filter(
              (r) => displayState(r.endpoint) === "up" && (r.history?.failed ?? 0) > 0
            ).length;
            const accent = groupStatusColor(sectionStates);
            return (
              <section key={section.id} className="group-section" style={{ ["--group-accent" as string]: accent }}>
                <div className="group-header">
                  <h2>{section.name}</h2>
                  <span className="group-count">{section.rows.length}</span>
                  {sDown > 0 && <span className="pill down">{sDown} down</span>}
                  {sDegraded > 0 && <span className="pill degraded">{sDegraded} degraded</span>}
                  {sDown === 0 && sDegraded === 0 && sRecovered > 0 && (
                    <span className="pill degraded" title={`Up now, but failed checks in the last ${WINDOW_HOURS}h`}>
                      {sRecovered} recovered
                    </span>
                  )}
                  {sDown === 0 && sRecovered === 0 && sUp === section.rows.length && (
                    <span className="pill up">all up</span>
                  )}
                </div>
                <div className="grid grid-auto">
                  {section.rows.map(({ endpoint, checks, history }) => (
                    <EndpointCard key={endpoint.id} endpoint={endpoint} checks={checks} history={history} />
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
