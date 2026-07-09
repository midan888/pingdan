"use client";

import { useCallback, useEffect, useState, type ReactNode } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Nav } from "@/components/Nav";
import { ResponseTimeChart, StatusTimeline } from "@/components/Charts";
import { EndpointForm, EndpointFormValues } from "@/components/EndpointForm";
import { channelIcon } from "@/lib/channels";
import {
  api,
  getToken,
  daysUntil,
  sslSeverity,
  SSL_ALERT_THRESHOLD_DAYS,
  type AlertChannel,
  type Assertion,
  type Check,
  type Endpoint,
  type EndpointDetail,
  type EndpointStats,
} from "@/lib/api";

const SSL_SEVERITY_COLOR: Record<string, string | undefined> = {
  ok: "var(--up)",
  warn: "var(--warn)",
  critical: "var(--down)",
  expired: "var(--down)",
};

function SSLCard({ endpoint, onChecked }: { endpoint: Endpoint; onChecked: (e: Endpoint) => void }) {
  const [checking, setChecking] = useState(false);

  // Only HTTPS endpoints have a certificate to monitor.
  if (!endpoint.url.startsWith("https://")) return null;

  async function checkNow() {
    setChecking(true);
    try {
      const updated = await api<Endpoint>(`/endpoints/${endpoint.id}/ssl-check`, { method: "POST" });
      onChecked(updated);
    } catch {
      /* surfaced via the error field on next load */
    } finally {
      setChecking(false);
    }
  }

  const expires = endpoint.sslExpiresAt;
  const daysLeft = expires ? daysUntil(expires) : null;
  const severity = daysLeft != null ? sslSeverity(daysLeft) : "ok";
  const color = SSL_SEVERITY_COLOR[severity];

  let headline: ReactNode;
  if (endpoint.sslLastError && !expires) {
    headline = <span style={{ color: "var(--down)" }}>Check failed</span>;
  } else if (daysLeft == null) {
    headline = <span className="faint">Not checked yet</span>;
  } else if (daysLeft < 0) {
    headline = <span style={{ color }}>Expired</span>;
  } else {
    headline = (
      <span style={{ color }}>
        {daysLeft} <small>{daysLeft === 1 ? "day left" : "days left"}</small>
      </span>
    );
  }

  return (
    <div className="card stat" style={{ marginBottom: "1rem" }}>
      <div className="spread">
        <div className="label">SSL certificate</div>
        <button className="ghost" onClick={checkNow} disabled={checking}>
          {checking ? "Checking…" : "Check now"}
        </button>
      </div>
      <div className="value" style={{ marginTop: "0.3rem" }}>{headline}</div>
      <p className="faint" style={{ margin: "0.4rem 0 0", fontSize: "0.8rem" }}>
        {expires && `Expires ${new Date(expires).toLocaleString()}. `}
        {daysLeft != null && daysLeft >= 0 && daysLeft <= SSL_ALERT_THRESHOLD_DAYS && (
          <span style={{ color: "var(--down)" }}>Daily renewal reminders are being sent. </span>
        )}
        {endpoint.sslLastError && <span style={{ color: "var(--down)" }}>{endpoint.sslLastError} </span>}
        {endpoint.sslLastCheckedAt && `Last checked ${new Date(endpoint.sslLastCheckedAt).toLocaleString()}.`}
      </p>
    </div>
  );
}

const RANGES = [
  { label: "1h", hours: 1 },
  { label: "6h", hours: 6 },
  { label: "24h", hours: 24 },
  { label: "7d", hours: 168 },
];

function fmtMs(v: number | null | undefined): string {
  return v == null ? "—" : `${Math.round(v)} ms`;
}

export default function EndpointDetailPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const id = params.id;

  const [endpoint, setEndpoint] = useState<Endpoint | null>(null);
  const [assertions, setAssertions] = useState<Assertion[]>([]);
  const [channelIds, setChannelIds] = useState<string[]>([]);
  const [channels, setChannels] = useState<AlertChannel[]>([]);
  const [checks, setChecks] = useState<Check[]>([]);
  const [stats, setStats] = useState<EndpointStats | null>(null);
  const [hours, setHours] = useState(24);
  const [editing, setEditing] = useState(false);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    const [detail, ck, st] = await Promise.all([
      api<EndpointDetail>(`/endpoints/${id}`),
      api<Check[]>(`/endpoints/${id}/checks?hours=${hours}&limit=2000`),
      api<EndpointStats>(`/endpoints/${id}/stats?hours=${hours}`),
    ]);
    setEndpoint(detail.endpoint);
    setAssertions(detail.assertions ?? []);
    setChannelIds(detail.channelIds ?? []);
    setChecks(ck);
    setStats(st);
    setLoading(false);
  }, [id, hours]);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    load().catch(() => setLoading(false));
    api<AlertChannel[]>("/alert-channels").then(setChannels).catch(() => {});
  }, [load, router]);

  // refresh checks/stats periodically for a live feel
  useEffect(() => {
    const t = setInterval(() => {
      if (!editing) load().catch(() => {});
    }, 15000);
    return () => clearInterval(t);
  }, [load, editing]);

  async function saveEdit(v: EndpointFormValues) {
    await api(`/endpoints/${id}`, { method: "PUT", body: JSON.stringify(v) });
    setEditing(false);
    await load();
  }

  async function remove() {
    if (!confirm("Delete this endpoint and all its history?")) return;
    await api(`/endpoints/${id}`, { method: "DELETE" });
    router.push("/endpoints");
  }

  const failedChecks = stats ? stats.total - stats.upCount : 0;

  if (loading) {
    return (
      <>
        <Nav />
        <div className="container"><p className="muted">Loading…</p></div>
      </>
    );
  }

  if (!endpoint) {
    return (
      <>
        <Nav />
        <div className="container"><div className="empty">Endpoint not found.</div></div>
      </>
    );
  }

  return (
    <>
      <Nav />
      <div className="container">
        <div style={{ marginBottom: "0.75rem" }}>
          <Link href="/endpoints" className="muted">← Endpoints</Link>
        </div>

        <div className="page-head">
          <div>
            <div className="row">
              <span className={`dot ${endpoint.currentState}`} />
              <h1 style={{ margin: 0 }}>{endpoint.name}</h1>
              <span className={`pill ${endpoint.currentState}`}>{endpoint.currentState}</span>
            </div>
            <a className="subtitle mono" href={endpoint.url} target="_blank" rel="noreferrer">
              {endpoint.method} {endpoint.url}
            </a>
          </div>
          <div className="row">
            <button onClick={() => setEditing((e) => !e)}>{editing ? "Close" : "Edit"}</button>
            <button className="danger" onClick={remove}>Delete</button>
          </div>
        </div>

        {editing ? (
          <EndpointForm
            initial={{
              name: endpoint.name,
              url: endpoint.url,
              method: endpoint.method,
              expectedStatus: endpoint.expectedStatus,
              intervalSec: endpoint.intervalSec,
              timeoutSec: endpoint.timeoutSec,
              failureThreshold: endpoint.failureThreshold,
              groupId: endpoint.groupId,
              assertions,
              channelIds,
            }}
            submitLabel="Save changes"
            onSubmit={saveEdit}
            onCancel={() => setEditing(false)}
          />
        ) : (
          <>
            {/* stat cards */}
            <div className="spread" style={{ marginBottom: "0.75rem" }}>
              <h2 style={{ margin: 0 }}>Overview</h2>
              <div className="segmented">
                {RANGES.map((r) => (
                  <button key={r.hours} className={hours === r.hours ? "active" : ""} onClick={() => setHours(r.hours)}>
                    {r.label}
                  </button>
                ))}
              </div>
            </div>

            <div className="grid grid-4" style={{ marginBottom: "1rem" }}>
              <div className="card stat">
                <div className="label">Uptime</div>
                <div className="value" style={{ color: stats && stats.uptimePct >= 99 ? "var(--up)" : stats && stats.uptimePct < 95 ? "var(--down)" : undefined }}>
                  {stats ? stats.uptimePct.toFixed(2) : "—"}<small>%</small>
                </div>
              </div>
              <div className="card stat">
                <div className="label">Avg response</div>
                <div className="value">{fmtMs(stats?.avgLatencyMs)}</div>
              </div>
              <div className="card stat">
                <div className="label">Checks</div>
                <div className="value">{stats?.total ?? 0}</div>
              </div>
              <div className="card stat">
                <div className="label">Failed checks</div>
                <div className="value" style={{ color: failedChecks > 0 ? "var(--down)" : undefined }}>{failedChecks}</div>
              </div>
            </div>

            {/* SSL certificate countdown */}
            <SSLCard endpoint={endpoint} onChecked={setEndpoint} />

            {/* attached alert channels */}
            <div className="card" style={{ marginBottom: "1rem" }}>
              <div className="spread">
                <h3 style={{ margin: 0 }}>Alerts</h3>
                <button className="ghost" onClick={() => setEditing(true)}>Edit</button>
              </div>
              {channelIds.length === 0 ? (
                <p className="faint" style={{ margin: "0.5rem 0 0", fontSize: "0.85rem" }}>
                  No channels attached — this endpoint won&apos;t notify anyone when it goes down.
                </p>
              ) : (
                <div className="chips" style={{ marginTop: "0.6rem" }}>
                  {channelIds.map((cid) => {
                    const c = channels.find((x) => x.id === cid);
                    return (
                      <span className="chip selected" key={cid} style={{ cursor: "default" }}>
                        <span aria-hidden>{c ? channelIcon(c.kind) : "•"}</span>
                        {c ? c.label : "Unknown channel"}
                      </span>
                    );
                  })}
                </div>
              )}
            </div>

            {/* response time chart */}
            <div className="card" style={{ marginBottom: "1rem" }}>
              <div className="spread" style={{ marginBottom: "0.75rem" }}>
                <h3 style={{ margin: 0 }}>Response time</h3>
                <span className="faint" style={{ fontSize: "0.78rem" }}>
                  p50 {fmtMs(stats?.p50LatencyMs)} · p95 {fmtMs(stats?.p95LatencyMs)} · min {fmtMs(stats?.minLatencyMs)} · max {fmtMs(stats?.maxLatencyMs)}
                </span>
              </div>
              <ResponseTimeChart checks={checks} />
            </div>

            {/* status timeline */}
            <div className="card" style={{ marginBottom: "1rem" }}>
              <h3>Status history</h3>
              <StatusTimeline checks={checks} />
            </div>

            {/* recent checks table */}
            <div className="card">
              <h3>Recent checks</h3>
              <div className="table-scroll">
              <table>
                <thead>
                  <tr>
                    <th>Time</th>
                    <th>Result</th>
                    <th>Status</th>
                    <th className="num">Latency</th>
                    <th>Detail</th>
                  </tr>
                </thead>
                <tbody>
                  {checks.slice(0, 50).map((c) => (
                    <tr key={c.id}>
                      <td className="mono">{new Date(c.checkedAt).toLocaleString()}</td>
                      <td><span className={`pill ${c.ok ? "up" : "down"}`}>{c.ok ? "pass" : "fail"}</span></td>
                      <td className="mono">{c.statusCode ?? "—"}</td>
                      <td className="num mono">{c.latencyMs != null ? `${c.latencyMs} ms` : "—"}</td>
                      <td className="faint" style={{ fontSize: "0.8rem" }}>
                        {c.failedAssertions && c.failedAssertions.length > 0
                          ? c.failedAssertions.map((f, i) => (
                              <div key={i} className="mono" style={{ color: "var(--down)" }}>
                                {f.source}{f.property ? `(${f.property})` : ""} {f.comparison} {f.target} → got {f.actual || "∅"}
                              </div>
                            ))
                          : c.error ?? ""}
                      </td>
                    </tr>
                  ))}
                  {checks.length === 0 && (
                    <tr><td colSpan={5} className="faint">No checks yet — first check runs within the interval.</td></tr>
                  )}
                </tbody>
              </table>
              </div>
            </div>
          </>
        )}
      </div>
    </>
  );
}
